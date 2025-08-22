package repository

import (
	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BlockListRepository 拉黑仓储

type BlockListRepository interface {
	Block(userID, blockedID uuid.UUID) error
	Unblock(userID, blockedID uuid.UUID) error
	IsBlocked(userID, blockedID uuid.UUID) (bool, error)
	List(userID uuid.UUID, page, limit int) ([]model.BlockList, int64, error)
}

type blockListRepository struct {
	db *gorm.DB
}

func NewBlockListRepository(db *gorm.DB) BlockListRepository {
	return &blockListRepository{db: db}
}

func (r *blockListRepository) Block(userID, blockedID uuid.UUID) error {
    bl := &model.BlockList{
        UserID:        userID,
        BlockedUserID: blockedID,
    }
    // 处理软删除导致的唯一键冲突：(user_id, blocked_user_id) 唯一
    // 如果存在被软删除的记录，使用 Upsert 将 deleted_at 置空，实现恢复
    return r.db.Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "user_id"}, {Name: "blocked_user_id"}},
        DoUpdates: clause.Assignments(map[string]any{
            "deleted_at": gorm.Expr("NULL"),
        }),
    }).Create(bl).Error
}

func (r *blockListRepository) Unblock(userID, blockedID uuid.UUID) error {
	return r.db.Where("user_id = ? AND blocked_user_id = ?", userID, blockedID).Delete(&model.BlockList{}).Error
}

func (r *blockListRepository) IsBlocked(userID, blockedID uuid.UUID) (bool, error) {
	var cnt int64
	if err := r.db.Model(&model.BlockList{}).Where("user_id = ? AND blocked_user_id = ?", userID, blockedID).Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (r *blockListRepository) List(userID uuid.UUID, page, limit int) ([]model.BlockList, int64, error) {
	var list []model.BlockList
	var total int64
	q := r.db.Model(&model.BlockList{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
