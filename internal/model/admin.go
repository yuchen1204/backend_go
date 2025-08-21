package model

// AdminLoginRequest 管理员登录请求结构
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminUpdatePasswordRequest 管理员更新用户密码请求结构
type AdminUpdatePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}
