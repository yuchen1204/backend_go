package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminActionLog 管理员行为日志
// 记录管理员在面板中的关键操作，便于审计与追踪
// 表字段命名遵循已有模型风格（使用UUID主键）
type AdminActionLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AdminUsername string    `gorm:"size:64;index;not null" json:"admin_username"`
	Action        string    `gorm:"size:64;index;not null" json:"action"`
	TargetUserID  *uuid.UUID `gorm:"type:uuid;index" json:"target_user_id,omitempty"`
	Details       string    `gorm:"type:text" json:"details"`
	IPAddress     string    `gorm:"size:64" json:"ip_address"`
	UserAgent     string    `gorm:"size:255" json:"user_agent"`
	CreatedAt     time.Time `json:"created_at"`
}

// BeforeCreate 钩子初始化UUID
func (a *AdminActionLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// AdminLogCreateRequest 创建日志请求体
// 允许前端或其他后端动作显式记录日志
// action 示例：update_user_status, delete_user, reset_user_password
// details 建议为JSON字符串，以便扩展
 type AdminLogCreateRequest struct {
	Action       string     `json:"action" binding:"required"`
	TargetUserID *uuid.UUID `json:"target_user_id"`
	Details      string     `json:"details"`
}

// AdminLogListResponse 列表返回结构
 type AdminLogListResponse struct {
	Logs  []AdminActionLog `json:"logs"`
	Total int64            `json:"total"`
}
