package main

import (
	"fmt"
	"football-backend/common/config"
	"football-backend/common/database"
	"football-backend/common/logger"
	"football-backend/internal/middleware"
	"football-backend/internal/repository/postgres"
	"football-backend/internal/router"
	"football-backend/internal/service"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func main() {
	// 初始化时区设置，确保整个应用基于日本时间或者强制 UTC
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		slog.Error("Failed to load timezone", slog.String("error", err.Error()))
		os.Exit(1)
	}
	time.Local = loc

	// 在程序启动时，首先显式加载所有配置
	config.Load()
	// 1. 基于加载的配置初始化日志
	logger.Init(config.App.Env)
	// 初始化数据库连接
	_ = database.Init(config.App.DB.DSN)
	db := database.DB // 获取底层的 *gorm.DB

	// 初始化领域层仓储 (Domain Repositories 具体实现)
	userRepo := postgres.NewUserRepository(db)
	teamRepo := postgres.NewTeamRepository(db)
	matchRepo := postgres.NewMatchRepository(db)
	bookingRepo := postgres.NewBookingRepository(db)

	// 初始化业务服务 (单例模式, 依赖精准注入)
	matchSvc := service.NewMatchService(matchRepo, bookingRepo, teamRepo)
	teamSvc := service.NewTeamService(teamRepo)
	authSvc := service.NewAuthService(userRepo)
	// 2. 基于加载的配置设置 Gin 框架的运行模式
	if config.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 注册 Validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 这里未来可以注册自定义的 tag 校验器比如: v.RegisterValidation("xxx", xxxFunc)
		_ = v
	}

	// 创建 Gin 引擎
	r := gin.New()

	// 注册中间件
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	// 设置路由
	router.SetupRouter(r, matchSvc, teamSvc, authSvc)

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
