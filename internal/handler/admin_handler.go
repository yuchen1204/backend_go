package handler

import (
	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	adminConfig config.AdminConfig
	jwtService  service.JwtService
	userService service.UserService
	adminLogService service.AdminLogService
	userActionLogService service.UserActionLogService
	fileService service.FileService
	friendBanRepo repository.FriendBanRepository
}

// AdminSetFriendBan 管理员：设置用户好友功能封禁
// @Summary 管理员设置用户好友功能封禁
// @Tags admin-users
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param request body model.AdminSetFriendBanRequest true "封禁请求体"
// @Success 200 {object} response.ResponseData
// @Failure 400 {object} response.ResponseData
// @Router /admin/users/{id}/friend-ban [post]
func (h *AdminHandler) AdminSetFriendBan(c *gin.Context) {
    userIDStr := c.Param("id")
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID格式", err.Error())
        return
    }
    var req model.AdminSetFriendBanRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
        return
    }
    if req.BannedUntil.Before(time.Now()) {
        response.ErrorResponse(c, http.StatusBadRequest, "封禁截止时间必须晚于当前时间", nil)
        return
    }
    if err := h.friendBanRepo.SetBan(userID, req.Reason, req.BannedUntil); err != nil {
        response.ErrorResponse(c, http.StatusInternalServerError, "设置封禁失败", err.Error())
        return
    }

    // 管理员操作日志
    if adminUsername, exists := c.Get("admin_username"); exists {
        detailsObj := map[string]any{
            "target_user_id": userID.String(),
            "reason": req.Reason,
            "banned_until": req.BannedUntil,
        }
        detailsBytes, _ := json.Marshal(detailsObj)
        logEntry := &model.AdminActionLog{
            AdminUsername: adminUsername.(string),
            Action:        "set_friend_ban",
            TargetUserID:  &userID,
            Details:       string(detailsBytes),
            IPAddress:     c.ClientIP(),
            UserAgent:     c.GetHeader("User-Agent"),
        }
        _ = h.adminLogService.Create(c.Request.Context(), logEntry)
    }
    response.SuccessResponse(c, http.StatusOK, "设置封禁成功", nil)
}

// AdminRemoveFriendBan 管理员：解除用户好友功能封禁
// @Summary 管理员解除用户好友功能封禁
// @Tags admin-users
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} response.ResponseData
// @Failure 400 {object} response.ResponseData
// @Router /admin/users/{id}/friend-ban [delete]
func (h *AdminHandler) AdminRemoveFriendBan(c *gin.Context) {
    userIDStr := c.Param("id")
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID格式", err.Error())
        return
    }
    if err := h.friendBanRepo.RemoveBan(userID); err != nil {
        response.ErrorResponse(c, http.StatusInternalServerError, "解除封禁失败", err.Error())
        return
    }
    if adminUsername, exists := c.Get("admin_username"); exists {
        detailsObj := map[string]any{
            "target_user_id": userID.String(),
        }
        detailsBytes, _ := json.Marshal(detailsObj)
        logEntry := &model.AdminActionLog{
            AdminUsername: adminUsername.(string),
            Action:        "remove_friend_ban",
            TargetUserID:  &userID,
            Details:       string(detailsBytes),
            IPAddress:     c.ClientIP(),
            UserAgent:     c.GetHeader("User-Agent"),
        }
        _ = h.adminLogService.Create(c.Request.Context(), logEntry)
    }
    response.SuccessResponse(c, http.StatusOK, "解除封禁成功", nil)
}

// AdminGetFriendBan 管理员：查询用户好友功能封禁状态
// @Summary 管理员查询用户好友功能封禁状态
// @Tags admin-users
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} response.ResponseData{data=map[string]any}
// @Failure 400 {object} response.ResponseData
// @Router /admin/users/{id}/friend-ban [get]
func (h *AdminHandler) AdminGetFriendBan(c *gin.Context) {
    userIDStr := c.Param("id")
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID格式", err.Error())
        return
    }
    ban, err := h.friendBanRepo.GetActiveBan(userID, time.Now())
    if err != nil {
        // 未找到或已过期：视为未封禁
        response.SuccessResponse(c, http.StatusOK, "查询成功", gin.H{
            "active": false,
        })
        return
    }
    response.SuccessResponse(c, http.StatusOK, "查询成功", gin.H{
        "active": true,
        "reason": ban.Reason,
        "banned_until": ban.BannedUntil,
    })
}

