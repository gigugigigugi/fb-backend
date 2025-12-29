package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderXRequestID = "X-Request-Id"
const ContextRequestID = "request_id"

// RequestID 生成唯一 ID 并注入 Context
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 尝试从请求头获取
		reqID := c.GetHeader(HeaderXRequestID)

		// 2. 如果没有，生成新的
		if reqID == "" {
			reqID = uuid.New().String()
		}

		// 3. 响应头带回去
		c.Header(HeaderXRequestID, reqID)

		// 4. 存入 Context 供后续使用
		c.Set(ContextRequestID, reqID)

		c.Next()
	}
}
