package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"github.com/google/uuid"
)

// UserActionLogService 用户行为日志服务接口
type UserActionLogService interface {
	Create(ctx context.Context, log *model.UserActionLog) error
	// ListByUser 按用户ID分页查询用户行为日志
	ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.UserActionLog, int64, error)
}

type userActionLogService struct {
	repo repository.UserActionLogRepository
}

func NewUserActionLogService(repo repository.UserActionLogRepository) UserActionLogService {
	return &userActionLogService{repo: repo}
}

func (s *userActionLogService) Create(ctx context.Context, log *model.UserActionLog) error {
	return s.repo.Create(ctx, log)
}

func (s *userActionLogService) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]model.UserActionLog, int64, error) {
	return s.repo.ListByUser(ctx, userID, page, limit)
}
