package service

import (
	"context"

	"backend/internal/model"
	"backend/internal/repository"
)

// AdminLogService 管理员日志服务接口
 type AdminLogService interface {
	Create(ctx context.Context, log *model.AdminActionLog) error
	List(ctx context.Context, page, limit int, adminUsername, action string) ([]model.AdminActionLog, int64, error)
}

// adminLogService 实现
 type adminLogService struct {
	repo repository.AdminLogRepository
}

// NewAdminLogService 创建服务实例
 func NewAdminLogService(repo repository.AdminLogRepository) AdminLogService {
	return &adminLogService{repo: repo}
}

func (s *adminLogService) Create(ctx context.Context, log *model.AdminActionLog) error {
	return s.repo.Create(ctx, log)
}

func (s *adminLogService) List(ctx context.Context, page, limit int, adminUsername, action string) ([]model.AdminActionLog, int64, error) {
	return s.repo.List(ctx, page, limit, adminUsername, action)
}