// AdminGetStorageInfo 管理员：获取存储信息（含本地与S3 bucket列表）
// @Summary 管理员获取存储信息
// @Tags admin-files
// @Produce json
// @Success 200 {object} response.ResponseData{data=model.StorageInfoResponse}
// @Router /admin/storage/info [get]
func (h *AdminHandler) AdminGetStorageInfo(c *gin.Context) {
    info, err := h.fileService.GetStorageInfo(c.Request.Context())
    if err != nil {
        response.ErrorResponse(c, http.StatusInternalServerError, "获取存储信息失败", err.Error())
        return
    }
    response.SuccessResponse(c, http.StatusOK, "获取成功", info)
}

// AdminListFiles 管理员：获取所有文件列表
// @Summary 管理员获取文件列表
// @Description 分页筛选所有文件（公开与私有）
// @Tags admin-files
// @Produce json
// @Param category query string false "文件分类"
// @Param storage_type query string false "存储类型"
// @Param storage_name query string false "存储名称"
// @Param is_public query boolean false "是否公开"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} response.ResponseData{data=model.FileListResponse}
// @Failure 500 {object} response.ResponseData
// @Router /admin/files [get]
func (h *AdminHandler) AdminListFiles(c *gin.Context) {
    var req model.FileListRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "参数绑定失败", err.Error())
        return
    }
    files, err := h.fileService.GetAllFiles(c.Request.Context(), &req)
    if err != nil {
        response.ErrorResponse(c, http.StatusInternalServerError, "获取文件列表失败", err.Error())
        return
    }
    response.SuccessResponse(c, http.StatusOK, "获取成功", files)
}

// AdminListPublicFiles 管理员：获取公开文件列表
// @Summary 管理员获取公开文件列表
// @Description 分页筛选公开文件
// @Tags admin-files
// @Produce json
// @Param category query string false "文件分类"
// @Param storage_type query string false "存储类型"
// @Param storage_name query string false "存储名称"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} response.ResponseData{data=model.FileListResponse}
// @Failure 500 {object} response.ResponseData
// @Router /admin/files/public [get]
func (h *AdminHandler) AdminListPublicFiles(c *gin.Context) {
    var req model.FileListRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "参数绑定失败", err.Error())
        return
    }
    // 强制只查公开文件
    t := true
    req.IsPublic = &t
    files, err := h.fileService.GetAllFiles(c.Request.Context(), &req)
    if err != nil {
        response.ErrorResponse(c, http.StatusInternalServerError, "获取文件列表失败", err.Error())
        return
    }
    response.SuccessResponse(c, http.StatusOK, "获取成功", files)
}

// AdminGetFile 管理员：获取文件详情
// @Summary 管理员获取文件详情
// @Tags admin-files
// @Produce json
// @Param id path string true "文件ID"
// @Success 200 {object} response.ResponseData{data=model.FileResponse}
// @Failure 404 {object} response.ResponseData
// @Router /admin/files/{id} [get]
func (h *AdminHandler) AdminGetFile(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "无效的文件ID", nil)
        return
    }
    file, err := h.fileService.AdminGetFile(c.Request.Context(), id)
    if err != nil {
        if err == response.ErrFileNotFound {
            response.ErrorResponse(c, http.StatusNotFound, "文件不存在", nil)
            return
        }
        response.ErrorResponse(c, http.StatusInternalServerError, "获取文件失败", err.Error())
        return
    }
    response.SuccessResponse(c, http.StatusOK, "获取成功", file)
}

// AdminUpdateFile 管理员：更新文件
// @Summary 管理员更新文件
// @Tags admin-files
// @Accept json
// @Produce json
// @Param id path string true "文件ID"
// @Param request body model.FileUpdateRequest true "更新请求"
// @Success 200 {object} response.ResponseData{data=model.FileResponse}
// @Failure 400 {object} response.ResponseData
// @Failure 404 {object} response.ResponseData
// @Router /admin/files/{id} [put]
func (h *AdminHandler) AdminUpdateFile(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "无效的文件ID", nil)
        return
    }
    var req model.FileUpdateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "参数绑定失败", err.Error())
        return
    }
    file, err := h.fileService.AdminUpdateFile(c.Request.Context(), id, &req)
    if err != nil {
        if err == response.ErrFileNotFound {
            response.ErrorResponse(c, http.StatusNotFound, "文件不存在", nil)
            return
        }
        response.ErrorResponse(c, http.StatusInternalServerError, "更新文件失败", err.Error())
        return
    }

    // 管理员操作日志
    if adminUsername, exists := c.Get("admin_username"); exists {
        detailsObj := map[string]any{
            "file_id": id.String(),
            "note":    "管理员更新文件信息",
        }
        detailsBytes, _ := json.Marshal(detailsObj)
        logEntry := &model.AdminActionLog{
            AdminUsername: adminUsername.(string),
            Action:        "update_file",
            Details:       string(detailsBytes),
            IPAddress:     c.ClientIP(),
            UserAgent:     c.GetHeader("User-Agent"),
        }
        _ = h.adminLogService.Create(c.Request.Context(), logEntry)
    }

    response.SuccessResponse(c, http.StatusOK, "更新成功", file)
}

