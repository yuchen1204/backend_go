package config

import (
	"fmt"
	"strings"
)

// FileStorageType 文件存储类型
type FileStorageType string

const (
	StorageTypeLocal FileStorageType = "local"
	StorageTypeS3    FileStorageType = "s3"
)

// LocalStorageConfig 本地存储配置
type LocalStorageConfig struct {
	BasePath string // 存储基础路径
	BaseURL  string // 访问基础URL
}

// S3StorageConfig S3存储配置
type S3StorageConfig struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string // 可选，用于兼容其他S3协议的服务
	BaseURL         string // 可选，自定义访问基础URL
}

// FileStorageConfig 文件存储配置
type FileStorageConfig struct {
	DefaultStorage string                         // 默认存储类型
	Local          map[string]*LocalStorageConfig // 本地存储配置（支持多个）
	S3             map[string]*S3StorageConfig    // S3存储配置（支持多个）
}

// GetFileStorageConfig 获取文件存储配置
func GetFileStorageConfig() *FileStorageConfig {
	config := &FileStorageConfig{
		DefaultStorage: getEnv("FILE_STORAGE_DEFAULT", "local_default"),
		Local:          make(map[string]*LocalStorageConfig),
		S3:             make(map[string]*S3StorageConfig),
	}

	// 解析本地存储配置
	config.parseLocalStorageConfigs()
	
	// 解析S3存储配置
	config.parseS3StorageConfigs()

	return config
}

// parseLocalStorageConfigs 解析本地存储配置
func (c *FileStorageConfig) parseLocalStorageConfigs() {
	// 支持配置格式：FILE_STORAGE_LOCAL_NAMES=default,avatar,document
	localNames := getEnv("FILE_STORAGE_LOCAL_NAMES", "default")
	if localNames != "" {
		names := strings.Split(localNames, ",")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name != "" {
				config := &LocalStorageConfig{
					BasePath: getEnv(fmt.Sprintf("FILE_STORAGE_LOCAL_%s_PATH", strings.ToUpper(name)), "./uploads/"+name),
					BaseURL:  getEnv(fmt.Sprintf("FILE_STORAGE_LOCAL_%s_URL", strings.ToUpper(name)), "http://localhost:8080/uploads/"+name),
				}
				c.Local[name] = config
			}
		}
	}
}

// parseS3StorageConfigs 解析S3存储配置
func (c *FileStorageConfig) parseS3StorageConfigs() {
	// 支持配置格式：FILE_STORAGE_S3_NAMES=main,backup,cdn
	s3Names := getEnv("FILE_STORAGE_S3_NAMES", "")
	if s3Names != "" {
		names := strings.Split(s3Names, ",")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name != "" {
				config := &S3StorageConfig{
					Region:          getEnv(fmt.Sprintf("FILE_STORAGE_S3_%s_REGION", strings.ToUpper(name)), "us-east-1"),
					Bucket:          getEnv(fmt.Sprintf("FILE_STORAGE_S3_%s_BUCKET", strings.ToUpper(name)), ""),
					AccessKeyID:     getEnv(fmt.Sprintf("FILE_STORAGE_S3_%s_ACCESS_KEY", strings.ToUpper(name)), ""),
					SecretAccessKey: getEnv(fmt.Sprintf("FILE_STORAGE_S3_%s_SECRET_KEY", strings.ToUpper(name)), ""),
					Endpoint:        getEnv(fmt.Sprintf("FILE_STORAGE_S3_%s_ENDPOINT", strings.ToUpper(name)), ""),
					BaseURL:         getEnv(fmt.Sprintf("FILE_STORAGE_S3_%s_BASE_URL", strings.ToUpper(name)), ""),
				}
				c.S3[name] = config
			}
		}
	}
}

// GetStorageConfig 根据存储名称获取存储配置
func (c *FileStorageConfig) GetStorageConfig(storageName string) (interface{}, FileStorageType, error) {
	// 检查本地存储
	if localConfig, exists := c.Local[storageName]; exists {
		return localConfig, StorageTypeLocal, nil
	}
	
	// 检查S3存储
	if s3Config, exists := c.S3[storageName]; exists {
		return s3Config, StorageTypeS3, nil
	}
	
	return nil, "", fmt.Errorf("storage config not found: %s", storageName)
}

// ValidateConfigs 验证配置有效性
func (c *FileStorageConfig) ValidateConfigs() error {
	// 检查默认存储是否存在
	_, _, err := c.GetStorageConfig(c.DefaultStorage)
	if err != nil {
		return fmt.Errorf("default storage config error: %w", err)
	}
	
	// 验证S3配置
	for name, s3Config := range c.S3 {
		if s3Config.Bucket == "" {
			return fmt.Errorf("S3 storage '%s': bucket is required", name)
		}
		if s3Config.AccessKeyID == "" {
			return fmt.Errorf("S3 storage '%s': access key is required", name)
		}
		if s3Config.SecretAccessKey == "" {
			return fmt.Errorf("S3 storage '%s': secret key is required", name)
		}
	}
	
	return nil
}

// GetAvailableStorages 获取所有可用的存储名称
func (c *FileStorageConfig) GetAvailableStorages() []string {
	var storages []string
	
	for name := range c.Local {
		storages = append(storages, name)
	}
	
	for name := range c.S3 {
		storages = append(storages, name)
	}
	
	return storages
} 