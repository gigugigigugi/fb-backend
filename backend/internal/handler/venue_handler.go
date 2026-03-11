package handler

import (
	"football-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VenueHandler 负责场地发现接口。
type VenueHandler struct {
	venueSvc *service.VenueService
}

// NewVenueHandler 创建场地处理器。
func NewVenueHandler(venueSvc *service.VenueService) *VenueHandler {
	return &VenueHandler{venueSvc: venueSvc}
}

// GetRegions 返回行政区树（prefecture -> city）。
func (h *VenueHandler) GetRegions(c *gin.Context) {
	regions, err := h.venueSvc.GetRegions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to get venue regions", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": regions})
}

// GetMap 返回地图模式的场地点位列表。
func (h *VenueHandler) GetMap(c *gin.Context) {
	var query struct {
		Prefecture string `form:"prefecture"`                               // 可选：一级行政区过滤。
		City       string `form:"city"`                                     // 可选：城市过滤。
		Limit      int    `form:"limit" binding:"omitempty,min=1,max=1000"` // 可选：返回数量上限。
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid query parameters: " + err.Error(), "data": nil})
		return
	}

	items, err := h.venueSvc.GetMapVenues(c.Request.Context(), query.Prefecture, query.City, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to get venue map data", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"items": items,
			"total": len(items),
		},
	})
}
