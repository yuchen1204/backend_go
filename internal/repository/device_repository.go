package repository

import (
	"backend/internal/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DeviceRepository 设备与设备验证仓储接口
// 负责 UserDevice 与 DeviceVerification 的持久化
 type DeviceRepository interface {
	// ---- UserDevice ----
	GetDeviceByUserAndFingerprint(userID uuid.UUID, deviceID string) (*model.UserDevice, error)
	CreateDevice(device *model.UserDevice) error
	UpdateDevice(device *model.UserDevice) error
	ListDevicesByUser(userID uuid.UUID) ([]*model.UserDevice, error)
	DeleteDevice(id uuid.UUID) error

	// ---- DeviceVerification ----
	CreateVerification(v *model.DeviceVerification) error
	GetLatestPendingVerification(userID uuid.UUID, deviceID string) (*model.DeviceVerification, error)
	MarkVerificationVerified(id uuid.UUID) error
	IncrementVerificationAttempt(id uuid.UUID) error
	DeleteExpiredVerifications(now time.Time) error
}

// deviceRepository 实现
 type deviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository 创建仓储实例
func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

// GetDeviceByUserAndFingerprint 查询用户的特定设备
func (r *deviceRepository) GetDeviceByUserAndFingerprint(userID uuid.UUID, deviceID string) (*model.UserDevice, error) {
	var d model.UserDevice
	if err := r.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

// CreateDevice 创建设备
func (r *deviceRepository) CreateDevice(device *model.UserDevice) error {
    // 处理软删除与唯一索引 (user_id, device_id) 冲突：
    // 如果已存在软删除记录，则通过 Upsert 将 deleted_at 置空并更新最新字段，实现“恢复或创建”。
    if err := r.db.Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "user_id"}, {Name: "device_id"}},
        DoUpdates: clause.Assignments(map[string]any{
            "deleted_at":  gorm.Expr("NULL"),
            "device_name": device.DeviceName,
            "device_type": device.DeviceType,
            "user_agent":  device.UserAgent,
            "ip_address":  device.IPAddress,
            "is_trusted":  device.IsTrusted,
            "last_login_at": device.LastLoginAt,
            "updated_at":   time.Now(),
        }),
    }).Create(device).Error; err != nil {
        return fmt.Errorf("create device error: %w", err)
    }
    return nil
}

// UpdateDevice 更新设备
func (r *deviceRepository) UpdateDevice(device *model.UserDevice) error {
	if err := r.db.Save(device).Error; err != nil {
		return fmt.Errorf("update device error: %w", err)
	}
	return nil
}

// ListDevicesByUser 列出用户的所有设备
func (r *deviceRepository) ListDevicesByUser(userID uuid.UUID) ([]*model.UserDevice, error) {
	var list []*model.UserDevice
	if err := r.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list devices error: %w", err)
	}
	return list, nil
}

// DeleteDevice 删除设备（软删除）
func (r *deviceRepository) DeleteDevice(id uuid.UUID) error {
	if err := r.db.Delete(&model.UserDevice{}, id).Error; err != nil {
		return fmt.Errorf("delete device error: %w", err)
	}
	return nil
}

// CreateVerification 创建设备验证记录
func (r *deviceRepository) CreateVerification(v *model.DeviceVerification) error {
	if err := r.db.Create(v).Error; err != nil {
		return fmt.Errorf("create device verification error: %w", err)
	}
	return nil
}

// GetLatestPendingVerification 获取最近一条未验证且未过期的记录
func (r *deviceRepository) GetLatestPendingVerification(userID uuid.UUID, deviceID string) (*model.DeviceVerification, error) {
	var v model.DeviceVerification
	now := time.Now()
	if err := r.db.Where("user_id = ? AND device_id = ? AND is_verified = ? AND expires_at > ?", userID, deviceID, false, now).
		Order("created_at DESC").First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

// MarkVerificationVerified 标记验证通过
func (r *deviceRepository) MarkVerificationVerified(id uuid.UUID) error {
	now := time.Now()
	updates := map[string]interface{}{
		"is_verified": true,
		"verified_at": now,
	}
	if err := r.db.Model(&model.DeviceVerification{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("mark verification verified error: %w", err)
	}
	return nil
}

// IncrementVerificationAttempt 增加尝试次数
func (r *deviceRepository) IncrementVerificationAttempt(id uuid.UUID) error {
	if err := r.db.Model(&model.DeviceVerification{}).
		Where("id = ?", id).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1")).Error; err != nil {
		return fmt.Errorf("increment verification attempt error: %w", err)
	}
	return nil
}

// DeleteExpiredVerifications 删除已过期的验证记录
func (r *deviceRepository) DeleteExpiredVerifications(now time.Time) error {
	if err := r.db.Where("expires_at <= ?", now).Delete(&model.DeviceVerification{}).Error; err != nil {
		return fmt.Errorf("delete expired verifications error: %w", err)
	}
	return nil
}

