package model

import (
	"time"

	"github.com/google/uuid"
)

// ChatRoom 一对一聊天房间（仅两个用户）
// 唯一约束 (user_a_id, user_b_id) 按字典序存储（较小者为 A）
// Status: active/inactive

type ChatRoom struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserAID  uuid.UUID `json:"user_a_id" gorm:"type:uuid;not null;index;uniqueIndex:uidx_chat_pair"`
	UserBID  uuid.UUID `json:"user_b_id" gorm:"type:uuid;not null;index;uniqueIndex:uidx_chat_pair"`
	Status   string    `json:"status" gorm:"type:varchar(16);not null;default:'active'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ChatRoom) TableName() string { return "chat_rooms" }
