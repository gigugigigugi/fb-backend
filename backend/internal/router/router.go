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

// SetupRouter configures the routes for the application on the given Gin engine.
func SetupRouter(r *gin.Engine, matchSvc *service.MatchService, teamSvc *service.TeamService, authSvc *service.AuthService) {
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

	// API 路由组
	api := r.Group("/api")

	// 初始化 HTTP Handlers (Controllers)
	authHandler := handler.NewAuthHandler(authSvc)
	matchHandler := handler.NewMatchHandler(matchSvc)
	teamHandler := handler.NewTeamHandler(teamSvc)
	bookingHandler := handler.NewBookingHandler(matchSvc)

	// --- 开放路由区 (不需要 Token 也能访问) ---
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/google", authHandler.GoogleLogin)
	}

	// 开放比赛查询路由 (游客可访问)
	publicMatches := api.Group("/matches")
	{
		publicMatches.GET("", matchHandler.GetMatches)
	}

	// --- 真实 Auth Middleware (带后门开关) ---
	api.Use(func(c *gin.Context) {
		// [配置项] 开发模式下是否关闭验证的后门。可以在配置项读取比如 config.App.AuthBypass
		authBypass := config.App.AuthBypass

		if authBypass {
			// 开发阶段后门，默认塞个 UserID 1 进去以满足调测
			c.Set("userID", uint(1))
			c.Next()
			return
		}

		// 1. 获取 Authorization Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Unauthorized: missing or invalid bearer token"})
			return
		}

		tokenString := authHeader[7:]

		// 2. 解密/验签 JWT
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Unauthorized: " + err.Error()})
			return
		}

		// 3. 将真实解开的 UserID 塞进 Context 供后面的 Handler 使用！
		c.Set("userID", claims.UserID)
		c.Next()
	})

	// --- 独立零散路由 ---
	api.POST("/matches/:id/join", matchHandler.JoinMatch)

	// --- 比赛模块路由 (Day 8 & Day 9) ---
	matches := api.Group("/matches")
	{
		matches.POST("/batch", matchHandler.CreateBatch)
	}

	// --- 订单/报名记录模块路由 ---
	bookings := api.Group("/bookings")
	{
		bookings.POST("/:id/cancel", bookingHandler.CancelBooking)
	}

	// --- 球队模块路由 (Day 4) ---
	teams := api.Group("/teams")
	{
		teams.POST("", teamHandler.CreateTeam)
		teams.GET("/:id", teamHandler.GetTeam)
	}
}
