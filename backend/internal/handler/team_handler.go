package handler

import (
	"football-backend/common/utils"
	"football-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamSvc *service.TeamService
}

func NewTeamHandler(teamSvc *service.TeamService) *TeamHandler {
	return &TeamHandler{teamSvc: teamSvc}
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
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
	team, err := h.teamSvc.CreateTeam(c.Request.Context(), userID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to create team", "data": err.Error()})
		return
	}

	// 4. 返回标准 JSON
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": team})
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamID, ok := utils.GetParamID(c, "id")
	if !ok {
		return
	}

	team, err := h.teamSvc.GetTeam(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Team not found", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": team})
}
