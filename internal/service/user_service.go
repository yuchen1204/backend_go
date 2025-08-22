package service

import (
	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	verificationCodeLength = 6
	verificationCodeTTL    = 5 * time.Minute
	// activationGracePeriod 注册后允许未激活登录的宽限时间
	activationGracePeriod  = 24 * time.Hour
)

// UserService 用户服务接口
type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, req *model.UserRegisterRequest) (*model.UserResponse, error)
	// Login 用户登录
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	// RefreshToken 刷新访问Token
	RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error)
	// Logout 用户登出
	Logout(ctx context.Context, req *model.LogoutRequest) error
	// UpdateProfile 更新用户基本信息
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *model.UpdateProfileRequest) (*model.UserResponse, error)
	// SendVerificationCode 发送注册验证码
	SendVerificationCode(ctx context.Context, req *model.SendCodeRequest, ip string) error
	// SendResetPasswordCode 发送重置密码验证码
	SendResetPasswordCode(ctx context.Context, req *model.SendResetCodeRequest, ip string) error
	// ResetPassword 重置密码
	ResetPassword(ctx context.Context, req *model.ResetPasswordRequest) error
	// GetByID 根据ID获取用户
	GetByID(id uuid.UUID) (*model.UserResponse, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(username string) (*model.UserResponse, error)
	// ValidatePassword 验证密码
	ValidatePassword(username, password string) (*model.User, error)
	// 管理员专用方法
	// GetUsersForAdmin 获取用户列表（管理员用）
	GetUsersForAdmin(page, limit int, search string) ([]*model.UserResponse, int64, error)
	// GetUserByID 根据ID获取用户（管理员用）
	GetUserByID(id uint) (*model.UserResponse, error)
	// UpdateUserStatus 更新用户状态
	UpdateUserStatus(id uint, status string) error
	// DeleteUser 删除用户
	DeleteUser(id uint) error
	// GetUserStats 获取用户统计信息
	GetUserStats() (map[string]interface{}, error)
	// UpdateUserStatusByUUID 根据UUID更新用户状态
	UpdateUserStatusByUUID(id uuid.UUID, status string) error
	// DeleteUserByUUID 根据UUID删除用户
	DeleteUserByUUID(id uuid.UUID) error
	// SendActivationCode 发送账户激活验证码
	SendActivationCode(ctx context.Context, req *model.SendActivationCodeRequest, ip string) error
	// ActivateAccount 激活账户
	ActivateAccount(ctx context.Context, req *model.ActivateAccountRequest) error
	// AdminUpdateUserPassword 管理员更新指定用户密码
	AdminUpdateUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error
}

// firstNonEmpty 返回第一个非空字符串
func firstNonEmpty(values ...string) string {
    for _, v := range values {
        if strings.TrimSpace(v) != "" {
            return v
        }
    }
    return ""
}

// RefreshToken handles refresh token requests
func (s *userService) RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error) {
    // 1. 验证Refresh Token格式和签名
    claims, err := s.jwtSvc.ValidateToken(req.RefreshToken)
    if err != nil {
        return nil, errors.New("无效的refresh token")
    }

    // 2. 确保这是一个Refresh Token
    if claims.TokenType != RefreshToken {
        return nil, errors.New("提供的token不是refresh token")
    }

    // 3. 验证Refresh Token是否存在于Redis中
    isValid, err := s.refreshTokenRepo.Validate(ctx, claims.UserID, req.RefreshToken)
    if err != nil {
        return nil, fmt.Errorf("验证refresh token失败: %w", err)
    }
    if !isValid {
        return nil, errors.New("refresh token已失效或不存在")
    }

    // 4. 检查用户当前状态
    user, err := s.userRepo.GetByID(claims.UserID)
    if err != nil {
        return nil, errors.New("用户不存在")
    }
    if user.Status == "banned" {
        // 封禁用户，立即删除其所有refresh token
        _ = s.refreshTokenRepo.Delete(ctx, claims.UserID)
        return nil, errors.New("账户已被封禁，无法刷新token")
    }
    if user.Status == "inactive" {
        return nil, errors.New("账户未激活，无法刷新token")
    }

    // 5. 生成新的Access Token
    newAccessToken, err := s.jwtSvc.GenerateAccessToken(claims.UserID, claims.Username)
    if err != nil {
        return nil, fmt.Errorf("生成新access token失败: %w", err)
    }

    // 5. 返回新的Access Token
    return &model.RefreshTokenResponse{
        AccessToken: newAccessToken,
    }, nil
}

