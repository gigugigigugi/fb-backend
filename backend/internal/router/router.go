package router

import (
	"football-backend/common/config" // 导入 config 包
	"football-backend/common/utils"
	"football-backend/internal/service"
	"net/http"
	"runtime"

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

	r.POST("/:id/join", func(c *gin.Context) {
		// 1. [简化] 获取并校验 URL 参数 ID
		matchID, ok := utils.GetParamID(c, "id")
		if !ok {
			return // utils 内部已经处理了 c.JSON 响应
		}

		// 2. [简化] 安全获取 UserID
		userID, ok := utils.GetUserID(c)
		if !ok {
			return // utils 内部已经处理了 401 或 500 响应
		}

		// 3. 调用业务逻辑
		svc := service.MatchService{}
		if err := svc.JoinMatch(matchID, userID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Joined successfully"})
	})
}
