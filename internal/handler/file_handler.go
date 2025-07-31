package handler

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/response"
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FileHandler 文件处理器
type FileHandler struct {
	fileService service.FileService
}

// NewFileHandler 创建文件处理器
func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// UploadFile 上传单个文件
// @Summary 上传单个文件
// @Description 上传单个文件到指定的存储位置
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "要上传的文件"
// @Param storage_name formData string false "存储名称（可选，默认使用系统默认存储）"
// @Param category formData string false "文件分类"
// @Param description formData string false "文件描述"
// @Param is_public formData boolean false "是否公开访问"
// @Success 201 {object} response.ResponseData{data=model.FileResponse} "上传成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 401 {object} response.ResponseData "未授权"
// @Failure 413 {object} response.ResponseData "文件过大"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /files/upload [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	// 获取用户ID（从JWT中间件设置）
	payload, exists := c.Get(middleware.AuthorizationPayloadKey)
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "无法获取授权信息", nil)
		return
	}

	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息格式错误", nil)
		return
	}

	userID := claims.UserID

	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "获取文件失败", err.Error())
		return
	}

	// 绑定表单数据
	var req model.FileUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "参数绑定失败", err.Error())
		return
	}

	// 上传文件
	result, err := h.fileService.UploadFile(c.Request.Context(), fileHeader, &userID, &req)
	if err != nil {
		if err == response.ErrFileTooLarge {
			response.ErrorResponse(c, http.StatusRequestEntityTooLarge, "文件过大", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "文件上传失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "文件上传成功", result)
}

// UploadFiles 上传多个文件
// @Summary 上传多个文件
// @Description 批量上传多个文件到指定的存储位置
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param files formData []file true "要上传的文件列表"
// @Param storage_name formData string false "存储名称（可选，默认使用系统默认存储）"
// @Param category formData string false "文件分类"
// @Param description formData string false "文件描述"
// @Param is_public formData boolean false "是否公开访问"
// @Success 201 {object} response.ResponseData{data=[]model.FileResponse} "上传成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 401 {object} response.ResponseData "未授权"
// @Failure 413 {object} response.ResponseData "文件过大"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /files/upload-multiple [post]
func (h *FileHandler) UploadFiles(c *gin.Context) {
	// 获取用户ID
	payload, exists := c.Get(middleware.AuthorizationPayloadKey)
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "无法获取授权信息", nil)
		return
	}

	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息格式错误", nil)
		return
	}

	userID := claims.UserID

	// 获取上传的文件列表
	form, err := c.MultipartForm()
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "获取文件列表失败", err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.ErrorResponse(c, http.StatusBadRequest, "未提供文件", nil)
		return
	}

	// 绑定表单数据
	var req model.MultiFileUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "参数绑定失败", err.Error())
		return
	}

	// 上传文件
	results, err := h.fileService.UploadFiles(c.Request.Context(), files, &userID, &req)
	if err != nil {
		if err == response.ErrFileTooLarge {
			response.ErrorResponse(c, http.StatusRequestEntityTooLarge, "文件过大", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "文件上传失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "文件上传成功", results)
}

// GetFile 获取文件详情
// @Summary 获取文件详情
// @Description 根据文件ID获取文件详细信息
// @Tags files
// @Produce json
// @Param id path string true "文件ID"
// @Success 200 {object} response.ResponseData{data=model.FileResponse} "获取成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 404 {object} response.ResponseData "文件不存在"
// @Failure 403 {object} response.ResponseData "访问被拒绝"
// @Router /files/{id} [get]
func (h *FileHandler) GetFile(c *gin.Context) {
	// 获取文件ID
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的文件ID", nil)
		return
	}

	// 获取用户ID（可选）
	var userID *uuid.UUID
	if payload, exists := c.Get(middleware.AuthorizationPayloadKey); exists {
		if claims, ok := payload.(*service.JWTClaims); ok {
			userID = &claims.UserID
		}
	}

	// 获取文件
	file, err := h.fileService.GetFile(c.Request.Context(), id, userID)
	if err != nil {
		if err == response.ErrFileNotFound {
			response.ErrorResponse(c, http.StatusNotFound, "文件不存在", nil)
			return
		}
		if err == response.ErrFileAccessDenied {
			response.ErrorResponse(c, http.StatusForbidden, "访问被拒绝", nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "获取文件失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取成功", file)
}

// GetUserFiles 获取用户文件列表
// @Summary 获取当前用户的文件列表
// @Description 获取当前登录用户的文件列表，支持分页和筛选
// @Tags files
// @Produce json
// @Security ApiKeyAuth
// @Param category query string false "文件分类筛选"
// @Param storage_type query string false "存储类型筛选"
// @Param storage_name query string false "存储名称筛选"
// @Param is_public query boolean false "是否公开筛选"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} response.ResponseData{data=model.FileListResponse} "获取成功"
// @Failure 401 {object} response.ResponseData "未授权"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /files/my [get]
func (h *FileHandler) GetUserFiles(c *gin.Context) {
	// 获取用户ID
	payload, exists := c.Get(middleware.AuthorizationPayloadKey)
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "无法获取授权信息", nil)
		return
	}

	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息格式错误", nil)
		return
	}

	userID := claims.UserID

	// 绑定查询参数
	var req model.FileListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "参数绑定失败", err.Error())
		return
	}

	// 获取文件列表
	files, err := h.fileService.GetUserFiles(c.Request.Context(), userID, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取文件列表失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取成功", files)
}