// AdminDeleteFile 管理员：删除文件
// @Summary 管理员删除文件
// @Tags admin-files
// @Produce json
// @Param id path string true "文件ID"
// @Success 200 {object} response.ResponseData
// @Failure 400 {object} response.ResponseData
// @Failure 404 {object} response.ResponseData
// @Router /admin/files/{id} [delete]
func (h *AdminHandler) AdminDeleteFile(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "无效的文件ID", nil)
        return
    }
    if err := h.fileService.AdminDeleteFile(c.Request.Context(), id); err != nil {
        if err == response.ErrFileNotFound {
            response.ErrorResponse(c, http.StatusNotFound, "文件不存在", nil)
            return
        }
        response.ErrorResponse(c, http.StatusInternalServerError, "删除文件失败", err.Error())
        return
    }

    // 管理员操作日志
    if adminUsername, exists := c.Get("admin_username"); exists {
        detailsObj := map[string]any{
            "file_id": id.String(),
            "note":    "管理员删除文件",
        }
        detailsBytes, _ := json.Marshal(detailsObj)
        logEntry := &model.AdminActionLog{
            AdminUsername: adminUsername.(string),
            Action:        "delete_file",
            Details:       string(detailsBytes),
            IPAddress:     c.ClientIP(),
            UserAgent:     c.GetHeader("User-Agent"),
        }
        _ = h.adminLogService.Create(c.Request.Context(), logEntry)
    }

    response.SuccessResponse(c, http.StatusOK, "删除成功", nil)
}

// UpdateUserPassword 管理员更新用户密码
// @Summary 管理员更新用户密码
// @Tags admin-users
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param request body model.AdminUpdatePasswordRequest true "密码更新请求"
// @Success 200 {object} response.ResponseData
// @Router /admin/users/{id}/password [put]
func (h *AdminHandler) UpdateUserPassword(c *gin.Context) {
    userIDStr := c.Param("id")
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID格式", err.Error())
        return
    }

    var req model.AdminUpdatePasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
        return
    }

    if err := h.userService.AdminUpdateUserPassword(c.Request.Context(), userID, req.NewPassword); err != nil {
        response.ErrorResponse(c, http.StatusInternalServerError, "更新用户密码失败", err.Error())
        return
    }

    // 自动记录管理员操作日志：重置用户密码
    if adminUsername, exists := c.Get("admin_username"); exists {
        detailsObj := map[string]any{
            "target_user_id": userID.String(),
            "note":            "管理员重置用户密码",
        }
        detailsBytes, _ := json.Marshal(detailsObj)
        logEntry := &model.AdminActionLog{
            AdminUsername: adminUsername.(string),
            Action:        "reset_user_password",
            TargetUserID:  &userID,
            Details:       string(detailsBytes),
            IPAddress:     c.ClientIP(),
            UserAgent:     c.GetHeader("User-Agent"),
        }
        _ = h.adminLogService.Create(c.Request.Context(), logEntry)
    }

    response.SuccessResponse(c, http.StatusOK, "用户密码更新成功", nil)
}

// AdminRefreshToken 刷新管理员Token
// @Summary 刷新管理员Token
// @Tags admin-auth
// @Produce json
// @Success 200 {object} response.ResponseData{data=map[string]string}
// @Router /admin/refresh-token [post]
func (h *AdminHandler) AdminRefreshToken(c *gin.Context) {
    adminUsername, exists := c.Get("admin_username")
    if !exists {
        response.ErrorResponse(c, http.StatusUnauthorized, "未授权", nil)
        return
    }

    token, err := h.jwtService.GenerateAdminToken(adminUsername.(string))
    if err != nil {
        response.ErrorResponse(c, http.StatusInternalServerError, "生成Token失败", err.Error())
        return
    }

    response.SuccessResponse(c, http.StatusOK, "刷新成功", gin.H{"token": token})
}

