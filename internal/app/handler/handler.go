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
	router.NoRoute(func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/")
	})

	router.GET("/", h.GetFeed)
	router.GET("/parallax_stars_feed", h.GetFeed)
	router.GET("/parallax_stars_feed/:id", h.GetFeed)

	router.GET("/add_parallax_star", h.GetDraft)
	router.GET("/parallax_stars_catalog", h.GetCatalog)

	router.POST("/add_parallax_star", h.CreateDraft)
	router.POST("/publish_parallax_star", h.PublishStar)
	router.POST("/delete_parallax_star", h.DeleteStar)
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
