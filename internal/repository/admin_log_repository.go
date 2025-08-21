package repository

import (
	"context"

	"backend/internal/model"
	"gorm.io/gorm"
)

// AdminLogRepository 管理员日志仓库接口
 type AdminLogRepository interface {
	Create(ctx context.Context, log *model.AdminActionLog) error
	List(ctx context.Context, page, limit int, adminUsername, action string) ([]model.AdminActionLog, int64, error)
}

// adminLogRepository 实现
 type adminLogRepository struct {
	db *gorm.DB
}

// NewAdminLogRepository 创建实例
 func NewAdminLogRepository(db *gorm.DB) AdminLogRepository {
	return &adminLogRepository{db: db}
}

func (r *adminLogRepository) Create(ctx context.Context, log *model.AdminActionLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *adminLogRepository) List(ctx context.Context, page, limit int, adminUsername, action string) ([]model.AdminActionLog, int64, error) {
	var logs []model.AdminActionLog
	var total int64

	q := r.db.WithContext(ctx).Model(&model.AdminActionLog{})
	if adminUsername != "" {
		q = q.Where("admin_username = ?", adminUsername)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
