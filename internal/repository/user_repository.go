package repository

import (
	"backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Create 创建用户
	Create(user *model.User) error
	// GetByID 根据ID获取用户
	GetByID(id uuid.UUID) (*model.User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(username string) (*model.User, error)
	// GetByEmail 根据邮箱获取用户
	GetByEmail(email string) (*model.User, error)
	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(username string) (bool, error)
	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(email string) (bool, error)
	// Update 更新用户信息
	Update(user *model.User) error
	// UpdateProfile 更新用户基本信息（昵称、简介、头像）
	UpdateProfile(userID uuid.UUID, nickname, bio, avatar string) error
	// UpdatePassword 更新用户密码
	UpdatePassword(userID uuid.UUID, passwordSalt string) error
	// Delete 删除用户
	Delete(id uuid.UUID) error
}

// userRepository 用户仓储实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// Update 更新用户信息
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateProfile 更新用户基本信息（昵称、简介、头像）
func (r *userRepository) UpdateProfile(userID uuid.UUID, nickname, bio, avatar string) error {
	updates := make(map[string]interface{})
	
	// 只更新非空字段
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if bio != "" {
		updates["bio"] = bio
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	
	// 如果没有要更新的字段，直接返回
	if len(updates) == 0 {
		return nil
	}
	
	// 更新 updated_at 字段
	updates["updated_at"] = time.Now()
	
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// UpdatePassword 更新用户密码
func (r *userRepository) UpdatePassword(userID uuid.UUID, passwordSalt string) error {
	updates := map[string]interface{}{
		"password_salt": passwordSalt,
		"updated_at":    time.Now(),
	}
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.User{}).Error
} 