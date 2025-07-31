package service

import (
	"backend/internal/config"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// FileStorageService 文件存储服务接口
type FileStorageService interface {
	// 上传文件
	UploadFile(ctx context.Context, file *multipart.FileHeader, storageName, category string) (*FileUploadResult, error)
	// 删除文件
	DeleteFile(ctx context.Context, storageName, storagePath string) error
	// 获取文件URL
	GetFileURL(storageName, storagePath string) (string, error)
	// 检查存储是否可用
	IsStorageAvailable(storageName string) bool
	// 获取存储信息
	GetStorageInfo() *StorageInfo
}

// FileUploadResult 文件上传结果
type FileUploadResult struct {
	StoredName  string // 存储文件名
	StoragePath string // 存储路径
	URL         string // 访问URL
	Size        int64  // 文件大小
	MimeType    string // MIME类型
}

// StorageInfo 存储信息
type StorageInfo struct {
	DefaultStorage    string   `json:"default_storage"`
	AvailableStorages []string `json:"available_storages"`
	LocalStorages     []string `json:"local_storages"`
	S3Storages        []string `json:"s3_storages"`
}

// fileStorageService 文件存储服务实现
type fileStorageService struct {
	config *config.FileStorageConfig
}

// NewFileStorageService 创建文件存储服务
func NewFileStorageService(config *config.FileStorageConfig) FileStorageService {
	return &fileStorageService{
		config: config,
	}
}

// UploadFile 上传文件
func (s *fileStorageService) UploadFile(ctx context.Context, file *multipart.FileHeader, storageName, category string) (*FileUploadResult, error) {
	// 如果未指定存储名称，使用默认存储
	if storageName == "" {
		storageName = s.config.DefaultStorage
	}

	// 获取存储配置
	storageConfig, storageType, err := s.config.GetStorageConfig(storageName)
	if err != nil {
		return nil, fmt.Errorf("get storage config error: %w", err)
	}

	// 生成存储文件名
	storedName := s.generateFileName(file.Filename)
	
	// 构建存储路径
	storagePath := s.buildStoragePath(category, storedName)

	// 根据存储类型进行上传
	switch storageType {
	case config.StorageTypeLocal:
		return s.uploadToLocal(file, storageConfig.(*config.LocalStorageConfig), storagePath, storedName)
	case config.StorageTypeS3:
		return s.uploadToS3(ctx, file, storageConfig.(*config.S3StorageConfig), storagePath, storedName)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// DeleteFile 删除文件
func (s *fileStorageService) DeleteFile(ctx context.Context, storageName, storagePath string) error {
	// 获取存储配置
	storageConfig, storageType, err := s.config.GetStorageConfig(storageName)
	if err != nil {
		return fmt.Errorf("get storage config error: %w", err)
	}

	// 根据存储类型进行删除
	switch storageType {
	case config.StorageTypeLocal:
		return s.deleteFromLocal(storageConfig.(*config.LocalStorageConfig), storagePath)
	case config.StorageTypeS3:
		return s.deleteFromS3(ctx, storageConfig.(*config.S3StorageConfig), storagePath)
	default:
		return fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// GetFileURL 获取文件访问URL
func (s *fileStorageService) GetFileURL(storageName, storagePath string) (string, error) {
	// 获取存储配置
	storageConfig, storageType, err := s.config.GetStorageConfig(storageName)
	if err != nil {
		return "", fmt.Errorf("get storage config error: %w", err)
	}

	// 根据存储类型构建URL
	switch storageType {
	case config.StorageTypeLocal:
		localConfig := storageConfig.(*config.LocalStorageConfig)
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(localConfig.BaseURL, "/"), storagePath), nil
	case config.StorageTypeS3:
		s3Config := storageConfig.(*config.S3StorageConfig)
		if s3Config.BaseURL != "" {
			return fmt.Sprintf("%s/%s", strings.TrimSuffix(s3Config.BaseURL, "/"), storagePath), nil
		}
		// 使用默认S3 URL格式
		endpoint := s3Config.Endpoint
		if endpoint == "" {
			endpoint = fmt.Sprintf("https://s3.%s.amazonaws.com", s3Config.Region)
		}
		return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(endpoint, "/"), s3Config.Bucket, storagePath), nil
	default:
		return "", fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// IsStorageAvailable 检查存储是否可用
func (s *fileStorageService) IsStorageAvailable(storageName string) bool {
	_, _, err := s.config.GetStorageConfig(storageName)
	return err == nil
}

// GetStorageInfo 获取存储信息
func (s *fileStorageService) GetStorageInfo() *StorageInfo {
	info := &StorageInfo{
		DefaultStorage:    s.config.DefaultStorage,
		AvailableStorages: s.config.GetAvailableStorages(),
		LocalStorages:     make([]string, 0),
		S3Storages:        make([]string, 0),
	}

	for name := range s.config.Local {
		info.LocalStorages = append(info.LocalStorages, name)
	}

	for name := range s.config.S3 {
		info.S3Storages = append(info.S3Storages, name)
	}

	return info
}

// uploadToLocal 上传文件到本地存储
func (s *fileStorageService) uploadToLocal(file *multipart.FileHeader, config *config.LocalStorageConfig, storagePath, storedName string) (*FileUploadResult, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open uploaded file error: %w", err)
	}
	defer src.Close()

	// 构建完整的文件路径
	fullPath := filepath.Join(config.BasePath, storagePath)
	
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("create directory error: %w", err)
	}

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create destination file error: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	size, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("copy file content error: %w", err)
	}

	// 构建访问URL
	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(config.BaseURL, "/"), storagePath)

	return &FileUploadResult{
		StoredName:  storedName,
		StoragePath: storagePath,
		URL:         url,
		Size:        size,
		MimeType:    file.Header.Get("Content-Type"),
	}, nil
}

// uploadToS3 上传文件到S3存储
func (s *fileStorageService) uploadToS3(ctx context.Context, file *multipart.FileHeader, config *config.S3StorageConfig, storagePath, storedName string) (*FileUploadResult, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open uploaded file error: %w", err)
	}
	defer src.Close()

	// 创建AWS配置
	awsConfig := aws.Config{
		Region:      config.Region,
		Credentials: credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""),
	}

	// 如果有自定义endpoint，设置它
	if config.Endpoint != "" {
		awsConfig.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: config.Endpoint,
			}, nil
		})
	}

	// 创建S3客户端
	s3Client := s3.NewFromConfig(awsConfig)

	// 创建上传管理器
	uploader := manager.NewUploader(s3Client)

	// 执行上传
	result, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(config.Bucket),
		Key:         aws.String(storagePath),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return nil, fmt.Errorf("upload to S3 error: %w", err)
	}

	// 构建访问URL
	var url string
	if config.BaseURL != "" {
		url = fmt.Sprintf("%s/%s", strings.TrimSuffix(config.BaseURL, "/"), storagePath)
	} else {
		url = result.Location
	}

	return &FileUploadResult{
		StoredName:  storedName,
		StoragePath: storagePath,
		URL:         url,
		Size:        file.Size,
		MimeType:    file.Header.Get("Content-Type"),
	}, nil
}

