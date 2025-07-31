package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseData 统一响应结构
type ResponseData struct {
	Code      int         `json:"code" example:"200"`
	Message   string      `json:"message" example:"成功"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp" example:"1640995200"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, httpCode int, message string, data interface{}) {
	c.JSON(httpCode, ResponseData{
		Code:      httpCode,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, httpCode int, message string, err interface{}) {
	c.JSON(httpCode, ResponseData{
		Code:      httpCode,
		Message:   message,
		Error:     err,
		Timestamp: time.Now().Unix(),
	})
} 