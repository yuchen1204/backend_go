package repository

import (
	"backend/internal/model"
	"backend/internal/response"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileRepository 文件仓储接口
type FileRepository interface {
	// 创建文件记录
	Create(file *model.File) error
	// 根据ID获取文件
	GetByID(id uuid.UUID) (*model.File, error)
	// 根据用户ID获取文件列表
	GetByUserID(userID uuid.UUID, req *model.FileListRequest) (*model.FileListResponse, error)
	// 获取公开文件列表
	GetPublicFiles(req *model.FileListRequest) (*model.FileListResponse, error)
	// 获取所有文件列表（管理员用）
	GetAllFiles(req *model.FileListRequest) (*model.FileListResponse, error)
	// 更新文件信息
	Update(file *model.File) error
	// 软删除文件
	Delete(id uuid.UUID) error
	// 物理删除文件
	HardDelete(id uuid.UUID) error
	// 根据存储路径获取文件
	GetByStoragePath(storageName, storagePath string) (*model.File, error)
}

// fileRepository 文件仓储实现
type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建文件仓储
func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{
		db: db,
	}
}

// Create 创建文件记录
func (r *fileRepository) Create(file *model.File) error {
	if err := r.db.Create(file).Error; err != nil {
		return fmt.Errorf("create file record error: %w", err)
	}
	return nil
}

// GetByID 根据ID获取文件
func (r *fileRepository) GetByID(id uuid.UUID) (*model.File, error) {
	var file model.File
	if err := r.db.Where("id = ?", id).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.ErrFileNotFound
		}
		return nil, fmt.Errorf("get file by id error: %w", err)
	}
	return &file, nil
}

// GetByUserID 根据用户ID获取文件列表
func (r *fileRepository) GetByUserID(userID uuid.UUID, req *model.FileListRequest) (*model.FileListResponse, error) {
	return r.getFileList(req, func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	})
}

// GetPublicFiles 获取公开文件列表
func (r *fileRepository) GetPublicFiles(req *model.FileListRequest) (*model.FileListResponse, error) {
	return r.getFileList(req, func(db *gorm.DB) *gorm.DB {
		return db.Where("is_public = ?", true)
	})
}

// GetAllFiles 获取所有文件列表（管理员用）
func (r *fileRepository) GetAllFiles(req *model.FileListRequest) (*model.FileListResponse, error) {
	return r.getFileList(req, nil)
}

// getFileList 获取文件列表的通用方法
func (r *fileRepository) getFileList(req *model.FileListRequest, scopeFunc func(*gorm.DB) *gorm.DB) (*model.FileListResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 构建查询
	query := r.db.Model(&model.File{})

	// 应用作用域函数
	if scopeFunc != nil {
		query = scopeFunc(query)
	}

	// 添加筛选条件
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.StorageType != "" {
		query = query.Where("storage_type = ?", req.StorageType)
	}
	if req.StorageName != "" {
		query = query.Where("storage_name = ?", req.StorageName)
	}
	if req.IsPublic != nil {
		query = query.Where("is_public = ?", *req.IsPublic)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count files error: %w", err)
	}

	// 获取文件列表
	var files []*model.File
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&files).Error; err != nil {
		return nil, fmt.Errorf("get files error: %w", err)
	}

	// 转换为响应格式
	fileResponses := make([]*model.FileResponse, len(files))
	for i, file := range files {
		fileResponses[i] = file.ToResponse()
	}

	// 计算总页数
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &model.FileListResponse{
		Files:      fileResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Update 更新文件信息
func (r *fileRepository) Update(file *model.File) error {
	if err := r.db.Save(file).Error; err != nil {
		return fmt.Errorf("update file error: %w", err)
	}
	return nil
}

// Delete 软删除文件
func (r *fileRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&model.File{}, id).Error; err != nil {
		return fmt.Errorf("delete file error: %w", err)
	}
	return nil
}

// HardDelete 物理删除文件
func (r *fileRepository) HardDelete(id uuid.UUID) error {
	if err := r.db.Unscoped().Delete(&model.File{}, id).Error; err != nil {
		return fmt.Errorf("hard delete file error: %w", err)
	}
	return nil
}

// GetByStoragePath 根据存储路径获取文件
func (r *fileRepository) GetByStoragePath(storageName, storagePath string) (*model.File, error) {
	var file model.File
	if err := r.db.Where("storage_name = ? AND storage_path = ?", storageName, storagePath).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.ErrFileNotFound
		}
		return nil, fmt.Errorf("get file by storage path error: %w", err)
	}
	return &file, nil
} 