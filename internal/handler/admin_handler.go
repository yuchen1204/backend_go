package handler

import (
	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/response"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	adminConfig config.AdminConfig
	jwtService  service.JwtService
	userService service.UserService
}

// NewAdminHandler 创建管理员处理器实例
func NewAdminHandler(adminConfig config.AdminConfig, jwtService service.JwtService, userService service.UserService) *AdminHandler {
	return &AdminHandler{
		adminConfig: adminConfig,
		jwtService:  jwtService,
		userService: userService,
	}
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
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID格式", err.Error())
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive banned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	err = h.userService.UpdateUserStatusByUUID(userID, req.Status)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "更新用户状态失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "用户状态更新成功", nil)
}

// DeleteUser 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID格式", err.Error())
		return
	}

	err = h.userService.DeleteUserByUUID(userID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "删除用户失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "用户删除成功", nil)
}

// GetUserStats 获取用户统计信息
func (h *AdminHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取用户统计失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取用户统计成功", stats)
}