// Logout handles user logout by invalidating both access and refresh tokens
func (s *userService) Logout(ctx context.Context, req *model.LogoutRequest) error {
    // 1. 验证Refresh Token格式和签名
    refreshClaims, err := s.jwtSvc.ValidateToken(req.RefreshToken)
    if err != nil {
        return errors.New("无效的refresh token")
    }

    // 2. 确保这是一个Refresh Token
    if refreshClaims.TokenType != RefreshToken {
        return errors.New("提供的token不是refresh token")
    }

    // 3. 验证Access Token格式和签名
    accessClaims, err := s.jwtSvc.ValidateToken(req.AccessToken)
    if err != nil {
        return errors.New("无效的access token")
    }

    // 4. 确保这是一个Access Token
    if accessClaims.TokenType != AccessToken {
        return errors.New("提供的token不是access token")
    }

    // 5. 确保两个token属于同一用户
    if refreshClaims.UserID != accessClaims.UserID {
        return errors.New("access token和refresh token不属于同一用户")
    }

    // 6. 将Access Token加入黑名单
    // 计算access token的剩余有效时间
    remainingTTL, err := s.jwtSvc.GetTokenRemainingTTL(req.AccessToken)
    if err != nil {
        // 如果token已过期，我们仍然继续删除refresh token
        log.Printf("Access token已过期或无效: %v", err)
    } else {
        // 将access token加入黑名单，使用剩余TTL作为过期时间
        err = s.accessTokenBlacklistRepo.Add(ctx, accessClaims.UserID, req.AccessToken, remainingTTL)
        if err != nil {
            return fmt.Errorf("将access token加入黑名单失败: %w", err)
        }
    }

    // 7. 从Redis中删除Refresh Token
    err = s.refreshTokenRepo.DeleteByToken(ctx, req.RefreshToken)
    if err != nil {
        return fmt.Errorf("删除refresh token失败: %w", err)
    }

    return nil
}

// UpdateProfile handles updating user profile information
func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *model.UpdateProfileRequest) (*model.UserResponse, error) {
    // 1. 验证用户是否存在
    _, err := s.userRepo.GetByID(userID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("用户不存在")
        }
        return nil, fmt.Errorf("获取用户信息失败: %w", err)
    }

    // 2. 更新用户信息
    err = s.userRepo.UpdateProfile(userID, req.Nickname, req.Bio, req.Avatar, req.BackgroundURL)
    if err != nil {
        return nil, fmt.Errorf("更新用户信息失败: %w", err)
    }

    // 3. 获取更新后的用户信息
    updatedUser, err := s.userRepo.GetByID(userID)
    if err != nil {
        return nil, fmt.Errorf("获取更新后用户信息失败: %w", err)
    }

    return updatedUser.ToResponse(), nil
}