// deleteFromLocal 从本地存储删除文件
func (s *fileStorageService) deleteFromLocal(config *config.LocalStorageConfig, storagePath string) error {
	fullPath := filepath.Join(config.BasePath, storagePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete local file error: %w", err)
	}
	return nil
}

// deleteFromS3 从S3存储删除文件
func (s *fileStorageService) deleteFromS3(ctx context.Context, config *config.S3StorageConfig, storagePath string) error {
	// 创建AWS配置
	awsConfig := aws.Config{
		Region:      config.Region,
		Credentials: credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""),
	}

	// 如果有自定义endpoint，设置它
	if config.Endpoint != "" {
		awsConfig.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: config.Endpoint,
			}, nil
		})
	}

	// 创建S3客户端
	s3Client := s3.NewFromConfig(awsConfig)

	// 删除对象
	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(config.Bucket),
		Key:    aws.String(storagePath),
	})
	if err != nil {
		return fmt.Errorf("delete S3 object error: %w", err)
	}

	return nil
}

// generateFileName 生成存储文件名
func (s *fileStorageService) generateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	return fmt.Sprintf("%s%s", uuid.New().String(), ext)
}

// buildStoragePath 构建存储路径
func (s *fileStorageService) buildStoragePath(category, storedName string) string {
	// 按日期和分类组织文件路径
	dateFolder := time.Now().Format("2006/01/02")
	
	if category != "" {
		return fmt.Sprintf("%s/%s/%s", category, dateFolder, storedName)
	}
	
	return fmt.Sprintf("files/%s/%s", dateFolder, storedName)
} 