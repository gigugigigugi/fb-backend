package handler

import (
	"football-backend/common/utils"
	"football-backend/internal/repository"
	"football-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MatchHandler struct {
	matchSvc *service.MatchService
}

func NewMatchHandler(matchSvc *service.MatchService) *MatchHandler {
	return &MatchHandler{matchSvc: matchSvc}
}

func (h *MatchHandler) JoinMatch(c *gin.Context) {
	matchID, ok := utils.GetParamID(c, "id")
	if !ok {
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	if err := h.matchSvc.JoinMatch(c.Request.Context(), matchID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Joined successfully", "data": nil})
}

func (h *MatchHandler) GetMatches(c *gin.Context) {
	var query struct {
		City      string `form:"city"`
		Status    string `form:"status"`
		Format    int    `form:"format"`
		StartTime string `form:"start_time"` // 日期格式由外部手动解析或者用自定义 unmarshaler
		Page      int    `form:"page" binding:"omitempty,min=1"`
		Limit     int    `form:"limit" binding:"omitempty,min=1,max=100"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid query parameters: " + err.Error()})
		return
	}

	// 解析查询条件装载进防腐层的 Filter
	filter := repository.MatchFilter{
		City:   query.City,
		Status: query.Status,
		Format: query.Format,
	}

	// 如果前端传了 start_time 比如 2026-03-05T19:00:00+09:00, 则解析它
	if query.StartTime != "" {
		parsedTime, err := utils.ParseTime(query.StartTime)
		if err == nil {
			filter.StartTime = parsedTime
		}
	}

	matchesList, total, err := h.matchSvc.SearchMatches(c.Request.Context(), filter, query.Page, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to fetch matches", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"items": matchesList,
			"total": total,
			"page":  query.Page,
			"limit": query.Limit,
		},
	})
}

func (h *MatchHandler) CreateBatch(c *gin.Context) {
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

	// 2. 获取当前发出批量建赛请求的队长/管理员 ID
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	// 3. 调用业务逻辑
	createdMatches, err := h.matchSvc.CreateMatchBatch(c.Request.Context(), userID, req.TeamID, req.VenueID, req.CommonInfo, req.Schedules)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to create matches", "data": err.Error()})
		return
	}

	// 3. 响应
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": createdMatches})
}