// NewAdminHandler 创建管理员处理器实例
func NewAdminHandler(adminConfig config.AdminConfig, jwtService service.JwtService, userService service.UserService, adminLogService service.AdminLogService, userActionLogService service.UserActionLogService, fileService service.FileService, friendBanRepo repository.FriendBanRepository) *AdminHandler {
    return &AdminHandler{
        adminConfig: adminConfig,
        jwtService:  jwtService,
        userService: userService,
        adminLogService: adminLogService,
        userActionLogService: userActionLogService,
        fileService: fileService,
        friendBanRepo: friendBanRepo,
    }
}

// GetTrafficStats 管理员获取网络流量统计（占位实现）
// @Summary 管理员获取网络流量统计
// @Tags admin-stats
// @Produce json
// @Success 200 {object} response.ResponseData{data=map[string]any}
// @Router /admin/stats/traffic [get]
func (h *AdminHandler) GetTrafficStats(c *gin.Context) {
    inBytes := middleware.GetTrafficInBytes()
    outBytes := middleware.GetTrafficOutBytes()
    data := map[string]any{
        "in_bytes":  inBytes,
        "out_bytes": outBytes,
        // 当前为进程自启动以来的累计值，后续可扩展时间窗口参数
        "window":    "since_start",
    }
    response.SuccessResponse(c, http.StatusOK, "获取成功", data)
}

// Login 管理员登录
func (h *AdminHandler) Login(c *gin.Context) {
	var req model.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	if req.Username != h.adminConfig.User || req.Password != h.adminConfig.Password {
		response.ErrorResponse(c, http.StatusUnauthorized, "用户名或密码错误", nil)
		return
	}

	// 生成管理员专用的JWT
	token, err := h.jwtService.GenerateAdminToken(req.Username)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "生成Token失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "登录成功", gin.H{"token": token})
}

// GetUsers 获取用户列表
// @Summary 管理员获取用户列表
// @Tags admin-users
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param search query string false "搜索关键词"
// @Success 200 {object} response.ResponseData{data=map[string]any}
// @Router /admin/users [get]
func (h *AdminHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	users, total, err := h.userService.GetUsersForAdmin(page, limit, search)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取用户列表失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取用户列表成功", gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetUserDetail 获取用户详情
// @Summary 管理员获取用户详情
// @Tags admin-users
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} response.ResponseData{data=model.UserResponse}
// @Router /admin/users/{id} [get]
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID格式", err.Error())
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "用户不存在", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取用户详情成功", user)
}

// UpdateUserStatus 更新用户状态
// @Summary 管理员更新用户状态
// @Tags admin-users
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param request body model.UserStatusUpdateRequest true "状态更新请求"
// @Success 200 {object} response.ResponseData
// @Router /admin/users/{id}/status [put]
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID格式", err.Error())
		return
	}

	var req model.UserStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	err = h.userService.UpdateUserStatusByUUID(userID, req.Status)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "更新用户状态失败", err.Error())
		return
	}

    // 自动记录管理员操作日志：更新用户状态
    if adminUsername, exists := c.Get("admin_username"); exists {
        detailsObj := map[string]any{
            "target_user_id": userID.String(),
            "new_status":     req.Status,
        }
        detailsBytes, _ := json.Marshal(detailsObj)
        logEntry := &model.AdminActionLog{
            AdminUsername: adminUsername.(string),
            Action:        "update_user_status",
            TargetUserID:  &userID,
            Details:       string(detailsBytes),
            IPAddress:     c.ClientIP(),
            UserAgent:     c.GetHeader("User-Agent"),
        }
        _ = h.adminLogService.Create(c.Request.Context(), logEntry)
    }

	response.SuccessResponse(c, http.StatusOK, "用户状态更新成功", nil)
}

