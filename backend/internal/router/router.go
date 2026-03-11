package router

import (
	"football-backend/common/config"
	"football-backend/common/utils"
	"football-backend/internal/handler"
	"football-backend/internal/service"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

// SetupRouter 注册应用全部路由。
func SetupRouter(r *gin.Engine, matchSvc *service.MatchService, teamSvc *service.TeamService, authSvc *service.AuthService, userSvc *service.UserService, venueSvc *service.VenueService) {
	env := config.App.Env

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "WELCOME",
			"env":        env,
			"go_version": runtime.Version(),
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "pong",
			"env":        env,
			"go_version": runtime.Version(),
		})
	})

	// 统一 API 前缀策略：全部挂在 /api/v1 下。
	api := r.Group("/api/v1")

	authHandler := handler.NewAuthHandler(authSvc)
	matchHandler := handler.NewMatchHandler(matchSvc)
	teamHandler := handler.NewTeamHandler(teamSvc)
	bookingHandler := handler.NewBookingHandler(matchSvc)
	userHandler := handler.NewUserHandler(userSvc)
	venueHandler := handler.NewVenueHandler(venueSvc)

	// 公开认证接口。
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/google", authHandler.GoogleLogin)
	}

	// 公开比赛查询接口。
	publicMatches := api.Group("/matches")
	{
		publicMatches.GET("", matchHandler.GetMatches)
	}

	publicVenues := api.Group("/venues")
	{
		publicVenues.GET("/regions", venueHandler.GetRegions)
		publicVenues.GET("/map", venueHandler.GetMap)
	}

	// 鉴权中间件：支持配置开关的开发绕过。
	api.Use(func(c *gin.Context) {
		authBypass := config.App.AuthBypass
		if authBypass {
			// 开发模式下给定固定 userID，便于联调。
			c.Set("userID", uint(1))
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Unauthorized: missing or invalid bearer token"})
			return
		}

		tokenString := authHeader[7:]
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Unauthorized: " + err.Error()})
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	})

	api.POST("/matches/:id/join", matchHandler.JoinMatch)

	matches := api.Group("/matches")
	{
		matches.POST("/batch", matchHandler.CreateBatch)
		matches.GET("/:id/details", matchHandler.GetMatchDetails)
		matches.POST("/:id/settlement", matchHandler.SettleMatch)
		matches.POST("/:id/subteams", matchHandler.AssignSubTeams)
	}

	bookings := api.Group("/bookings")
	{
		bookings.POST("/:id/cancel", bookingHandler.CancelBooking)
	}

	users := api.Group("/users")
	{
		users.GET("/me", userHandler.GetMe)
		users.PUT("/me", userHandler.UpdateMe)
		users.GET("/me/bookings", bookingHandler.GetUserBookings)
		users.POST("/me/verify/email/send", authHandler.SendEmailVerificationCode)
		users.POST("/me/verify/email/confirm", authHandler.VerifyEmailCode)
		users.POST("/me/verify/phone/send", authHandler.SendPhoneVerificationCode)
		users.POST("/me/verify/phone/confirm", authHandler.VerifyPhoneCode)
	}

	teams := api.Group("/teams")
	{
		teams.POST("", teamHandler.CreateTeam)
		teams.GET("/:id", teamHandler.GetTeam)
	}
}
