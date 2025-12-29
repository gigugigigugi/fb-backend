package router

import (
	"football-backend/common/config" // 导入 config 包
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
}
