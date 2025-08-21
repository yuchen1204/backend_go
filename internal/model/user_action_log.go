package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserActionLog 用户行为日志
// 记录用户的关键操作：登录、更新资料、重置密码等
// 注意：为避免敏感信息泄漏，Details 内不应包含明文密码
type UserActionLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Username   string     `gorm:"size:64;index" json:"username"`
	Action     string     `gorm:"size:64;index;not null" json:"action"`
	DeviceID   string     `gorm:"size:128" json:"device_id"`
	DeviceName string     `gorm:"size:100" json:"device_name"`
	DeviceType string     `gorm:"size:20" json:"device_type"`
	IPAddress  string     `gorm:"size:64" json:"ip_address"`
	UserAgent  string     `gorm:"size:255" json:"user_agent"`
	Details    string     `gorm:"type:text" json:"details"`
	CreatedAt  time.Time  `json:"created_at"`
}

// TableName 指定表名
func (UserActionLog) TableName() string { return "user_action_logs" }

// BeforeCreate 钩子初始化UUID
func (u *UserActionLog) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
