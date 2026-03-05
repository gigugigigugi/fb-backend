package handler

import (
	"football-backend/common/utils"
	"football-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register 接口处理器
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Nickname string `json:"nickname" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters: " + err.Error()})
		return
	}

	token, err := h.authSvc.RegisterEmail(c.Request.Context(), req.Email, req.Password, req.Nickname)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Registered successfully", "data": gin.H{"token": token}})
}

// Login 接口处理器
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters"})
		return
	}

	token, err := h.authSvc.LoginEmail(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Login success", "data": gin.H{"token": token}})
}

// GoogleLogin 处理器 (验证前端送来的 id_token)
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req struct {
		IDToken string `json:"id_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid IDToken"})
		return
	}

	// Google 官方解码客户端 (需要在 Config 里配 Google Client ID)
	// mock client_id 为空以跳过严格 audience 校验，仅做解码安全展示。正常上线需要配你的 web client id
	payload, err := idtoken.Validate(c.Request.Context(), req.IDToken, "")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Invalid Google Token: " + err.Error()})
		return
	}

	// 提取 Payload 中的关键资料
	googleID := payload.Subject
	email := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	token, err := h.authSvc.LoginGoogle(c.Request.Context(), googleID, email, name, picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Google Authorized", "data": gin.H{"token": token}})
}

func (h *AuthHandler) SendEmailVerificationCode(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	if err := h.authSvc.SendEmailVerificationCode(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Email verification code sent"})
}

func (h *AuthHandler) VerifyEmailCode(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	var req struct {
		Code string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters: " + err.Error()})
		return
	}

	if err := h.authSvc.VerifyEmailCode(c.Request.Context(), userID, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Email verified"})
}

func (h *AuthHandler) SendPhoneVerificationCode(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters: " + err.Error()})
		return
	}

	if err := h.authSvc.SendPhoneVerificationCode(c.Request.Context(), userID, req.Phone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Phone verification code sent"})
}

func (h *AuthHandler) VerifyPhoneCode(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters: " + err.Error()})
		return
	}

	if err := h.authSvc.VerifyPhoneCode(c.Request.Context(), userID, req.Phone, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Phone verified"})
}
