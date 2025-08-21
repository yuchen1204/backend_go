package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Friendship 好友关系（双向各一条记录）
// 唯一约束 (user_id, friend_id)
// remark 为“我对好友的备注”，仅对自己可见

type Friendship struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index;uniqueIndex:uidx_user_friend"`
	FriendID  uuid.UUID      `json:"friend_id" gorm:"type:uuid;not null;index;uniqueIndex:uidx_user_friend"`
	Remark    string         `json:"remark" gorm:"type:varchar(50)"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Friendship) TableName() string { return "friendships" }
