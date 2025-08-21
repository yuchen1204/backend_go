package repository

import (
	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatRoomRepository 一对一聊天房间仓储
// 房间参与者严格两个，使用 (user_a_id, user_b_id) 按字典序唯一

type ChatRoomRepository interface {
	GetByID(id uuid.UUID) (*model.ChatRoom, error)
	GetOrCreateByUsers(a, b uuid.UUID) (*model.ChatRoom, error)
	DeactivateByUsers(a, b uuid.UUID) error
}

type chatRoomRepository struct {
	db *gorm.DB
}

func NewChatRoomRepository(db *gorm.DB) ChatRoomRepository { return &chatRoomRepository{db: db} }

func orderPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() { return a, b }
	return b, a
}

func (r *chatRoomRepository) GetByID(id uuid.UUID) (*model.ChatRoom, error) {
	var room model.ChatRoom
	if err := r.db.First(&room, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *chatRoomRepository) GetOrCreateByUsers(a, b uuid.UUID) (*model.ChatRoom, error) {
	a1, b1 := orderPair(a, b)
	var room model.ChatRoom
	if err := r.db.Where("user_a_id = ? AND user_b_id = ?", a1, b1).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			room = model.ChatRoom{UserAID: a1, UserBID: b1, Status: "active"}
			if err := r.db.Create(&room).Error; err != nil { return nil, err }
			return &room, nil
		}
		return nil, err
	}
	// 确保恢复为 active
	if room.Status != "active" {
		if err := r.db.Model(&model.ChatRoom{}).Where("id = ?", room.ID).Update("status", "active").Error; err != nil {
			return nil, err
		}
		room.Status = "active"
	}
	return &room, nil
}

func (r *chatRoomRepository) DeactivateByUsers(a, b uuid.UUID) error {
	a1, b1 := orderPair(a, b)
	return r.db.Model(&model.ChatRoom{}).
		Where("user_a_id = ? AND user_b_id = ?", a1, b1).
		Update("status", "inactive").Error
}
