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
	// UpdateProfile 更新用户基本信息（昵称、简介、头像、背景图）
	UpdateProfile(userID uuid.UUID, nickname, bio, avatar, backgroundURL string) error
	// UpdatePassword 更新用户密码
	UpdatePassword(userID uuid.UUID, passwordSalt string) error
	// UpdateLastLoginAt 更新用户最后登录时间
	UpdateLastLoginAt(userID uuid.UUID, t time.Time) error
	// Delete 删除用户
	Delete(id uuid.UUID) error
	// 管理员专用方法
	// GetUsersWithPagination 分页获取用户列表
	GetUsersWithPagination(page, limit int, search string) ([]*model.User, int64, error)
	// GetByUintID 根据uint类型ID获取用户（管理员用）
	GetByUintID(id uint) (*model.User, error)
	// UpdateStatus 更新用户状态
	UpdateStatus(id uint, status string) error
	// DeleteByUintID 根据uint类型ID删除用户（管理员用）
	DeleteByUintID(id uint) error
	// GetUserStats 获取用户统计信息
	GetUserStats() (map[string]interface{}, error)
	// UpdateStatusByUUID 根据UUID更新用户状态
	UpdateStatusByUUID(id uuid.UUID, status string) error
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

// UpdateProfile 更新用户基本信息（昵称、简介、头像、背景图）
func (r *userRepository) UpdateProfile(userID uuid.UUID, nickname, bio, avatar, backgroundURL string) error {
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
	if backgroundURL != "" {
		updates["background_url"] = backgroundURL
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

// UpdateLastLoginAt 更新用户最后登录时间
func (r *userRepository) UpdateLastLoginAt(userID uuid.UUID, t time.Time) error {
	updates := map[string]interface{}{
		"last_login_at": t,
		"updated_at":    time.Now(),
	}
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.User{}).Error
}

// GetUsersWithPagination 分页获取用户列表
func (r *userRepository) GetUsersWithPagination(page, limit int, search string) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64
	
	query := r.db.Model(&model.User{})
	
	// 如果有搜索条件，添加搜索
	if search != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ? OR nickname ILIKE ?", 
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

// GetByUintID 根据uint类型ID获取用户（管理员用）
func (r *userRepository) GetByUintID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateStatus 更新用户状态
func (r *userRepository) UpdateStatus(id uint, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteByUintID 根据uint类型ID删除用户（管理员用）
func (r *userRepository) DeleteByUintID(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.User{}).Error
}

// GetUserStats 获取用户统计信息
func (r *userRepository) GetUserStats() (map[string]interface{}, error) {
	var totalUsers int64
	var activeUsers int64
	var inactiveUsers int64
	var bannedUsers int64
	
	// 总用户数
	if err := r.db.Model(&model.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}
	
	// 活跃用户数
	if err := r.db.Model(&model.User{}).Where("status = ?", "active").Count(&activeUsers).Error; err != nil {
		return nil, err
	}
	
	// 非活跃用户数
	if err := r.db.Model(&model.User{}).Where("status = ?", "inactive").Count(&inactiveUsers).Error; err != nil {
		return nil, err
	}
	
	// 被封禁用户数
	if err := r.db.Model(&model.User{}).Where("status = ?", "banned").Count(&bannedUsers).Error; err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_users":    totalUsers,
		"active_users":   activeUsers,
		"inactive_users": inactiveUsers,
		"banned_users":   bannedUsers,
	}, nil
}

// UpdateStatusByUUID 根据UUID更新用户状态
func (r *userRepository) UpdateStatusByUUID(id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
} 