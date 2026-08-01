package middleware

import (
	"errors"
	"strings"

	"forum/internal/apperror"
	"forum/internal/auth"
	"forum/internal/model"
	"forum/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	ContextUserID   = "user_id"
	ContextUsername = "username"
	ContextRole     = "role"
)

func RequiredAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			response.Abort(c, apperror.ErrUnauthorized)
			return
		}
		claims, err := tokens.Parse(raw)
		if err != nil {
			response.Abort(c, apperror.ErrUnauthorized)
			return
		}
		setClaims(c, claims)
		c.Next()
	}
}

func OptionalAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := bearerToken(c.GetHeader("Authorization"))
		if ok {
			if claims, err := tokens.Parse(raw); err == nil {
				setClaims(c, claims)
			}
		}
		c.Next()
	}
}

func AdminOnly(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := UserID(c)
		if !ok {
			response.Abort(c, apperror.ErrUnauthorized)
			return
		}
		var user model.User
		if err := db.WithContext(c.Request.Context()).Select("id", "role").First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Abort(c, apperror.ErrUnauthorized)
			} else {
				response.Abort(c, err)
			}
			return
		}
		if user.Role != model.RoleAdmin {
			response.Abort(c, apperror.New(403, 403, "admin only"))
			return
		}
		c.Next()
	}
}

func UserID(c *gin.Context) (uint64, bool) {
	value, ok := c.Get(ContextUserID)
	if !ok {
		return 0, false
	}
	id, ok := value.(uint64)
	return id, ok
}

func bearerToken(value string) (string, bool) {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func setClaims(c *gin.Context, claims *auth.Claims) {
	c.Set(ContextUserID, claims.UserID)
	c.Set(ContextUsername, claims.Username)
	c.Set(ContextRole, claims.Role)
}
