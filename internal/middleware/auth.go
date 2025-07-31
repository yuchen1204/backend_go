package middleware

import (
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// AuthorizationHeaderKey is the key for the authorization header.
	AuthorizationHeaderKey = "Authorization"
	// AuthorizationTypeBearer is the type of the authorization token.
	AuthorizationTypeBearer = "bearer"
	// AuthorizationPayloadKey is the key for the authorization payload in the context.
	AuthorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware creates a gin middleware for authentication.
func AuthMiddleware(jwtSvc service.JwtService, blacklistRepo repository.AccessTokenBlacklistRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the authorization header.
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authHeader) == 0 {
			response.ErrorResponse(c, http.StatusUnauthorized, "未提供授权头", nil)
			c.Abort()
			return
		}

		// 2. Check if the authorization header is in the correct format.
		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			response.ErrorResponse(c, http.StatusUnauthorized, "授权头格式无效", nil)
			c.Abort()
			return
		}

		// 3. Check if the authorization type is Bearer.
		authType := strings.ToLower(fields[0])
		if authType != AuthorizationTypeBearer {
			response.ErrorResponse(c, http.StatusUnauthorized, "不支持的授权类型", nil)
			c.Abort()
			return
		}

		// 4. Validate the token.
		accessToken := fields[1]
		payload, err := jwtSvc.ValidateToken(accessToken)
		if err != nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "无效的token", err.Error())
			c.Abort()
			return
		}

		// 5. Ensure this is an access token
		if payload.TokenType != service.AccessToken {
			response.ErrorResponse(c, http.StatusUnauthorized, "必须使用access token", nil)
			c.Abort()
			return
		}

		// 6. Check if the access token is blacklisted
		isBlacklisted, err := blacklistRepo.IsBlacklisted(c.Request.Context(), accessToken)
		if err != nil {
			response.ErrorResponse(c, http.StatusInternalServerError, "验证token黑名单状态失败", err.Error())
			c.Abort()
			return
		}
		if isBlacklisted {
			response.ErrorResponse(c, http.StatusUnauthorized, "token已被撤销", nil)
			c.Abort()
			return
		}

		// 7. Set the payload in the context.
		c.Set(AuthorizationPayloadKey, payload)
		c.Next()
	}
} 