// SendVerificationCode 发送注册验证码
func (s *userService) SendVerificationCode(ctx context.Context, req *model.SendCodeRequest, ip string) error {
    // 1. IP频率限制检查
    count, err := s.rateLimitRepo.Increment(ctx, ip)
    if err != nil {
        // 即使频率限制检查失败，也应该记录错误但不一定立即阻断流程，
        // 取决于安全策略。这里我们选择记录并继续，但也可以返回错误。
        log.Printf("无法检查IP (%s) 的请求频率: %v", ip, err)
    }
    if count > int64(s.securityCfg.MaxRequestsPerIPPerDay) {
        return fmt.Errorf("请求过于频繁，请24小时后再试 (IP: %s)", ip)
    }

    // 2. 检查用户名是否已存在
    exists, err := s.userRepo.ExistsByUsername(req.Username)
    if err != nil {
        return fmt.Errorf("检查用户名失败: %w", err)
    }
    if exists {
        return errors.New("该用户名已被注册")
    }

    // 3. 检查邮箱是否已被注册
    exists, err = s.userRepo.ExistsByEmail(req.Email)
    if err != nil {
        return fmt.Errorf("检查邮箱失败: %w", err)
    }
    if exists {
        return errors.New("该邮箱已被注册")
    }

    // 生成验证码
    code, err := s.generateVerificationCode(verificationCodeLength)
    if err != nil {
        return fmt.Errorf("生成验证码失败: %w", err)
    }

    // 将验证码存入Redis，有效期5分钟
    if err := s.codeRepo.Set(ctx, req.Email, code, verificationCodeTTL); err != nil {
        return fmt.Errorf("存储验证码失败: %w", err)
    }

    // 发送邮件
    if err := s.mailSvc.SendVerificationCode(req.Email, code); err != nil {
        return fmt.Errorf("发送验证码邮件失败: %w", err)
    }

    return nil
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *model.UserRegisterRequest) (*model.UserResponse, error) {
    // 验证验证码
    storedCode, err := s.codeRepo.Get(ctx, req.Email)
    if err != nil {
        return nil, fmt.Errorf("获取验证码失败: %w", err)
    }
    if storedCode == "" {
        return nil, errors.New("验证码已过期或不存在，请重新获取")
    }
    if storedCode != req.VerificationCode {
        return nil, errors.New("验证码错误")
    }

    // 立即删除验证码，防止重放攻击
    if err := s.codeRepo.Delete(ctx, req.Email); err != nil {
        log.Printf("删除验证码失败: %v", err)
    }

    // 生成密码盐和哈希
    salt, err := s.generateSalt()
    if err != nil {
        return nil, fmt.Errorf("生成密码盐失败: %w", err)
    }

    passwordHash := s.hashPassword(req.Password, salt)

    // 创建用户对象
    user := &model.User{
        ID:           uuid.New(),
        Username:     req.Username,
        Email:        req.Email,
        PasswordSalt: fmt.Sprintf("%s:%s", salt, passwordHash),
        Nickname:     req.Nickname,
        Bio:          req.Bio,
        Avatar:       req.Avatar,
        Status:       "inactive", // 新用户默认为未激活状态
    }

    // 原子性创建用户，依赖数据库唯一约束处理竞态条件
    if err := s.userRepo.Create(user); err != nil {
        // 检查是否是唯一约束违反
        if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
            if strings.Contains(err.Error(), "username") {
                return nil, errors.New("用户名已存在")
            }
            if strings.Contains(err.Error(), "email") {
                return nil, errors.New("邮箱已存在")
            }
            return nil, errors.New("用户名或邮箱已存在")
        }
        return nil, fmt.Errorf("创建用户失败: %w", err)
    }

    return user.ToResponse(), nil
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(id uuid.UUID) (*model.UserResponse, error) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("用户不存在")
        }
        return nil, fmt.Errorf("获取用户失败: %w", err)
    }
    return user.ToResponse(), nil
}

// GetByUsername 根据用户名获取用户
func (s *userService) GetByUsername(username string) (*model.UserResponse, error) {
    user, err := s.userRepo.GetByUsername(username)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("用户不存在")
        }
        return nil, fmt.Errorf("获取用户失败: %w", err)
    }
    return user.ToResponse(), nil
}

// ValidatePassword 验证密码
func (s *userService) ValidatePassword(username, password string) (*model.User, error) {
    user, err := s.userRepo.GetByUsername(username)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("用户名或密码错误")
        }
        return nil, fmt.Errorf("查询用户失败: %w", err)
    }

    // 解析存储的密码盐和哈希
    parts := strings.Split(user.PasswordSalt, ":")
    if len(parts) != 2 {
        return nil, errors.New("密码格式错误")
    }

    salt := parts[0]
    storedHash := parts[1]

    // 计算输入密码的哈希
    inputHash := s.hashPassword(password, salt)

    // 比较哈希值
    if inputHash != storedHash {
        return nil, errors.New("用户名或密码错误")
    }

    // 检查用户状态
    if user.Status == "banned" {
        return nil, errors.New("账户已被封禁，无法登录")
    }
    if user.Status == "inactive" {
        // 未激活账户：若在注册后宽限期内，允许登录；否则要求先激活
        if time.Since(user.CreatedAt) > activationGracePeriod {
            return nil, errors.New("账户未激活，无法登录")
        }
    }

    return user, nil
}

