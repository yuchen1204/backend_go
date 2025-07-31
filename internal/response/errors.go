package response

import "errors"

// 文件相关错误定义
var (
	ErrFileNotFound       = errors.New("文件不存在")
	ErrFileAccessDenied   = errors.New("文件访问被拒绝")
	ErrFileTooLarge       = errors.New("文件大小超过限制")
	ErrInvalidFileName    = errors.New("无效的文件名")
	ErrNoFilesProvided    = errors.New("未提供文件")
	ErrInvalidFileType    = errors.New("不支持的文件类型")
	ErrStorageNotFound    = errors.New("存储配置不存在")
	ErrUploadFailed       = errors.New("文件上传失败")
	ErrDeleteFailed       = errors.New("文件删除失败")
	ErrInvalidStorageName = errors.New("无效的存储名称")
) 