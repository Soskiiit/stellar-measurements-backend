package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"stellar-measurements-backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	// ==================== REST API МАРШРУТЫ (/api) ====================
	api := router.Group("/api")
	{
		// Домен услуги (звезды):
		api.GET("/parallax_stars", h.GetParallaxStarsAPI)
		api.GET("/parallax_stars/feed", h.GetParallaxStarFeedAPI)
		api.GET("/parallax_stars/draft", h.GetDraftAPI)
		api.POST("/parallax_stars", h.AddParallaxStarAPI)
		api.PUT("/parallax_stars/:id/publish", h.PublishParallaxStarAPI)
		api.DELETE("/parallax_stars/:id", h.DeleteParallaxStarAPI)
		api.POST("/parallax_stars/:id/like", h.LikeParallaxStarAPI)

		// Домен пользователя:
		api.POST("/users/register", h.RegisterUserAPI)
		api.POST("/users/login", h.LoginUserAPI)
		api.POST("/users/logout", h.LogoutUserAPI)
	}

	// Fallback для несуществующих маршрутов
	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Маршрут не найден",
		})
	})
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, errMsg string) {
	logrus.Error(errMsg)
	ctx.JSON(errorStatusCode, gin.H{
		"error": errMsg,
	})
}