// GetPublicFiles 获取公开文件列表
// @Summary 获取公开文件列表
// @Description 获取所有公开访问的文件列表，支持分页和筛选
// @Tags files
// @Produce json
// @Param category query string false "文件分类筛选"
// @Param storage_type query string false "存储类型筛选"
// @Param storage_name query string false "存储名称筛选"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} response.ResponseData{data=model.FileListResponse} "获取成功"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /files/public [get]
func (h *FileHandler) GetPublicFiles(c *gin.Context) {
	// 绑定查询参数
	var req model.FileListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "参数绑定失败", err.Error())
		return
	}

	// 获取公开文件列表
	files, err := h.fileService.GetPublicFiles(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取文件列表失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取成功", files)
}

// UpdateFile 更新文件信息
// @Summary 更新文件信息
// @Description 更新文件的分类、描述等信息（仅文件所有者可操作）
// @Tags files
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "文件ID"
// @Param request body model.FileUpdateRequest true "更新请求"
// @Success 200 {object} response.ResponseData{data=model.FileResponse} "更新成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 401 {object} response.ResponseData "未授权"
// @Failure 403 {object} response.ResponseData "访问被拒绝"
// @Failure 404 {object} response.ResponseData "文件不存在"
// @Router /files/{id} [put]
func (h *FileHandler) UpdateFile(c *gin.Context) {
	// 获取文件ID
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的文件ID", nil)
		return
	}

	// 获取用户ID
	payload, exists := c.Get(middleware.AuthorizationPayloadKey)
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "无法获取授权信息", nil)
		return
	}

	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息格式错误", nil)
		return
	}

	userID := claims.UserID

	// 绑定请求体
	var req model.FileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "参数绑定失败", err.Error())
		return
	}

	// 更新文件
	file, err := h.fileService.UpdateFile(c.Request.Context(), id, &userID, &req)
	if err != nil {
		if err == response.ErrFileNotFound {
			response.ErrorResponse(c, http.StatusNotFound, "文件不存在", nil)
			return
		}
		if err == response.ErrFileAccessDenied {
			response.ErrorResponse(c, http.StatusForbidden, "访问被拒绝", nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "更新文件失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "更新成功", file)
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 删除指定的文件（仅文件所有者可操作）
// @Tags files
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "文件ID"
// @Success 200 {object} response.ResponseData "删除成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 401 {object} response.ResponseData "未授权"
// @Failure 403 {object} response.ResponseData "访问被拒绝"
// @Failure 404 {object} response.ResponseData "文件不存在"
// @Router /files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	// 获取文件ID
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的文件ID", nil)
		return
	}

	// 获取用户ID
	payload, exists := c.Get(middleware.AuthorizationPayloadKey)
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "无法获取授权信息", nil)
		return
	}

	claims, ok := payload.(*service.JWTClaims)
	if !ok {
		response.ErrorResponse(c, http.StatusUnauthorized, "授权信息格式错误", nil)
		return
	}

	userID := claims.UserID

	// 删除文件
	err = h.fileService.DeleteFile(c.Request.Context(), id, &userID)
	if err != nil {
		if err == response.ErrFileNotFound {
			response.ErrorResponse(c, http.StatusNotFound, "文件不存在", nil)
			return
		}
		if err == response.ErrFileAccessDenied {
			response.ErrorResponse(c, http.StatusForbidden, "访问被拒绝", nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "删除文件失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "删除成功", nil)
}

// GetStorageInfo 获取存储信息
// @Summary 获取存储信息
// @Description 获取系统可用的存储配置信息
// @Tags files
// @Produce json
// @Success 200 {object} response.ResponseData{data=model.StorageInfoResponse} "获取成功"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /files/storages [get]
func (h *FileHandler) GetStorageInfo(c *gin.Context) {
	info, err := h.fileService.GetStorageInfo(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取存储信息失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取成功", info)
} 