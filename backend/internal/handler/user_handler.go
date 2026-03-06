package handler

import (
	"football-backend/common/utils"
	"football-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserHandler 负责用户自身资料相关接口。
type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler 创建用户处理器。
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GetMe 获取当前登录用户资料。
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	user, err := h.userSvc.GetMe(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": user})
}

// UpdateMe 更新当前登录用户资料（昵称、头像）。
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	var req struct {
		Nickname *string `json:"nickname"` // 昵称，可选。
		Avatar   *string `json:"avatar"`   // 头像 URL，可选。
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters: " + err.Error()})
		return
	}

	user, err := h.userSvc.UpdateMe(c.Request.Context(), userID, service.UpdateMeInput{
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Profile updated", "data": user})
}
