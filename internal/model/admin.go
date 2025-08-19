package model

// AdminLoginRequest 管理员登录请求结构
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