// DeleteUser 删除用户
// @Summary 管理员删除用户
// @Tags admin-users
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} response.ResponseData
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID格式", err.Error())
		return
	}

	// 防止管理员删除自己的账户
	if adminUserID, exists := c.Get("admin_user_id"); exists {
		if adminUUID, ok := adminUserID.(uuid.UUID); ok && adminUUID == userID {
			response.ErrorResponse(c, http.StatusForbidden, "不能删除自己的账户", nil)
			return
		}
	}

	err = h.userService.DeleteUserByUUID(userID)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "用户不存在" {
			response.ErrorResponse(c, http.StatusNotFound, errMsg, nil)
			return
		}
		if errMsg == "系统保护用户不能删除" {
			response.ErrorResponse(c, http.StatusForbidden, errMsg, nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "删除用户失败", err.Error())
		return
	}

    // 自动记录管理员操作日志：删除用户
    if adminUsername, exists := c.Get("admin_username"); exists {
        detailsObj := map[string]any{
            "target_user_id": userID.String(),
            "note":            "管理员删除用户",
        }
        detailsBytes, _ := json.Marshal(detailsObj)
        logEntry := &model.AdminActionLog{
            AdminUsername: adminUsername.(string),
            Action:        "delete_user",
            TargetUserID:  &userID,
            Details:       string(detailsBytes),
            IPAddress:     c.ClientIP(),
            UserAgent:     c.GetHeader("User-Agent"),
        }
        _ = h.adminLogService.Create(c.Request.Context(), logEntry)
    }

	response.SuccessResponse(c, http.StatusOK, "用户删除成功", nil)
}

// GetUserStats 获取用户统计信息
// @Summary 管理员获取用户统计信息
// @Tags admin-stats
// @Produce json
// @Success 200 {object} response.ResponseData{data=map[string]any}
// @Router /admin/stats/users [get]
func (h *AdminHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取用户统计失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取用户统计成功", stats)
}

// CreateAdminLog 创建管理员行为日志
// @Summary 创建管理员行为日志
// @Tags admin-logs
// @Accept json
// @Produce json
// @Param request body model.AdminLogCreateRequest true "日志创建请求"
// @Success 200 {object} response.ResponseData{data=map[string]any}
// @Router /admin/logs [post]
func (h *AdminHandler) CreateAdminLog(c *gin.Context) {
    var req model.AdminLogCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
        return
    }

    adminUsername, _ := c.Get("admin_username")
    ip := c.ClientIP()
    ua := c.GetHeader("User-Agent")

    logEntry := &model.AdminActionLog{
        AdminUsername: adminUsername.(string),
        Action:        req.Action,
        TargetUserID:  req.TargetUserID,
        Details:       req.Details,
        IPAddress:     ip,
        UserAgent:     ua,
    }

    if err := h.adminLogService.Create(c.Request.Context(), logEntry); err != nil {
        response.ErrorResponse(c, http.StatusInternalServerError, "创建日志失败", err.Error())
        return
    }

    response.SuccessResponse(c, http.StatusOK, "日志创建成功", gin.H{"id": logEntry.ID})
}

// ListAdminLogs 查询管理员行为日志
// @Summary 查询管理员行为日志
// @Tags admin-logs
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param admin_username query string false "管理员用户名"
// @Param action query string false "操作类型"
// @Success 200 {object} response.ResponseData{data=map[string]any}
// @Router /admin/logs [get]
func (h *AdminHandler) ListAdminLogs(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    adminUsername := c.Query("admin_username")
    action := c.Query("action")

    logs, total, err := h.adminLogService.List(c.Request.Context(), page, limit, adminUsername, action)
    if err != nil {
        response.ErrorResponse(c, http.StatusInternalServerError, "获取日志失败", err.Error())
        return
    }

    response.SuccessResponse(c, http.StatusOK, "获取日志成功", gin.H{
        "logs": logs,
        "total": total,
        "page": page,
        "limit": limit,
    })
}

// ListUserActionLogs 分页查询某用户的行为日志
// @Summary 管理员查询用户行为日志
// @Tags admin-users
// @Produce json
// @Param id path string true "用户ID"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} response.ResponseData{data=map[string]any}
// @Router /admin/users/{id}/action-logs [get]
func (h *AdminHandler) ListUserActionLogs(c *gin.Context) {
    userIDStr := c.Param("id")
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID格式", err.Error())
        return
    }

    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    logs, total, err := h.userActionLogService.ListByUser(c.Request.Context(), userID, page, limit)
    if err != nil {
        response.ErrorResponse(c, http.StatusInternalServerError, "获取用户行为日志失败", err.Error())
        return
    }

    response.SuccessResponse(c, http.StatusOK, "获取用户行为日志成功", gin.H{
        "logs": logs,
        "total": total,
        "page": page,
        "limit": limit,
    })
}
