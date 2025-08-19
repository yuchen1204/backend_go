package middleware

import (
	"backend/internal/response"
	"backend/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware 创建一个管理员认证中间件
func AdminAuthMiddleware(jwtService service.JwtService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ErrorResponse(c, http.StatusUnauthorized, "请求未包含认证信息", nil)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.ErrorResponse(c, http.StatusUnauthorized, "认证信息格式错误", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtService.ValidateAdminToken(tokenString)
		if err != nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "无效或已过期的Token", err.Error())
			c.Abort()
			return
		}

		c.Set("admin_username", claims.Username)
		c.Next()
	}
}
