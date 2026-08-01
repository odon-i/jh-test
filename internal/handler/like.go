package handler

import (
	"forum/internal/apperror"
	"forum/internal/middleware"
	"forum/internal/response"
	"forum/internal/service"
	"github.com/gin-gonic/gin"
)

const maxLikeStatusIDs = 100

type SocialHandler struct {
	service *service.SocialService
}

func NewSocialHandler(service *service.SocialService) *SocialHandler {
	return &SocialHandler{service: service}
}

func (h *SocialHandler) ToggleLike(c *gin.Context) {
	postID, ok := parseID(c.Param("post_id"))
	if !ok {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	userID, _ := middleware.UserID(c)
	result, err := h.service.ToggleLike(c.Request.Context(), userID, postID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

type likeStatusesRequest struct {
	PostIDs []uint64 `json:"post_ids" binding:"required,min=1,max=100,dive,gt=0"`
}

func (h *SocialHandler) LikeStatuses(c *gin.Context) {
	var request likeStatusesRequest
	if err := c.ShouldBindJSON(&request); err != nil || len(request.PostIDs) > maxLikeStatusIDs {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	userID, _ := middleware.UserID(c)
	result, err := h.service.LikeStatuses(c.Request.Context(), userID, request.PostIDs)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

type commentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

func (h *SocialHandler) AddComment(c *gin.Context) {
	postID, ok := parseID(c.Param("post_id"))
	if !ok {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	var request commentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	userID, _ := middleware.UserID(c)
	comment, err := h.service.AddComment(c.Request.Context(), userID, postID, request.Content)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, comment)
}