// SendResetPasswordCode 发送重置密码验证码
func (s *userService) SendResetPasswordCode(ctx context.Context, req *model.SendResetCodeRequest, ip string) error {
    // 1. IP频率限制检查
    count, err := s.rateLimitRepo.Increment(ctx, ip)
    if err != nil {
        log.Printf("无法检查IP (%s) 的请求频率: %v", ip, err)
    }
    if count > int64(s.securityCfg.MaxRequestsPerIPPerDay) {
        return fmt.Errorf("请求过于频繁，请24小时后再试 (IP: %s)", ip)
    }

    // 2. 检查邮箱是否存在注册用户
    exists, err := s.userRepo.ExistsByEmail(req.Email)
    if err != nil {
        return fmt.Errorf("检查邮箱失败: %w", err)
    }
    if !exists {
        // 为了安全考虑，即使邮箱不存在也返回成功，避免邮箱枚举攻击
        return nil
    }

    // 3. 生成验证码
    code, err := s.generateVerificationCode(verificationCodeLength)
    if err != nil {
        return fmt.Errorf("生成验证码失败: %w", err)
    }

    // 4. 将验证码存入Redis，使用专用前缀区分重置密码验证码
    resetCodeKey := "reset:" + req.Email
    if err := s.codeRepo.Set(ctx, resetCodeKey, code, verificationCodeTTL); err != nil {
        return fmt.Errorf("存储验证码失败: %w", err)
    }

    // 5. 发送重置密码邮件
    if err := s.mailSvc.SendResetPasswordCode(req.Email, code); err != nil {
        return fmt.Errorf("发送重置密码验证码邮件失败: %w", err)
    }

    return nil
}

// ResetPassword 重置密码
func (s *userService) ResetPassword(ctx context.Context, req *model.ResetPasswordRequest) error {
    // 1. 验证重置密码验证码
    resetCodeKey := "reset:" + req.Email
    storedCode, err := s.codeRepo.Get(ctx, resetCodeKey)
    if err != nil {
        return fmt.Errorf("获取验证码失败: %w", err)
    }
    if storedCode == "" {
        return errors.New("验证码已过期或不存在，请重新获取")
    }
    if storedCode != req.VerificationCode {
        return errors.New("验证码错误")
    }

    // 2. 检查用户是否存在
    user, err := s.userRepo.GetByEmail(req.Email)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("邮箱未注册")
        }
        return fmt.Errorf("查询用户失败: %w", err)
    }

    // 3. 生成新的密码盐和哈希
    salt, err := s.generateSalt()
    if err != nil {
        return fmt.Errorf("生成密码盐失败: %w", err)
    }

    passwordHash := s.hashPassword(req.NewPassword, salt)
    newPasswordSalt := fmt.Sprintf("%s:%s", salt, passwordHash)

    // 4. 更新用户密码
    if err := s.userRepo.UpdatePassword(user.ID, newPasswordSalt); err != nil {
        return fmt.Errorf("更新密码失败: %w", err)
    }

    // 5. 删除已使用的验证码
    _ = s.codeRepo.Delete(ctx, resetCodeKey)

    // 6. 撤销该用户的所有refresh token，强制重新登录
    _ = s.refreshTokenRepo.Delete(ctx, user.ID)

    return nil
}

// generateSalt
func (s *userService) generateSalt() (string, error) {
    bytes := make([]byte, 16)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}

// hashPassword
func (s *userService) hashPassword(password, salt string) string {
    hash := sha256.Sum256([]byte(password + salt))
    return hex.EncodeToString(hash[:])
}

// generateVerificationCode 生成指定长度的数字验证码
func (s *userService) generateVerificationCode(length int) (string, error) {
    code := ""
    for i := 0; i < length; i++ {
        n, err := rand.Int(rand.Reader, big.NewInt(10))
        if err != nil {
            return "", err
        }
        code += n.String()
    }
    return code, nil
}

// UpdateUserStatus 更新用户状态
func (s *userService) UpdateUserStatus(id uint, status string) error {
    return s.userRepo.UpdateStatus(id, status)
}

// AdminUpdateUserPassword 允许管理员为指定用户直接设置新密码
func (s *userService) AdminUpdateUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
    // 1. 检查用户是否存在
    _, err := s.userRepo.GetByID(userID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("用户不存在")
        }
        return fmt.Errorf("查询用户失败: %w", err)
    }

    // 2. 生成新的盐和哈希
    salt, err := s.generateSalt()
    if err != nil {
        return fmt.Errorf("生成密码盐失败: %w", err)
    }
    passwordHash := s.hashPassword(newPassword, salt)
    newPasswordSalt := fmt.Sprintf("%s:%s", salt, passwordHash)

    // 3. 更新密码
    if err := s.userRepo.UpdatePassword(userID, newPasswordSalt); err != nil {
        return fmt.Errorf("更新密码失败: %w", err)
    }

    // 4. 撤销该用户所有refresh token，强制重新登录
    _ = s.refreshTokenRepo.Delete(ctx, userID)

    return nil
}

