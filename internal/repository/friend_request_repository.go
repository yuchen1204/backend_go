package repository

import (
	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FriendRequestRepository 好友请求仓储
// 仅定义必要方法，具体实现后续补充细节（索引/唯一约束由模型定义保证）

type FriendRequestRepository interface {
	Create(req *model.FriendRequest) error
	GetByID(id uuid.UUID) (*model.FriendRequest, error)
	ListIncoming(userID uuid.UUID, status model.FriendRequestStatus, page, limit int) ([]model.FriendRequest, int64, error)
	ListOutgoing(userID uuid.UUID, status model.FriendRequestStatus, page, limit int) ([]model.FriendRequest, int64, error)
	FindPending(requesterID, receiverID uuid.UUID) (*model.FriendRequest, error)
	Update(req *model.FriendRequest) error
	DeleteByID(id uuid.UUID) error
}

type friendRequestRepository struct {
	db *gorm.DB
}

func NewFriendRequestRepository(db *gorm.DB) FriendRequestRepository {
	return &friendRequestRepository{db: db}
}

func (r *friendRequestRepository) Create(req *model.FriendRequest) error {
	return r.db.Create(req).Error
}

func (r *friendRequestRepository) GetByID(id uuid.UUID) (*model.FriendRequest, error) {
	var fr model.FriendRequest
	if err := r.db.First(&fr, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &fr, nil
}

func (r *friendRequestRepository) ListIncoming(userID uuid.UUID, status model.FriendRequestStatus, page, limit int) ([]model.FriendRequest, int64, error) {
	var list []model.FriendRequest
	var total int64
	q := r.db.Model(&model.FriendRequest{}).Where("receiver_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *friendRequestRepository) ListOutgoing(userID uuid.UUID, status model.FriendRequestStatus, page, limit int) ([]model.FriendRequest, int64, error) {
	var list []model.FriendRequest
	var total int64
	q := r.db.Model(&model.FriendRequest{}).Where("requester_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *friendRequestRepository) FindPending(requesterID, receiverID uuid.UUID) (*model.FriendRequest, error) {
	var fr model.FriendRequest
	if err := r.db.Where("requester_id = ? AND receiver_id = ? AND status = ?", requesterID, receiverID, model.FriendRequestPending).First(&fr).Error; err != nil {
		return nil, err
	}
	return &fr, nil
}

func (r *friendRequestRepository) Update(req *model.FriendRequest) error {
	return r.db.Save(req).Error
}

func (r *friendRequestRepository) DeleteByID(id uuid.UUID) error {
	return r.db.Delete(&model.FriendRequest{}, "id = ?", id).Error
}
