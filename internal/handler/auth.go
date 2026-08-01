package handler

import (
	"net/http"
	"regexp"

	"forum/internal/apperror"
	"forum/internal/config"
	"forum/internal/response"
	"forum/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service              *service.AuthService
	adminRegistrationKey string
	releaseMode          bool
}

func NewAuthHandler(service *service.AuthService, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		service:              service,
		adminRegistrationKey: cfg.Auth.AdminRegistrationKey,
		releaseMode:          cfg.Server.Mode == "release",
	}
}

var usernamePattern = regexp.MustCompile(`^[0-9]+$`)

type registerRequest struct {
	Username string `json:"username" binding:"required,min=1,max=32"`
	Name     string `json:"name" binding:"required,min=1,max=32"`
	Password string `json:"password" binding:"required,min=8,max=16"`
	Role     string `json:"role" binding:"required,oneof=student admin"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var request registerRequest
	if err := c.ShouldBindJSON(&request); err != nil || !usernamePattern.MatchString(request.Username) {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	if request.Role == "admin" && !h.adminRegistrationAllowed(c) {
		response.Error(c, apperror.New(http.StatusForbidden, http.StatusForbidden, "管理员账户注册已受限"))
		return
	}
	user, err := h.service.Register(service.RegisterInput{
		Username: request.Username,
		Name:     request.Name,
		Password: request.Password,
		Role:     request.Role,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, user)
}

func (h *AuthHandler) adminRegistrationAllowed(c *gin.Context) bool {
	if h.adminRegistrationKey != "" {
		return c.GetHeader("X-Admin-Registration-Key") == h.adminRegistrationKey
	}
	return !h.releaseMode
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	result, err := h.service.Login(request.Username, request.Password)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
