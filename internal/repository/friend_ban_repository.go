package repository

import (
	"backend/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FriendBanRepository 好友申请禁用仓储

type FriendBanRepository interface {
	SetBan(userID uuid.UUID, reason string, until time.Time) error
	RemoveBan(userID uuid.UUID) error
	GetActiveBan(userID uuid.UUID, now time.Time) (*model.FriendBan, error)
}

type friendBanRepository struct{ db *gorm.DB }

func NewFriendBanRepository(db *gorm.DB) FriendBanRepository { return &friendBanRepository{db: db} }

func (r *friendBanRepository) SetBan(userID uuid.UUID, reason string, until time.Time) error {
	ban := &model.FriendBan{UserID: userID}
	return r.db.Where("user_id = ?", userID).Assign(map[string]any{
		"reason":       reason,
		"banned_until": until,
	}).FirstOrCreate(ban).Error
}

func (r *friendBanRepository) RemoveBan(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.FriendBan{}).Error
}

func (r *friendBanRepository) GetActiveBan(userID uuid.UUID, now time.Time) (*model.FriendBan, error) {
	var ban model.FriendBan
	if err := r.db.Where("user_id = ? AND banned_until > ?", userID, now).First(&ban).Error; err != nil {
		return nil, err
	}
	return &ban, nil
}
