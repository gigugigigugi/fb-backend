package utils

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserID 从 Gin Context 中安全获取 UserID。
// 它自动处理多种可能的类型 (uint, int, float64, string)，并处理错误日志。
// 如果获取失败，它会自动返回 HTTP 错误响应并返回 (0, false)，调用者应直接 return。
func GetUserID(c *gin.Context) (uint, bool) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: UserID not found"})
		return 0, false
	}

	var userID uint
	var err error

	switch v := val.(type) {
	case uint:
		userID = v
	case int:
		if v < 0 {
			err = fmt.Errorf("negative integer")
		} else {
			userID = uint(v)
		}
	case float64:
		if v < 0 {
			err = fmt.Errorf("negative float")
		} else {
			userID = uint(v)
		}
	case string:
		var uid uint64
		uid, err = strconv.ParseUint(v, 10, 64)
		userID = uint(uid)
	default:
		err = fmt.Errorf("unsupported type %T", v)
	}

	if err != nil {
		// 记录严重的开发级错误日志
		slog.Error("Critical: Invalid userID type in context",
			slog.String("path", c.Request.URL.Path),
			slog.Any("value", val),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error: Invalid User Context"})
		return 0, false
	}

	return userID, true
}

// GetParamID 解析 URL 中的 ID 参数 (例如 /matches/:id/join)
// 如果解析失败，自动返回 400 错误并返回 (0, false)
func GetParamID(c *gin.Context, key string) (uint, bool) {
	idStr := c.Param(key)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid %s format", key)})
		return 0, false
	}
	return uint(id), true
}