// userService 用户服务实现
type userService struct {
	userRepo                 repository.UserRepository
	deviceRepo               repository.DeviceRepository
	codeRepo                 repository.CodeRepository
	refreshTokenRepo         repository.RefreshTokenRepository
	rateLimitRepo            repository.RateLimitRepository
	accessTokenBlacklistRepo repository.AccessTokenBlacklistRepository
	mailSvc                  MailService
	jwtSvc                   JwtService
	securityCfg              *config.SecurityConfig
}

// NewUserService 创建用户服务实例
func NewUserService(
	userRepo repository.UserRepository,
	deviceRepo repository.DeviceRepository,
	codeRepo repository.CodeRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	rateLimitRepo repository.RateLimitRepository,
	accessTokenBlacklistRepo repository.AccessTokenBlacklistRepository,
	mailSvc MailService,
	jwtSvc JwtService,
	securityCfg *config.SecurityConfig,
) UserService {
	return &userService{
		userRepo:                 userRepo,
		deviceRepo:               deviceRepo,
		codeRepo:                 codeRepo,
		refreshTokenRepo:         refreshTokenRepo,
		rateLimitRepo:            rateLimitRepo,
		accessTokenBlacklistRepo: accessTokenBlacklistRepo,
		mailSvc:                  mailSvc,
		jwtSvc:                   jwtSvc,
		securityCfg:              securityCfg,
	}
}

