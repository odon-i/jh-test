package router

import (
	"net/http"

	"forum/internal/auth"
	"forum/internal/config"
	"forum/internal/handler"
	"forum/internal/middleware"
	"forum/internal/response"
	"forum/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config config.Config
	DB     *gorm.DB
}

func New(deps Dependencies) *gin.Engine {
	gin.SetMode(deps.Config.Server.Mode)
	gin.EnableJsonDecoderDisallowUnknownFields()
	engine := gin.New()

	tokens := auth.NewTokenManager(deps.Config.JWT)
	authHandler := handler.NewAuthHandler(service.NewAuthService(deps.DB, tokens), deps.Config)
	postService := service.NewPostService(deps.DB)
	postHandler := handler.NewPostHandler(postService)
	socialHandler := handler.NewSocialHandler(service.NewSocialService(deps.DB))

	engine.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	v1 := engine.Group("/api/v1")
	{
		authRoutes := v1.Group("/auth")
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)

		protected := v1.Group("")
		protected.Use(middleware.RequiredAuth(tokens))
		{
			protected.POST("/posts", postHandler.Create)
			protected.GET("/posts", postHandler.List)
			protected.POST("/posts/likes", socialHandler.LikeStatuses)
			protected.GET("/posts/:post_id", postHandler.Detail)
			protected.DELETE("/posts/:post_id", postHandler.DeleteOwn)
			protected.POST("/posts/:post_id/like", socialHandler.ToggleLike)
			protected.POST("/posts/:post_id/comment", socialHandler.AddComment)

			admin := protected.Group("/admin")
			admin.Use(middleware.AdminOnly(deps.DB))
			admin.DELETE("/posts/:post_id", postHandler.DeleteAny)
		}
	}

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, response.Envelope{Code: http.StatusNotFound, Msg: "接口不存在", Data: nil})
	})
	return engine
}
