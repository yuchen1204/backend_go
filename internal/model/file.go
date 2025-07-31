package model

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// File 文件模型
type File struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OriginalName string   `json:"original_name" gorm:"not null;size:255"`        // 原始文件名
	StoredName   string   `json:"stored_name" gorm:"not null;size:255"`          // 存储文件名
	MimeType     string   `json:"mime_type" gorm:"not null;size:100"`            // MIME类型
	Size         int64    `json:"size" gorm:"not null"`                          // 文件大小（字节）
	StorageType  string   `json:"storage_type" gorm:"not null;size:50"`          // 存储类型（local/s3）
	StorageName  string   `json:"storage_name" gorm:"not null;size:100"`         // 存储名称
	StoragePath  string   `json:"storage_path" gorm:"not null;size:500"`         // 存储路径
	URL          string   `json:"url" gorm:"not null;size:500"`                  // 访问URL
	UserID       *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`              // 上传用户ID（可选）
	Category     string   `json:"category" gorm:"size:50;index"`                 // 文件分类（avatar, document, image等）
	Description  string   `json:"description" gorm:"size:500"`                   // 文件描述
	IsPublic     bool     `json:"is_public" gorm:"default:false"`                // 是否公开访问
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}

// FileUploadRequest 文件上传请求
type FileUploadRequest struct {
	StorageName string `form:"storage_name" binding:"omitempty"`                    // 存储名称（可选，使用默认存储）
	Category    string `form:"category" binding:"omitempty,max=50"`                 // 文件分类
	Description string `form:"description" binding:"omitempty,max=500"`             // 文件描述
	IsPublic    *bool  `form:"is_public" binding:"omitempty"`                       // 是否公开（可选）
}

// MultiFileUploadRequest 多文件上传请求
type MultiFileUploadRequest struct {
	StorageName string `form:"storage_name" binding:"omitempty"`                    // 存储名称（可选，使用默认存储）
	Category    string `form:"category" binding:"omitempty,max=50"`                 // 文件分类
	Description string `form:"description" binding:"omitempty,max=500"`             // 文件描述
	IsPublic    *bool  `form:"is_public" binding:"omitempty"`                       // 是否公开（可选）
}

// FileResponse 文件响应结构
type FileResponse struct {
	ID           uuid.UUID  `json:"id"`
	OriginalName string     `json:"original_name"`
	MimeType     string     `json:"mime_type"`
	Size         int64      `json:"size"`
	StorageType  string     `json:"storage_type"`
	StorageName  string     `json:"storage_name"`
	URL          string     `json:"url"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	Category     string     `json:"category"`
	Description  string     `json:"description"`
	IsPublic     bool       `json:"is_public"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ToResponse 将File转换为FileResponse
func (f *File) ToResponse() *FileResponse {
	return &FileResponse{
		ID:           f.ID,
		OriginalName: f.OriginalName,
		MimeType:     f.MimeType,
		Size:         f.Size,
		StorageType:  f.StorageType,
		StorageName:  f.StorageName,
		URL:          f.URL,
		UserID:       f.UserID,
		Category:     f.Category,
		Description:  f.Description,
		IsPublic:     f.IsPublic,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}

// FileListRequest 文件列表请求
type FileListRequest struct {
	Category    string `form:"category" binding:"omitempty"`     // 按分类筛选
	StorageType string `form:"storage_type" binding:"omitempty"` // 按存储类型筛选
	StorageName string `form:"storage_name" binding:"omitempty"` // 按存储名称筛选
	IsPublic    *bool  `form:"is_public" binding:"omitempty"`    // 按公开状态筛选
	Page        int    `form:"page" binding:"omitempty,min=1"`   // 页码
	PageSize    int    `form:"page_size" binding:"omitempty,min=1,max=100"` // 每页数量
}

// FileListResponse 文件列表响应
type FileListResponse struct {
	Files      []*FileResponse `json:"files"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// FileUpdateRequest 文件更新请求
type FileUpdateRequest struct {
	Category    string `json:"category" binding:"omitempty,max=50"`     // 文件分类
	Description string `json:"description" binding:"omitempty,max=500"` // 文件描述
	IsPublic    *bool  `json:"is_public" binding:"omitempty"`           // 是否公开
}

// StorageInfoResponse 存储信息响应
type StorageInfoResponse struct {
	DefaultStorage      string   `json:"default_storage"`
	AvailableStorages   []string `json:"available_storages"`
	LocalStorages       []string `json:"local_storages"`
	S3Storages          []string `json:"s3_storages"`
} 