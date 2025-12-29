package router

import (
	"football-backend/common/config" // 导入 config 包
	"football-backend/internal/service"
	"net/http"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures the routes for the application on the given Gin engine.
func SetupRouter(r *gin.Engine) {
	// 从配置中获取当前环境
	env := config.App.Env

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "WELCOME",
			"env":        env, // 显示正确的环境
			"go_version": runtime.Version(),
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "pong",
			"env":        env, // 显示正确的环境
			"go_version": runtime.Version(),
		})
	})

	// 只需要负责解析参数，和返回 JSON
	r.POST("/:id/join", func(c *gin.Context) {
		// 1. 获取参数 (HTTP 层面全是字符串)
		matchIDStr := c.Param("id")

		// 2. 参数转换与校验 (Router 层的职责)
		// ParseUint(字符串, 进制, 位数)
		id, err := strconv.ParseUint(matchIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID format"})
			return
		}
		matchID := uint(id) // 转换为 uint 以匹配 Model 定义

		// 假设从中间件获取 userID (类型通常是 float64 或 uint，取决于 JWT 库，这里假设已转好)
		// 如果是用 context.Set 存入的，取出来通常需要断言
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDVal.(uint) // 类型断言

		// 3. 调用 Service (传入强类型的业务数据)
		svc := service.MatchService{}
		if err := svc.JoinMatch(matchID, userID); err != nil {
			// 根据错误类型返回不同的 HTTP 状态码（这是进阶做法，目前先统一下）
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 4. 返回结果
		c.JSON(http.StatusOK, gin.H{"message": "Joined successfully"})
	})
}
