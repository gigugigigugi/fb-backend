package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger 替换 Gin 默认的 Logger
func RequestLogger() gin.HandlerFunc {
	skipPaths := map[string]bool{
		"/ping":        true,
		"/health":      true,
		"/favicon.ico": true,
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		if skipPaths[path] {
			c.Next()
			return
		}

		c.Next()

		cost := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		attrs := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", clientIP),
			slog.String("user-agent", c.Request.UserAgent()),
			slog.Duration("cost", cost),
		}

		// [修改点] 统一从 Context 获取 Request ID
		if reqID, exists := c.Get("request_id"); exists {
			attrs = append(attrs, slog.Any("req_id", reqID))
		}

		if uid, exists := c.Get("userID"); exists {
			attrs = append(attrs, slog.Any("user_id", uid))
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				attrs = append(attrs, slog.String("error", e))
			}
			slog.LogAttrs(c, slog.LevelError, "Request Failed", attrs...)
		} else {
			level := slog.LevelInfo
			if status >= 500 {
				level = slog.LevelError
			} else if status >= 400 {
				level = slog.LevelWarn
			}
			slog.LogAttrs(c, level, "Request Processed", attrs...)
		}
	}
}
