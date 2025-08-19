package handler

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/response"
	"backend/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// SendVerificationCode 发送注册验证码
// @Summary 发送注册验证码
// @Description 在发送验证码前，会预先检查用户名和邮箱是否都未被注册。都通过后，才会向指定邮箱发送一个用于注册的6位数验证码（5分钟内有效）。
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body model.SendCodeRequest true "用户名和邮箱信息"
// @Success 200 {object} response.ResponseData "验证码发送成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 409 {object} response.ResponseData "该用户名或邮箱已被注册"
// @Failure 429 {object} response.ResponseData "请求过于频繁"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/send-code [post]
func (h *UserHandler) SendVerificationCode(c *gin.Context) {
	var req model.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取客户端IP
	clientIP := c.ClientIP()

	err := h.userService.SendVerificationCode(c.Request.Context(), &req, clientIP)
	if err != nil {
		if err.Error() == "该邮箱已被注册" || err.Error() == "该用户名已被注册" {
			response.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		// 检查是否是频率限制错误
		if strings.Contains(err.Error(), "请求过于频繁") {
			response.ErrorResponse(c, http.StatusTooManyRequests, err.Error(), nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "发送验证码失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "验证码已发送至您的邮箱，请注意查收", nil)
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名和密码登录，成功后返回包含Access Token、Refresh Token和用户信息的对象
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "登录凭证"
// @Success 200 {object} response.ResponseData{data=model.LoginResponse} "登录成功"
// @Failure 400 {object} response.ResponseData "请求参数错误或设备验证码相关错误"
// @Failure 401 {object} response.ResponseData "用户名或密码错误"
// @Failure 403 {object} response.ResponseData "账户已被封禁或未激活"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 由服务器端填充来源信息（IP 和 User-Agent），避免客户端伪造
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.Request.UserAgent()

	res, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		errMsg := err.Error()
		
		// 处理认证相关错误
		if errMsg == "用户名或密码错误" {
			response.ErrorResponse(c, http.StatusUnauthorized, errMsg, nil)
			return
		}
		
		// 处理账户状态相关错误
		if errMsg == "账户已被封禁，无法登录" {
			response.ErrorResponse(c, http.StatusForbidden, errMsg, nil)
			return
		}
		
		if errMsg == "账户未激活，无法登录" {
			response.ErrorResponse(c, http.StatusForbidden, errMsg, nil)
			return
		}
		
		// 处理设备验证相关错误
		if strings.Contains(errMsg, "验证码") {
			response.ErrorResponse(c, http.StatusBadRequest, errMsg, nil)
			return
		}
		
		// 其他服务器内部错误
		response.ErrorResponse(c, http.StatusInternalServerError, "登录失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "登录成功", res)
}

// RefreshToken 刷新访问Token
// @Summary 刷新访问Token
// @Description 使用有效的Refresh Token获取新的Access Token
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "刷新Token请求"
// @Success 200 {object} response.ResponseData{data=model.RefreshTokenResponse} "刷新成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 401 {object} response.ResponseData "Refresh Token无效或已过期"
// @Failure 403 {object} response.ResponseData "账户已被封禁或未激活"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	res, err := h.userService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		errMsg := err.Error()
		
		// 处理Token相关错误
		if strings.Contains(errMsg, "无效") || strings.Contains(errMsg, "失效") || strings.Contains(errMsg, "不存在") {
			response.ErrorResponse(c, http.StatusUnauthorized, errMsg, nil)
			return
		}
		
		// 处理账户状态相关错误
		if strings.Contains(errMsg, "已被封禁") {
			response.ErrorResponse(c, http.StatusForbidden, errMsg, nil)
			return
		}
		
		if strings.Contains(errMsg, "未激活") {
			response.ErrorResponse(c, http.StatusForbidden, errMsg, nil)
			return
		}
		
		response.ErrorResponse(c, http.StatusInternalServerError, "刷新Token失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "刷新成功", res)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 登出用户并撤销所有Token（Access Token和Refresh Token）。Access Token将被加入黑名单，立即失效。
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body model.LogoutRequest true "登出请求（需要提供access_token和refresh_token）"
// @Success 200 {object} response.ResponseData "登出成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 401 {object} response.ResponseData "Token无效或两个Token不属于同一用户"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	var req model.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	err := h.userService.Logout(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "无效") {
			response.ErrorResponse(c, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "登出失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "登出成功", nil)
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 使用验证码重置密码
// @Tags 密码管理
// @Accept json
// @Produce json
// @Param request body model.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} response.ResponseData "密码重置成功"
// @Failure 400 {object} response.ResponseData "请求参数错误或验证码错误"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	err := h.userService.ResetPassword(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "密码重置失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "密码重置成功", nil)
}

// SendActivationCode 发送账户激活验证码
// @Summary 发送账户激活验证码
// @Description 为非活跃账户发送激活验证码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body model.SendActivationCodeRequest true "激活验证码请求"
// @Success 200 {object} response.ResponseData "验证码发送成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 429 {object} response.ResponseData "请求过于频繁"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/send-activation-code [post]
func (h *UserHandler) SendActivationCode(c *gin.Context) {
	var req model.SendActivationCodeRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取客户端IP
	ip := c.ClientIP()

	// 调用服务层发送激活验证码
	err := h.userService.SendActivationCode(c.Request.Context(), &req, ip)
	if err != nil {
		if err.Error() == "请求过于频繁，请稍后再试" {
			response.ErrorResponse(c, http.StatusTooManyRequests, "请求过于频繁", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "发送激活验证码失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "激活验证码已发送到您的邮箱", nil)
}

// ActivateAccount 激活账户
// @Summary 激活账户
// @Description 使用验证码激活非活跃账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body model.ActivateAccountRequest true "账户激活请求"
// @Success 200 {object} response.ResponseData "账户激活成功"
// @Failure 400 {object} response.ResponseData "请求参数错误或验证码错误"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/activate [post]
func (h *UserHandler) ActivateAccount(c *gin.Context) {
	var req model.ActivateAccountRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 调用服务层激活账户
	err := h.userService.ActivateAccount(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "账户激活失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "账户激活成功，现在可以正常登录", nil)
}

// GetMe 获取当前登录用户信息
// @Summary 获取当前用户信息
// @Description 根据请求头中的JWT获取当前登录用户的详细信息
// @Tags 用户管理
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} response.ResponseData{data=model.UserResponse} "获取成功"
// @Failure 401 {object} response.ResponseData "未授权或Token无效"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	// 从上下文中获取payload
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

	// 使用payload中的用户ID获取用户信息
	user, err := h.userService.GetByID(claims.UserID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "获取用户信息失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取成功", user)
}

// UpdateProfile 更新当前用户信息
// @Summary 更新当前用户信息
// @Description 更新当前登录用户的基本信息（昵称、简介、头像）
// @Tags 用户管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body model.UpdateProfileRequest true "更新信息"
// @Success 200 {object} response.ResponseData{data=model.UserResponse} "更新成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 401 {object} response.ResponseData "未授权或Token无效"
// @Failure 404 {object} response.ResponseData "用户不存在"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// 从上下文中获取payload
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

	// 绑定请求参数
	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 调用服务层更新用户信息
	updatedUser, err := h.userService.UpdateProfile(c.Request.Context(), claims.UserID, &req)
	if err != nil {
		if err.Error() == "用户不存在" {
			response.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "更新用户信息失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "更新成功", updatedUser)
}

// Register 用户注册
// @Summary 用户注册
// @Description 使用邮箱验证码创建新用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body model.UserRegisterRequest true "注册信息"
// @Success 201 {object} response.ResponseData{data=model.UserResponse} "注册成功"
// @Failure 400 {object} response.ResponseData "请求参数错误或验证码错误"
// @Failure 409 {object} response.ResponseData "用户名或邮箱已存在"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req model.UserRegisterRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 调用服务层进行注册
	user, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		// 根据错误类型返回不同的状态码
		if err.Error() == "用户名已存在" || err.Error() == "邮箱已存在" {
			response.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		if err.Error() == "验证码错误" || err.Error() == "验证码已过期或不存在，请重新获取" {
			response.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "注册失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusCreated, "注册成功", user)
}

// GetUserByID 根据ID获取用户信息
// @Summary 根据ID获取用户信息
// @Description 通过用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} response.ResponseData{data=model.UserResponse} "获取成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 404 {object} response.ResponseData "用户不存在"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	
	// 解析UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID", err.Error())
		return
	}

	// 获取用户信息
	user, err := h.userService.GetByID(id)
	if err != nil {
		if err.Error() == "用户不存在" {
			response.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "获取用户信息失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取成功", user)
}

// GetUserByUsername 根据用户名获取用户信息
// @Summary 根据用户名获取用户信息
// @Description 通过用户名获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param username path string true "用户名"
// @Success 200 {object} response.ResponseData{data=model.UserResponse} "获取成功"
// @Failure 404 {object} response.ResponseData "用户不存在"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/username/{username} [get]
func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	
	// 获取用户信息
	user, err := h.userService.GetByUsername(username)
	if err != nil {
		if err.Error() == "用户不存在" {
			response.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "获取用户信息失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "获取成功", user)
}

// SendResetPasswordCode 发送重置密码验证码
// @Summary 发送重置密码验证码
// @Description 向指定邮箱发送用于重置密码的6位数验证码（5分钟内有效）。如果邮箱未注册，为了安全考虑也会返回成功，避免邮箱枚举攻击。
// @Tags 密码管理
// @Accept json
// @Produce json
// @Param request body model.SendResetCodeRequest true "发送重置密码验证码请求"
// @Success 200 {object} response.ResponseData "验证码发送成功"
// @Failure 400 {object} response.ResponseData "请求参数错误"
// @Failure 429 {object} response.ResponseData "请求过于频繁"
// @Failure 500 {object} response.ResponseData "服务器内部错误"
// @Router /users/send-reset-code [post]
func (h *UserHandler) SendResetPasswordCode(c *gin.Context) {
	var req model.SendResetCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取客户端IP
	ip := c.ClientIP()

	err := h.userService.SendResetPasswordCode(c.Request.Context(), &req, ip)
	if err != nil {
		if strings.Contains(err.Error(), "请求过于频繁") {
			response.ErrorResponse(c, http.StatusTooManyRequests, err.Error(), nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "发送验证码失败", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "验证码已发送至您的邮箱，请注意查收", nil)
}

 