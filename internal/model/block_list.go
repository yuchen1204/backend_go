package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BlockList 拉黑名单（user_id 拉黑 blocked_user_id）
// 唯一约束 (user_id, blocked_user_id)

type BlockList struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index;uniqueIndex:uidx_user_blocked"`
	BlockedUserID  uuid.UUID      `json:"blocked_user_id" gorm:"type:uuid;not null;index;uniqueIndex:uidx_user_blocked"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func (BlockList) TableName() string { return "block_lists" }
