package handler

import (
	"math"
	"strconv"

	"forum/internal/apperror"
	"forum/internal/middleware"
	"forum/internal/response"
	"forum/internal/service"
	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	service *service.PostService
}

func NewPostHandler(service *service.PostService) *PostHandler {
	return &PostHandler{service: service}
}

type createPostRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

func (h *PostHandler) Create(c *gin.Context) {
	var request createPostRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	userID, _ := middleware.UserID(c)
	post, err := h.service.Create(userID, request.Content)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, post)
}

func (h *PostHandler) List(c *gin.Context) {
	page, ok := parsePositiveQuery(c, "page", 1, 0)
	if !ok {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	pageSize, ok := parsePositiveQuery(c, "page_size", 20, 100)
	if !ok {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	if page > math.MaxInt/pageSize+1 {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	sort := c.DefaultQuery("sort", "latest")
	if sort != "latest" && sort != "hot" {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	result, err := h.service.List(page, pageSize, sort)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *PostHandler) Detail(c *gin.Context) {
	postID, ok := parseID(c.Param("post_id"))
	if !ok {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	post, err := h.service.Detail(postID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, post)
}

func (h *PostHandler) DeleteOwn(c *gin.Context) {
	postID, ok := parseID(c.Param("post_id"))
	if !ok {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	userID, _ := middleware.UserID(c)
	if err := h.service.DeleteOwn(userID, postID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *PostHandler) DeleteAny(c *gin.Context) {
	postID, ok := parseID(c.Param("post_id"))
	if !ok {
		response.Error(c, apperror.ErrBadRequest)
		return
	}
	if err := h.service.DeleteAny(postID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func parseID(value string) (uint64, bool) {
	id, err := strconv.ParseInt(value, 10, 64)
	return uint64(id), err == nil && id > 0
}

func parsePositiveQuery(c *gin.Context, key string, fallback, max int) (int, bool) {
	value := c.Query(key)
	if value == "" {
		return fallback, true
	}
	number, err := strconv.Atoi(value)
	if err != nil || number < 1 || (max > 0 && number > max) {
		return 0, false
	}
	return number, true
}
