package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/response"
	"context"
	"fmt"
	"mime/multipart"

	"github.com/google/uuid"
)

// FileService 文件服务接口
type FileService interface {
	// 上传单个文件
	UploadFile(ctx context.Context, file *multipart.FileHeader, userID *uuid.UUID, req *model.FileUploadRequest) (*model.FileResponse, error)
	// 上传多个文件
	UploadFiles(ctx context.Context, files []*multipart.FileHeader, userID *uuid.UUID, req *model.MultiFileUploadRequest) ([]*model.FileResponse, error)
	// 获取文件详情
	GetFile(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*model.FileResponse, error)
	// 获取用户文件列表
	GetUserFiles(ctx context.Context, userID uuid.UUID, req *model.FileListRequest) (*model.FileListResponse, error)
	// 获取公开文件列表
	GetPublicFiles(ctx context.Context, req *model.FileListRequest) (*model.FileListResponse, error)
	// 更新文件信息
	UpdateFile(ctx context.Context, id uuid.UUID, userID *uuid.UUID, req *model.FileUpdateRequest) (*model.FileResponse, error)
	// 删除文件
	DeleteFile(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error
	// 获取存储信息
	GetStorageInfo(ctx context.Context) (*model.StorageInfoResponse, error)
	// 管理员：获取所有文件列表
	GetAllFiles(ctx context.Context, req *model.FileListRequest) (*model.FileListResponse, error)
	// 管理员：获取任意文件详情
	AdminGetFile(ctx context.Context, id uuid.UUID) (*model.FileResponse, error)
	// 管理员：更新任意文件
	AdminUpdateFile(ctx context.Context, id uuid.UUID, req *model.FileUpdateRequest) (*model.FileResponse, error)
	// 管理员：删除任意文件
	AdminDeleteFile(ctx context.Context, id uuid.UUID) error
}

// fileService 文件服务实现
type fileService struct {
	fileRepo       repository.FileRepository
	fileStorageSvc FileStorageService
}

// NewFileService 创建文件服务
func NewFileService(fileRepo repository.FileRepository, fileStorageSvc FileStorageService) FileService {
	return &fileService{
		fileRepo:       fileRepo,
		fileStorageSvc: fileStorageSvc,
	}
}

// UploadFile 上传单个文件
func (s *fileService) UploadFile(ctx context.Context, file *multipart.FileHeader, userID *uuid.UUID, req *model.FileUploadRequest) (*model.FileResponse, error) {
	// 验证文件
	if err := s.validateFile(file); err != nil {
		return nil, err
	}

	// 上传文件到存储
	result, err := s.fileStorageSvc.UploadFile(ctx, file, req.StorageName, req.Category)
	if err != nil {
		return nil, fmt.Errorf("upload file to storage error: %w", err)
	}

	// 创建文件记录
	fileModel := &model.File{
		OriginalName: file.Filename,
		StoredName:   result.StoredName,
		MimeType:     result.MimeType,
		Size:         result.Size,
		StorageType:  string(s.getStorageType(req.StorageName)),
		StorageName:  s.getStorageName(req.StorageName),
		StoragePath:  result.StoragePath,
		URL:          result.URL,
		UserID:       userID,
		Category:     req.Category,
		Description:  req.Description,
		IsPublic:     s.getBoolValue(req.IsPublic, false),
	}

	if err := s.fileRepo.Create(fileModel); err != nil {
		// 如果数据库操作失败，删除已上传的文件
		s.fileStorageSvc.DeleteFile(ctx, s.getStorageName(req.StorageName), result.StoragePath)
		return nil, fmt.Errorf("create file record error: %w", err)
	}

	return fileModel.ToResponse(), nil
}

// UploadFiles 上传多个文件
func (s *fileService) UploadFiles(ctx context.Context, files []*multipart.FileHeader, userID *uuid.UUID, req *model.MultiFileUploadRequest) ([]*model.FileResponse, error) {
	if len(files) == 0 {
		return nil, response.ErrNoFilesProvided
	}

	var results []*model.FileResponse
	var uploadedFiles []string // 记录已上传的文件路径，用于回滚

	for _, file := range files {
		// 转换请求格式
		uploadReq := &model.FileUploadRequest{
			StorageName: req.StorageName,
			Category:    req.Category,
			Description: req.Description,
			IsPublic:    req.IsPublic,
		}

		result, err := s.UploadFile(ctx, file, userID, uploadReq)
		if err != nil {
			// 上传失败，回滚已上传的文件
			s.rollbackUploadedFiles(ctx, uploadedFiles)
			return nil, fmt.Errorf("upload file %s error: %w", file.Filename, err)
		}

		results = append(results, result)
		uploadedFiles = append(uploadedFiles, result.StorageName+":"+result.StorageName) // 用于回滚的标识
	}

	return results, nil
}

// GetFile 获取文件详情
func (s *fileService) GetFile(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*model.FileResponse, error) {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if !file.IsPublic && (userID == nil || file.UserID == nil || *file.UserID != *userID) {
		return nil, response.ErrFileAccessDenied
	}

	return file.ToResponse(), nil
}

// GetUserFiles 获取用户文件列表
func (s *fileService) GetUserFiles(ctx context.Context, userID uuid.UUID, req *model.FileListRequest) (*model.FileListResponse, error) {
	return s.fileRepo.GetByUserID(userID, req)
}

// GetPublicFiles 获取公开文件列表
func (s *fileService) GetPublicFiles(ctx context.Context, req *model.FileListRequest) (*model.FileListResponse, error) {
	return s.fileRepo.GetPublicFiles(req)
}

// UpdateFile 更新文件信息
func (s *fileService) UpdateFile(ctx context.Context, id uuid.UUID, userID *uuid.UUID, req *model.FileUpdateRequest) (*model.FileResponse, error) {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 检查权限（只有文件所有者可以更新）
	if userID == nil || file.UserID == nil || *file.UserID != *userID {
		return nil, response.ErrFileAccessDenied
	}

	// 更新字段
	if req.Category != "" {
		file.Category = req.Category
	}
	if req.Description != "" {
		file.Description = req.Description
	}
	if req.IsPublic != nil {
		file.IsPublic = *req.IsPublic
	}

	if err := s.fileRepo.Update(file); err != nil {
		return nil, err
	}

	return file.ToResponse(), nil
}

// DeleteFile 删除文件
func (s *fileService) DeleteFile(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 检查权限（只有文件所有者可以删除）
	if userID == nil || file.UserID == nil || *file.UserID != *userID {
		return response.ErrFileAccessDenied
	}

	// 从存储中删除文件
	if err := s.fileStorageSvc.DeleteFile(ctx, file.StorageName, file.StoragePath); err != nil {
		// 记录日志但不阻断删除流程
		fmt.Printf("Warning: failed to delete file from storage: %v\n", err)
	}

	// 从数据库中删除记录
	return s.fileRepo.Delete(id)
}

// GetAllFiles 管理员：获取所有文件列表
func (s *fileService) GetAllFiles(ctx context.Context, req *model.FileListRequest) (*model.FileListResponse, error) {
	return s.fileRepo.GetAllFiles(req)
}

// AdminGetFile 管理员：获取任意文件详情（不做权限校验）
func (s *fileService) AdminGetFile(ctx context.Context, id uuid.UUID) (*model.FileResponse, error) {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return file.ToResponse(), nil
}

// AdminUpdateFile 管理员：更新任意文件
func (s *fileService) AdminUpdateFile(ctx context.Context, id uuid.UUID, req *model.FileUpdateRequest) (*model.FileResponse, error) {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Category != "" {
		file.Category = req.Category
	}
	if req.Description != "" {
		file.Description = req.Description
	}
	if req.IsPublic != nil {
		file.IsPublic = *req.IsPublic
	}

	if err := s.fileRepo.Update(file); err != nil {
		return nil, err
	}

	return file.ToResponse(), nil
}

// AdminDeleteFile 管理员：删除任意文件
func (s *fileService) AdminDeleteFile(ctx context.Context, id uuid.UUID) error {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 先尝试从存储删除
	if err := s.fileStorageSvc.DeleteFile(ctx, file.StorageName, file.StoragePath); err != nil {
		// 记录警告，不阻断删除
		fmt.Printf("Warning: admin failed to delete file from storage: %v\n", err)
	}

	return s.fileRepo.Delete(id)
}

// GetStorageInfo 获取存储信息
func (s *fileService) GetStorageInfo(ctx context.Context) (*model .StorageInfoResponse, error) {
	info := s.fileStorageSvc.GetStorageInfo()
	
	var localStorages, s3Storages []string
	for _, storage := range info.LocalStorages {
		localStorages = append(localStorages, storage)
	}
	for _, storage := range info.S3Storages {
		s3Storages = append(s3Storages, storage)
	}

	return &model.StorageInfoResponse{
		DefaultStorage:    info.DefaultStorage,
		AvailableStorages: info.AvailableStorages,
		LocalStorages:     localStorages,
		S3Storages:        s3Storages,
	}, nil
}

// validateFile 验证文件
func (s *fileService) validateFile(file *multipart.FileHeader) error {
	// 文件大小限制（默认10MB）
	const maxFileSize = 10 * 1024 * 1024
	if file.Size > maxFileSize {
		return response.ErrFileTooLarge
	}

	// 文件名不能为空
	if file.Filename == "" {
		return response.ErrInvalidFileName
	}

	return nil
}

// getStorageType 根据存储名称获取存储类型
func (s *fileService) getStorageType(storageName string) string {
	if storageName == "" {
		storageName = s.fileStorageSvc.GetStorageInfo().DefaultStorage
	}
	// 这里可以通过配置获取实际的存储类型
	// 为简单起见，先返回默认值
	return "local"
}

// getStorageName 获取实际的存储名称
func (s *fileService) getStorageName(storageName string) string {
	if storageName == "" {
		return s.fileStorageSvc.GetStorageInfo().DefaultStorage
	}
	return storageName
}

// getBoolValue 获取布尔值，如果为nil则返回默认值
func (s *fileService) getBoolValue(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}

// rollbackUploadedFiles 回滚已上传的文件
func (s *fileService) rollbackUploadedFiles(_ context.Context, uploadedFiles []string) {
	for _, fileInfo := range uploadedFiles {
		// 解析存储名称和路径
		// 这里可以根据实际需要实现更复杂的回滚逻辑
		fmt.Printf("Rolling back uploaded file: %s\n", fileInfo)
	}
}