package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserDevice 用户设备表
// 记录用户受信任设备，用于判断是否为陌生设备
type UserDevice struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index:uidx_user_device,unique"`
	DeviceID    string         `json:"device_id" gorm:"not null;size:64;index:uidx_user_device,unique"`       // 设备指纹哈希（SHA256）
	DeviceName  string         `json:"device_name" gorm:"size:100"`                   // 用户自定义名称
	DeviceType  string         `json:"device_type" gorm:"size:50"`                    // mobile/desktop/tablet
	UserAgent   string         `json:"user_agent" gorm:"size:500"`
	IPAddress   string         `json:"ip_address" gorm:"size:45"`
	Location    string         `json:"location" gorm:"size:100"`                      // 可选：地理位置（城市/国家）
	IsTrusted   bool           `json:"is_trusted" gorm:"default:false"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (UserDevice) TableName() string {
	return "user_devices"
}

// DeviceVerification 设备验证表
// 当检测到陌生设备登录时，发送验证码并在此保存验证记录
type DeviceVerification struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID           uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	DeviceID         string         `json:"device_id" gorm:"not null;size:64;index"`
	VerificationCode string         `json:"-" gorm:"not null;size:6"`                // 6位验证码
	AttemptCount     int            `json:"attempt_count" gorm:"not null;default:0"` // 尝试次数
	IPAddress        string         `json:"ip_address" gorm:"size:45"`
	UserAgent        string         `json:"user_agent" gorm:"size:500"`
	IsVerified       bool           `json:"is_verified" gorm:"default:false"`
	ExpiresAt        time.Time      `json:"expires_at"`
	VerifiedAt       *time.Time     `json:"verified_at"`
	CreatedAt        time.Time      `json:"created_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (DeviceVerification) TableName() string {
	return "device_verifications"
}