// Login handles user login.
func (s *userService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	// 1. 验证用户名和密码
	user, err := s.ValidatePassword(req.Username, req.Password)
	if err != nil {
		// 检查是否是账户状态相关的错误，如果是则直接返回具体错误信息
		errMsg := err.Error()
		if errMsg == "账户已被封禁，无法登录" || errMsg == "账户未激活，无法登录" {
			return nil, err
		}
		// 其他错误（如用户名不存在、密码错误等）统一返回通用错误信息，避免用户枚举攻击
		return nil, errors.New("用户名或密码错误")
	}

	// 首次登录（LastLoginAt为空）跳过设备验证，直接签发Token
	if user.LastLoginAt == nil {
		// 若提供了设备指纹，则记录并信任该设备
		if strings.TrimSpace(req.DeviceID) != "" {
			now := time.Now()
			// 尝试查找现有设备
			if d, derr := s.deviceRepo.GetDeviceByUserAndFingerprint(user.ID, req.DeviceID); derr == nil && d != nil {
				d.IsTrusted = true
				d.IPAddress = req.IPAddress
				d.UserAgent = req.UserAgent
				if req.DeviceName != "" { d.DeviceName = req.DeviceName }
				if req.DeviceType != "" { d.DeviceType = req.DeviceType }
				d.LastLoginAt = &now
				_ = s.deviceRepo.UpdateDevice(d)
			} else {
				dev := &model.UserDevice{
					ID:          uuid.New(),
					UserID:      user.ID,
					DeviceID:    req.DeviceID,
					DeviceName:  req.DeviceName,
					DeviceType:  req.DeviceType,
					UserAgent:   req.UserAgent,
					IPAddress:   req.IPAddress,
					IsTrusted:   true,
					LastLoginAt: func() *time.Time { t := now; return &t }(),
				}
				_ = s.deviceRepo.CreateDevice(dev)
			}
		}
		return s.issueTokenPair(ctx, user)
	}

	// 非首次登录：需要设备指纹并进行陌生设备验证
	if strings.TrimSpace(req.DeviceID) == "" {
		return &model.LoginResponse{
			User:                 user.ToResponse(),
			VerificationRequired: true,
		}, errors.New("为了账户安全，请提供设备指纹信息")
	}

	// 2. 检查设备是否已信任
	var device *model.UserDevice
	d, derr := s.deviceRepo.GetDeviceByUserAndFingerprint(user.ID, req.DeviceID)
	if derr != nil {
		if errors.Is(derr, gorm.ErrRecordNotFound) {
			device = nil
		} else {
			return nil, fmt.Errorf("查询设备失败: %w", derr)
		}
	} else {
		device = d
	}

	// 已存在且已信任 -> 更新登录信息并直接签发Token
	if device != nil && device.IsTrusted {
		now := time.Now()
		device.IPAddress = req.IPAddress
		device.UserAgent = req.UserAgent
		if req.DeviceName != "" {
			device.DeviceName = req.DeviceName
		}
		if req.DeviceType != "" {
			device.DeviceType = req.DeviceType
		}
		device.LastLoginAt = &now
		_ = s.deviceRepo.UpdateDevice(device)
		return s.issueTokenPair(ctx, user)
	}

	// 未信任设备：若带验证码则校验，否则发送验证码并提示二次验证
	if strings.TrimSpace(req.DeviceVerifyCode) != "" {
		v, err := s.deviceRepo.GetLatestPendingVerification(user.ID, req.DeviceID)
		if err != nil || v == nil || v.IsVerified || time.Now().After(v.ExpiresAt) {
			return nil, errors.New("验证码已过期或不存在，请重新获取")
		}

		// 检查尝试次数限制
		if v.AttemptCount >= 5 {
			return nil, errors.New("验证码尝试次数过多，请重新获取")
		}

		if v.VerificationCode != req.DeviceVerifyCode {
			if err := s.deviceRepo.IncrementVerificationAttempt(v.ID); err != nil {
				log.Printf("增加验证尝试次数失败: %v", err)
			}
			return nil, errors.New("验证码错误")
		}

		// 验证通过
		if err := s.deviceRepo.MarkVerificationVerified(v.ID); err != nil {
			return nil, fmt.Errorf("标记设备验证通过失败: %w", err)
		}

		// 信任并更新/创建设备
		now := time.Now()
		if device != nil {
			device.IsTrusted = true
			device.IPAddress = req.IPAddress
			device.UserAgent = req.UserAgent
			if req.DeviceName != "" {
				device.DeviceName = req.DeviceName
			}
			if req.DeviceType != "" {
				device.DeviceType = req.DeviceType
			}
			device.LastLoginAt = &now
			if err := s.deviceRepo.UpdateDevice(device); err != nil {
				return nil, fmt.Errorf("更新设备失败: %w", err)
			}
		} else {
			dev := &model.UserDevice{
				ID:         uuid.New(),
				UserID:     user.ID,
				DeviceID:   req.DeviceID,
				DeviceName: req.DeviceName,
				DeviceType: req.DeviceType,
				UserAgent:  req.UserAgent,
				IPAddress:  req.IPAddress,
				IsTrusted:  true,
				LastLoginAt: func() *time.Time { t := now; return &t }(),
			}
			if err := s.deviceRepo.CreateDevice(dev); err != nil {
				return nil, fmt.Errorf("创建设备失败: %w", err)
			}
		}

		// 通过后签发token
		return s.issueTokenPair(ctx, user)
	}

	// 发送设备验证码并返回需要验证标记
	code, err := s.generateVerificationCode(verificationCodeLength)
	if err != nil {
		return nil, fmt.Errorf("生成验证码失败: %w", err)
	}
	v := &model.DeviceVerification{
		ID:               uuid.New(),
		UserID:           user.ID,
		DeviceID:         req.DeviceID,
		VerificationCode: code,
		AttemptCount:     0,
		IPAddress:        req.IPAddress,
		UserAgent:        req.UserAgent,
		IsVerified:       false,
		ExpiresAt:        time.Now().Add(verificationCodeTTL),
	}
	if err := s.deviceRepo.CreateVerification(v); err != nil {
		return nil, fmt.Errorf("创建设备验证记录失败: %w", err)
	}

	// 发送邮件
	if err := s.mailSvc.SendDeviceVerificationCode(user.Email, code, firstNonEmpty(req.DeviceName, "未知设备"), req.IPAddress, req.UserAgent); err != nil {
		return nil, fmt.Errorf("发送设备验证邮件失败: %w", err)
	}

	return &model.LoginResponse{
		User:                 user.ToResponse(),
		VerificationRequired: true,
	}, nil
}

func (s *userService) issueTokenPair(ctx context.Context, user *model.User) (*model.LoginResponse, error) {
	tokenPair, err := s.jwtSvc.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}
	refreshTokenExpiration := time.Duration(s.securityCfg.JwtRefreshTokenExpiresInDays) * 24 * time.Hour
	if err := s.refreshTokenRepo.Store(ctx, user.ID, tokenPair.RefreshToken, refreshTokenExpiration); err != nil {
		return nil, fmt.Errorf("存储refresh token失败: %w", err)
	}
	// 更新用户最后登录时间（忽略错误以不中断登录流程）
	now := time.Now()
	_ = s.userRepo.UpdateLastLoginAt(user.ID, now)
	user.LastLoginAt = &now
	return &model.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         user.ToResponse(),
	}, nil
}

