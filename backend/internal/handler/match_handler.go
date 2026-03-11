package handler

import (
	"errors"
	"football-backend/common/utils"
	"football-backend/internal/repository"
	"football-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MatchHandler 负责比赛相关 HTTP 接口。
type MatchHandler struct {
	matchSvc *service.MatchService
}

// NewMatchHandler 创建比赛处理器。
func NewMatchHandler(matchSvc *service.MatchService) *MatchHandler {
	return &MatchHandler{matchSvc: matchSvc}
}

// JoinMatch 报名指定比赛。
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

// GetMatches 按查询参数分页获取比赛列表。
func (h *MatchHandler) GetMatches(c *gin.Context) {
	var query struct {
		City      string `form:"city"`
		Status    string `form:"status"`
		Format    int    `form:"format"`
		StartTime string `form:"start_time"`                              // 起始时间过滤条件。
		Page      int    `form:"page" binding:"omitempty,min=1"`          // 页码（从 1 开始）。
		Limit     int    `form:"limit" binding:"omitempty,min=1,max=100"` // 每页数量。
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid query parameters: " + err.Error()})
		return
	}

	filter := repository.MatchFilter{
		City:   query.City,
		Status: query.Status,
		Format: query.Format,
	}

	// 仅在传入 start_time 时尝试解析，解析失败时忽略该过滤条件。
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

// CreateBatch 批量创建比赛。
func (h *MatchHandler) CreateBatch(c *gin.Context) {
	var req struct {
		TeamID     uint                    `json:"team_id" binding:"required"`         // 球队 ID。
		VenueID    uint                    `json:"venue_id" binding:"required"`        // 场地 ID。
		CommonInfo service.MatchCommonInfo `json:"common_info" binding:"required"`     // 公共比赛配置。
		Schedules  []service.MatchSchedule `json:"schedules" binding:"required,min=1"` // 比赛时间列表。
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters", "data": err.Error()})
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	createdMatches, err := h.matchSvc.CreateMatchBatch(c.Request.Context(), userID, req.TeamID, req.VenueID, req.CommonInfo, req.Schedules)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to create matches", "data": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": createdMatches})
}

// GetMatchDetails 获取比赛详情（比赛信息、阵容、评论、当前用户状态）。
func (h *MatchHandler) GetMatchDetails(c *gin.Context) {
	matchID, ok := utils.GetParamID(c, "id")
	if !ok {
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	detail, err := h.matchSvc.GetMatchDetails(c.Request.Context(), matchID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": detail})
}

// SettleMatch 处理赛后结算，更新报名记录的支付状态。
func (h *MatchHandler) SettleMatch(c *gin.Context) {
	matchID, ok := utils.GetParamID(c, "id")
	if !ok {
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	var req struct {
		PaymentStatus string `json:"payment_status" binding:"required"` // 目标支付状态：UNPAID/PAID/REFUNDED。
		BookingIDs    []uint `json:"booking_ids"`                       // 可选：仅结算指定报名 ID。
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters", "data": err.Error()})
		return
	}

	updatedCount, err := h.matchSvc.SettleMatch(c.Request.Context(), matchID, userID, req.PaymentStatus, req.BookingIDs)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPaymentStatus):
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		case errors.Is(err, service.ErrMatchManageForbidden):
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": err.Error(), "data": nil})
		case errors.Is(err, service.ErrMatchNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": err.Error(), "data": nil})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to settle match", "data": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"updated_count": updatedCount,
		},
	})
}

// AssignSubTeams 处理赛后分队结果写入。
func (h *MatchHandler) AssignSubTeams(c *gin.Context) {
	matchID, ok := utils.GetParamID(c, "id")
	if !ok {
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	var req struct {
		Assignments []service.SubTeamAssignment `json:"assignments" binding:"required,min=1"` // 分队结果列表。
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters", "data": err.Error()})
		return
	}

	if err := h.matchSvc.AssignMatchSubTeams(c.Request.Context(), matchID, userID, req.Assignments); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidSubTeamAssignments):
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		case errors.Is(err, service.ErrMatchManageForbidden):
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": err.Error(), "data": nil})
		case errors.Is(err, service.ErrMatchNotFound):
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": err.Error(), "data": nil})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to assign subteams", "data": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": nil})
}
