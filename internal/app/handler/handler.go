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

	// ==================== SSR ШАБЛОНЫ (ЛР1 и ЛР2) ====================
	router.GET("/", h.GetFeed)
	router.GET("/parallax_stars_feed", h.GetFeed)
	router.GET("/parallax_stars_feed/:id", h.GetFeed)

	router.GET("/add_parallax_star", h.GetDraft)
	router.GET("/parallax_stars_catalog", h.GetCatalog)

	router.POST("/add_parallax_star", h.CreateDraft)
	router.POST("/publish_parallax_star", h.PublishStar)
	router.POST("/delete_parallax_star", h.DeleteStar)

	// Fallback для несуществующих маршрутов
	router.NoRoute(func(ctx *gin.Context) {
		if len(ctx.Request.URL.Path) >= 4 && ctx.Request.URL.Path[:4] == "/api" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "API маршрут не найден",
			})
			return
		}
		ctx.Redirect(http.StatusFound, "/")
	})
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, errMsg string) {
	logrus.Error(errMsg)
	if errorStatusCode == http.StatusNotFound {
		if ctx.Request.URL.Path != "/" {
			ctx.Redirect(http.StatusFound, "/")
			return
		}
	}
	ctx.JSON(errorStatusCode, gin.H{
		"error": errMsg,
	})
}
