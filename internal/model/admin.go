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

// UserStatusUpdateRequest 管理员更新用户状态请求结构
type UserStatusUpdateRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive banned"`
}
