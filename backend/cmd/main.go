package main

import (
	"fmt"
	"football-backend/common/config"
	"football-backend/common/database"
	"football-backend/common/logger"
	"football-backend/common/notification"
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

// main 是服务启动入口，负责依赖装配与 HTTP Server 启动。
func main() {
	// 统一时区，避免时间比较与入库出现偏差。
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		slog.Error("Failed to load timezone", slog.String("error", err.Error()))
		os.Exit(1)
	}
	time.Local = loc

	config.Load()
	logger.Init(config.App.Env)
	db := database.Init(config.App.DB.DSN)

	// 初始化数据仓储。
	userRepo := postgres.NewUserRepository(db)
	teamRepo := postgres.NewTeamRepository(db)
	matchRepo := postgres.NewMatchRepository(db)
	bookingRepo := postgres.NewBookingRepository(db)
	verificationRepo := postgres.NewVerificationRepository(db)

	// 初始化通知分发器（用于候补提醒）。
	notifyDispatcher := notification.NewDispatcher(256)
	notifyDispatcher.RegisterNotifier(notification.NewEmailNotifier())
	notifyDispatcher.RegisterNotifier(notification.NewSMSNotifier())

	// 初始化业务服务。
	matchSvc := service.NewMatchService(matchRepo, bookingRepo, teamRepo, userRepo, notifyDispatcher)
	teamSvc := service.NewTeamService(teamRepo)
	authSvc := service.NewAuthService(userRepo, verificationRepo)
	userSvc := service.NewUserService(userRepo)

	if config.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 预留自定义校验器注册位置。
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v
	}

	r := gin.New()
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	router.SetupRouter(r, matchSvc, teamSvc, authSvc, userSvc)

	slog.Debug("Server starting",
		slog.String("port", config.App.Port),
		slog.String("env", config.App.Env),
	)

	addr := fmt.Sprintf(":%s", config.App.Port)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("Failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
