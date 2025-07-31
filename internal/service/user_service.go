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
}

// userService 用户服务实现
type userService struct {
	userRepo                 repository.UserRepository
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
		// To avoid user enumeration, return a generic error message.
		return nil, errors.New("用户名或密码错误")
	}

	// 2. 生成Token对
	tokenPair, err := s.jwtSvc.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	// 3. 存储Refresh Token到Redis
	refreshTokenExpiration := time.Duration(s.securityCfg.JwtRefreshTokenExpiresInDays) * 24 * time.Hour
	err = s.refreshTokenRepo.Store(ctx, user.ID, tokenPair.RefreshToken, refreshTokenExpiration)
	if err != nil {
		return nil, fmt.Errorf("存储refresh token失败: %w", err)
	}

	// 4. 构建响应
	loginResponse := &model.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         user.ToResponse(),
	}

	return loginResponse, nil
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

	// 4. 生成新的Access Token
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
	err = s.userRepo.UpdateProfile(userID, req.Nickname, req.Bio, req.Avatar)
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

	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在 (双重检查)
	exists, err = s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return nil, errors.New("邮箱已存在")
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
	}

	// 保存到数据库
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 注册成功后删除验证码
	_ = s.codeRepo.Delete(ctx, req.Email)

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

// validateRegisterRequest - this function is no longer needed as gin does the validation
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