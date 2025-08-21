package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FriendRequest 好友请求
// 说明：采用状态机控制，pending -> accepted/rejected/cancelled
// 申请备注记录在 note
// 建议在业务层保证 (requester_id, receiver_id) pending 唯一
// 索引：requester_id, receiver_id, status

type FriendRequestStatus string

const (
	FriendRequestPending  FriendRequestStatus = "pending"
	FriendRequestAccepted FriendRequestStatus = "accepted"
	FriendRequestRejected FriendRequestStatus = "rejected"
	FriendRequestCanceled FriendRequestStatus = "cancelled"
)

type FriendRequest struct {
	ID          uuid.UUID           `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RequesterID uuid.UUID           `json:"requester_id" gorm:"type:uuid;not null;index:idx_req_rec_status"`
	ReceiverID  uuid.UUID           `json:"receiver_id" gorm:"type:uuid;not null;index:idx_req_rec_status"`
	Status      FriendRequestStatus `json:"status" gorm:"type:varchar(20);not null;index:idx_req_rec_status"`
	Note        string              `json:"note" gorm:"type:varchar(200)"` // 申请备注，业务层做长度校验

	HandledAt *time.Time    `json:"handled_at"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (FriendRequest) TableName() string { return "friend_requests" }
