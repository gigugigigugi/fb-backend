package main

import (
	"fmt"
	"football-backend/common/config" // 导入新的 config 包
	"football-backend/common/logger"
	"football-backend/internal/middleware"
	"football-backend/internal/router"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// init 函数会在 main 函数执行前自动运行
func init() {
	// 在程序启动时，首先加载所有配置
	config.Load()
}

func main() {
	// 1. 基于加载的配置初始化日志
	logger.Init(config.App.Env)

	// 2. 基于加载的配置设置 Gin 框架的运行模式
	if config.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建 Gin 引擎
	r := gin.New()

	// 注册中间件
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	// 设置路由
	router.SetupRouter(r)

	// 3. 从加载的配置中获取端口号

	slog.Info("Server starting",
		slog.String("port", config.App.Port),
		slog.String("env", config.App.Env),
	)

	// 启动服务
	addr := fmt.Sprintf(":%s", config.App.Port)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("Failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
