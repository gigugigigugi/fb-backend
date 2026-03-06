package main

import (
	"fmt"
	"football-backend/common/config"
	"football-backend/common/database"
	"football-backend/common/logger"
	"football-backend/common/notification"
	verificationcode "football-backend/common/verification"
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
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		slog.Error("Failed to load timezone", slog.String("error", err.Error()))
		os.Exit(1)
	}
	time.Local = loc

	config.Load()
	logger.Init(config.App.Env)
	db := database.Init(config.App.DB.DSN)

	userRepo := postgres.NewUserRepository(db)
	teamRepo := postgres.NewTeamRepository(db)
	matchRepo := postgres.NewMatchRepository(db)
	bookingRepo := postgres.NewBookingRepository(db)
	verificationRepo := postgres.NewVerificationRepository(db)

	notifyDispatcher := notification.NewDispatcher(256)
	notifyDispatcher.RegisterNotifier(notification.NewEmailNotifier())
	notifyDispatcher.RegisterNotifier(notification.NewSMSNotifier())

	codeProvider, err := verificationcode.NewCodeProviderFromConfig(config.App.Verification)
	if err != nil {
		slog.Error("Failed to init verification code provider", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Verification provider initialized", slog.String("mode", codeProvider.Mode()))

	matchSvc := service.NewMatchService(matchRepo, bookingRepo, teamRepo, userRepo, notifyDispatcher)
	teamSvc := service.NewTeamService(teamRepo)
	authSvc := service.NewAuthService(userRepo, verificationRepo, codeProvider)
	userSvc := service.NewUserService(userRepo)

	if config.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

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