// GetUsersForAdmin 获取用户列表（管理员用）
func (s *userService) GetUsersForAdmin(page, limit int, search string) ([]*model.UserResponse, int64, error) {
	users, total, err := s.userRepo.GetUsersWithPagination(page, limit, search)
	if err != nil {
		return nil, 0, err
	}

	var userResponses []*model.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, &model.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			Nickname:    user.Nickname,
			Avatar:      user.Avatar,
			Status:      user.Status,
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		})
	}

	return userResponses, total, nil
}

// GetUserByID 根据ID获取用户（管理员用）
func (s *userService) GetUserByID(id uint) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByUintID(id)
	if err != nil {
		return nil, err
	}

	return &model.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Nickname:    user.Nickname,
		Bio:         user.Bio,
		Avatar:      user.Avatar,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil

}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	return s.userRepo.DeleteByUintID(id)
}

// GetUserStats 获取用户统计信息
func (s *userService) GetUserStats() (map[string]interface{}, error) {
	stats, err := s.userRepo.GetUserStats()
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// UpdateUserStatusByUUID 根据UUID更新用户状态
func (s *userService) UpdateUserStatusByUUID(id uuid.UUID, status string) error {
	return s.userRepo.UpdateStatusByUUID(id, status)
}

// DeleteUserByUUID 根据UUID删除用户
func (s *userService) DeleteUserByUUID(id uuid.UUID) error {
	// 1. 检查用户是否存在
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 2. 检查是否为系统保护用户（可根据需要扩展）
	if user.Username == "admin" || user.Username == "system" {
		return errors.New("系统保护用户不能删除")
	}

	// 3. 执行软删除
	return s.userRepo.Delete(id)
}

// SendActivationCode 发送账户激活验证码
func (s *userService) SendActivationCode(ctx context.Context, req *model.SendActivationCodeRequest, ip string) error {
	// 1. IP频率限制检查
	count, err := s.rateLimitRepo.Increment(ctx, ip)
	if err != nil {
		log.Printf("无法检查IP (%s) 的请求频率: %v", ip, err)
	}
	if count > int64(s.securityCfg.MaxRequestsPerIPPerDay) {
		return errors.New("请求过于频繁，请稍后再试")
	}

	// 2. 检查用户是否存在
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("邮箱不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 3. 检查用户状态
	if user.Status == "active" {
		return errors.New("账户已激活，无需重复激活")
	}
	if user.Status == "banned" {
		return errors.New("账户已被封禁，无法激活")
	}

	// 4. 生成验证码
	code, err := s.generateVerificationCode(6)
	if err != nil {
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	// 5. 存储验证码到Redis（5分钟过期）
	if err := s.codeRepo.Set(ctx, req.Email, code, 5*time.Minute); err != nil {
		return fmt.Errorf("存储验证码失败: %w", err)
	}

	// 6. 发送激活邮件
	if err := s.mailSvc.SendVerificationCode(req.Email, code); err != nil {
		return fmt.Errorf("发送激活邮件失败: %w", err)
	}

	return nil
}

// ActivateAccount 激活账户
func (s *userService) ActivateAccount(ctx context.Context, req *model.ActivateAccountRequest) error {
	// 1. 验证验证码
	storedCode, err := s.codeRepo.Get(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("获取验证码失败: %w", err)
	}
	if storedCode == "" {
		return errors.New("验证码已过期或不存在，请重新获取")
	}
	if storedCode != req.VerificationCode {
		return errors.New("验证码错误")
	}

	// 2. 检查用户是否存在
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("邮箱不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 3. 检查用户状态
	if user.Status == "active" {
		// 删除验证码
		_ = s.codeRepo.Delete(ctx, req.Email)
		return errors.New("账户已激活")
	}
	if user.Status == "banned" {
		return errors.New("账户已被封禁，无法激活")
	}

	// 4. 激活账户
	err = s.userRepo.UpdateStatusByUUID(user.ID, "active")
	if err != nil {
		return fmt.Errorf("激活账户失败: %w", err)
	}

	// 5. 删除验证码
	_ = s.codeRepo.Delete(ctx, req.Email)

	return nil
}