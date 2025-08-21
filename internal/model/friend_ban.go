package model

import (
	"time"

	"github.com/google/uuid"
)

// FriendBan 好友申请禁用表（管理员基于举报临时禁用某用户发起好友请求的能力）
// 若当前时间 < BannedUntil，则拦截该用户发起好友请求

type FriendBan struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	Reason      string    `json:"reason" gorm:"type:varchar(255)"`
	BannedUntil time.Time `json:"banned_until" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (FriendBan) TableName() string { return "friend_bans" }

// AdminSetFriendBanRequest 管理员设置好友功能封禁请求体
// BannedUntil 使用RFC3339时间（例如：2025-01-31T23:59:59Z）
type AdminSetFriendBanRequest struct {
    Reason      string    `json:"reason" binding:"required"`
    BannedUntil time.Time `json:"banned_until" binding:"required"`
}
