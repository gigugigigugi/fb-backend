package handler

import (
	"football-backend/common/utils"
	"football-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	matchSvc *service.MatchService
}

func NewBookingHandler(matchSvc *service.MatchService) *BookingHandler {
	return &BookingHandler{matchSvc: matchSvc}
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingID, ok := utils.GetParamID(c, "id")
	if !ok {
		return
	}

	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	if err := h.matchSvc.CancelBooking(c.Request.Context(), bookingID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Booking canceled successfully", "data": nil})
}
