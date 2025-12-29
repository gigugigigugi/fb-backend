package main

import (
	"fmt"
	"football-backend/common/config"
	"football-backend/common/database"
	"football-backend/common/logger"
	"football-backend/internal/middleware"
	"football-backend/internal/router"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 在程序启动时，首先显式加载所有配置
	config.Load()
	// 1. 基于加载的配置初始化日志
	logger.Init(config.App.Env)
	// 初始化数据库连接
	database.Init(config.App.DB.DSN)
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
	slog.Debug("Server starting",
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
