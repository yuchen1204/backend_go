package repository

import (
	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FriendshipRepository 好友关系仓储

type FriendshipRepository interface {
	Create(f *model.Friendship) error
	CreateTx(tx *gorm.DB, f *model.Friendship) error
	// CreateIgnoreDuplicate inserts and ignores if (user_id, friend_id) already exists
	CreateIgnoreDuplicate(f *model.Friendship) error
	DeletePairTx(tx *gorm.DB, userID, friendID uuid.UUID) error
	DeletePair(userID, friendID uuid.UUID) error
	List(userID uuid.UUID, search string, page, limit int) ([]model.Friendship, int64, error)
	CountByUser(userID uuid.UUID) (int64, error)
	Exists(userID, friendID uuid.UUID) (bool, error)
	UpdateRemark(userID, friendID uuid.UUID, remark string) error
}

type friendshipRepository struct {
	db *gorm.DB
}

func NewFriendshipRepository(db *gorm.DB) FriendshipRepository {
	return &friendshipRepository{db: db}
}

func (r *friendshipRepository) Create(f *model.Friendship) error {
	return r.db.Create(f).Error
}

func (r *friendshipRepository) CreateTx(tx *gorm.DB, f *model.Friendship) error {
	return tx.Create(f).Error
}

func (r *friendshipRepository) CreateIgnoreDuplicate(f *model.Friendship) error {
    // Upsert：若存在冲突则将 deleted_at 置空，恢复软删除的记录
    return r.db.Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "user_id"}, {Name: "friend_id"}},
        DoUpdates: clause.Assignments(map[string]any{
            "deleted_at": gorm.Expr("NULL"),
        }),
    }).Create(f).Error
}

func (r *friendshipRepository) DeletePairTx(tx *gorm.DB, userID, friendID uuid.UUID) error {
    // 使用硬删除，避免软删除记录占用唯一索引 (user_id, friend_id) 导致无法重新建立好友
    if err := tx.Unscoped().Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&model.Friendship{}).Error; err != nil {
        return err
    }
    if err := tx.Unscoped().Where("user_id = ? AND friend_id = ?", friendID, userID).Delete(&model.Friendship{}).Error; err != nil {
        return err
    }
    return nil
}

func (r *friendshipRepository) DeletePair(userID, friendID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return r.DeletePairTx(tx, userID, friendID)
	})
}

func (r *friendshipRepository) List(userID uuid.UUID, search string, page, limit int) ([]model.Friendship, int64, error) {
	var list []model.Friendship
	var total int64
	q := r.db.Model(&model.Friendship{}).Where("user_id = ?", userID)
	// search 逻辑可以在Service层联表User实现，这里仅做占位
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *friendshipRepository) CountByUser(userID uuid.UUID) (int64, error) {
	var cnt int64
	if err := r.db.Model(&model.Friendship{}).Where("user_id = ?", userID).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r *friendshipRepository) Exists(userID, friendID uuid.UUID) (bool, error) {
	var cnt int64
	if err := r.db.Model(&model.Friendship{}).Where("user_id = ? AND friend_id = ?", userID, friendID).Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (r *friendshipRepository) UpdateRemark(userID, friendID uuid.UUID, remark string) error {
	return r.db.Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Update("remark", remark).Error
}
