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
func SetupRouter(r *gin.Engine, matchSvc *service.MatchService, teamSvc *service.TeamService) {
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

	// 临时 Mock Auth Middleware (用于 Day 28 真 Auth 之前的开发测试)
	api.Use(func(c *gin.Context) {
		// 默认 mock userID 为 1
		c.Set("userID", uint(1))
		c.Next()
	})

	api.POST("/matches/:id/join", func(c *gin.Context) {
		matchID, ok := utils.GetParamID(c, "id")
		if !ok {
			return
		}

		userID, ok := utils.GetUserID(c)
		if !ok {
			return
		}

		if err := matchSvc.JoinMatch(c.Request.Context(), matchID, userID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Joined successfully", "data": nil})
	})

	// 比赛模块路由 (Day 8: Match Creation)
	matches := api.Group("/matches")
	{
		matches.POST("/batch", func(c *gin.Context) {
			// 1. 定义接受请求的结构体并增加 validator 校验
			var req struct {
				TeamID     uint                    `json:"team_id" binding:"required"`
				VenueID    uint                    `json:"venue_id" binding:"required"`
				CommonInfo service.MatchCommonInfo `json:"common_info" binding:"required"`
				Schedules  []service.MatchSchedule `json:"schedules" binding:"required,min=1"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters", "data": err.Error()})
				return
			}

			// 2. 调用业务逻辑
			createdMatches, err := matchSvc.CreateMatchBatch(c.Request.Context(), req.TeamID, req.VenueID, req.CommonInfo, req.Schedules)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to create matches", "data": err.Error()})
				return
			}

			// 3. 响应
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": createdMatches})
		})
	}

	// 球队模块路由 (Day 4)
	teams := api.Group("/teams")
	{
		// 创建球队 POST /api/teams
		teams.POST("", func(c *gin.Context) {
			// 1. 定义接受请求的结构体并增加 validator 校验 tags
			var req struct {
				Name string `json:"name" binding:"required,min=2,max=20"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters", "data": err.Error()})
				return
			}

			// 2. 获取当前用户 ID
			userID, ok := utils.GetUserID(c)
			if !ok {
				return
			}

			// 3. 调用 Service
			team, err := teamSvc.CreateTeam(c.Request.Context(), userID, req.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to create team", "data": err.Error()})
				return
			}

			// 4. 返回标准 JSON
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": team})
		})

		// 查球队详情 GET /api/teams/:id
		teams.GET("/:id", func(c *gin.Context) {
			teamID, ok := utils.GetParamID(c, "id")
			if !ok {
				return
			}

			team, err := teamSvc.GetTeam(c.Request.Context(), teamID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Team not found", "data": nil})
				return
			}

			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": team})
		})
	}
}
