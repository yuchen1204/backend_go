package model

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null;size:100"`
	PasswordSalt string    `json:"-" gorm:"not null;size:255"` // 密码加盐哈希，不在JSON中返回
	Nickname     string    `json:"nickname" gorm:"size:100"`
	Bio          string    `json:"bio" gorm:"type:text"`
	Avatar       string    `json:"avatar" gorm:"size:255"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserRegisterRequest 用户注册请求结构
type UserRegisterRequest struct {
	Username         string `json:"username" binding:"required,min=3,max=50" example:"testuser"`
	Email            string `json:"email" binding:"required,email" example:"test@example.com"`
	Password         string `json:"password" binding:"required,min=6,max=100" example:"password123"`
	VerificationCode string `json:"verification_code" binding:"required,len=6" example:"123456"`
	Nickname         string `json:"nickname" binding:"omitempty,max=100" example:"测试用户"`
	Bio              string `json:"bio" binding:"omitempty,max=500" example:"这是我的个人简介"`
	Avatar           string `json:"avatar" binding:"omitempty,url" example:"https://example.com/avatar.jpg"`
}

// SendCodeRequest 发送验证码请求结构
type SendCodeRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"testuser"`
	Email    string `json:"email" binding:"required,email" example:"test@example.com"`
}

// LoginRequest 用户登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"testuser"`
	Password string `json:"password" binding:"required" example:"password123"`
	// 设备指纹（建议为客户端计算的SHA256十六进制字符串，长度64）。
	// 若提供该字段，系统将进行陌生设备校验；未提供则按旧逻辑直接登录。
	DeviceID         string `json:"device_id" binding:"omitempty,max=128" example:"e3b0c44298fc1c149afbf4c8996fb924..."`
	// 可选：设备名称、类型用于记录（不参与校验）
	DeviceName       string `json:"device_name" binding:"omitempty,max=100" example:"John's iPhone"`
	DeviceType       string `json:"device_type" binding:"omitempty,oneof=mobile desktop tablet" example:"mobile"`
	// 如果是第二步校验，客户端可在同一登录接口提交邮箱验证码完成验证
	DeviceVerifyCode string `json:"device_verification_code" binding:"omitempty,len=6" example:"123456"`
	// 由服务器端在处理器中自动填充的请求来源信息
	IPAddress        string `json:"ip_address" binding:"omitempty,max=45" example:"203.0.113.1"`
	UserAgent        string `json:"user_agent" binding:"omitempty,max=500" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64)..."`
}

// LoginResponse 用户登录响应结构
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *UserResponse `json:"user"`
	// 若为陌生设备首次登录，将不会返回token，而是提示需要进行设备验证码验证
	VerificationRequired bool `json:"verification_required,omitempty"`
}

// RefreshTokenRequest 刷新Token请求结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshTokenResponse 刷新Token响应结构
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// LogoutRequest 登出请求结构
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	AccessToken  string `json:"access_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// UpdateProfileRequest 更新用户信息请求结构
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,max=100" example:"新昵称"`
	Bio      string `json:"bio" binding:"omitempty,max=500" example:"新的个人简介"`
	Avatar   string `json:"avatar" binding:"omitempty,url" example:"https://example.com/new-avatar.jpg"`
}

// SendResetCodeRequest 发送重置密码验证码请求结构
type SendResetCodeRequest struct {
	Email string `json:"email" binding:"required,email" example:"test@example.com"`
}

// ResetPasswordRequest 重置密码请求结构
type ResetPasswordRequest struct {
	Email            string `json:"email" binding:"required,email" example:"test@example.com"`
	VerificationCode string `json:"verification_code" binding:"required,len=6" example:"123456"`
	NewPassword      string `json:"new_password" binding:"required,min=6,max=100" example:"newpassword123"`
}

// UserResponse 用户响应结构（不包含敏感信息）
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	Bio       string    `json:"bio"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse 将User转换为UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Nickname:  u.Nickname,
		Bio:       u.Bio,
		Avatar:    u.Avatar,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
} 