package repository

import (
	"backend/internal/model"
	"context"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

// UserActionLogRepository 用户行为日志仓库
type UserActionLogRepository interface {
	Create(ctx context.Context, log *model.UserActionLog) error
	// ListByUser 按用户ID分页查询用户行为日志（按创建时间倒序）
	ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.UserActionLog, int64, error)
}

type userActionLogRepository struct {
	db *gorm.DB
}

func NewUserActionLogRepository(db *gorm.DB) UserActionLogRepository {
	return &userActionLogRepository{db: db}
}

func (r *userActionLogRepository) Create(ctx context.Context, log *model.UserActionLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *userActionLogRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.UserActionLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	var (
		logs  []model.UserActionLog
		total int64
	)

	q := r.db.WithContext(ctx).Model(&model.UserActionLog{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
