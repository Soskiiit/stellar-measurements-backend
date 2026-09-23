package handler

import (
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
	router.GET("/", h.GetFeed)
	router.GET("/feed/:id", h.GetFeed)
	router.GET("/add", h.GetDraft)
	router.GET("/catalog", h.GetCatalog)

	router.POST("/stars", h.CreateDraft)
	router.POST("/stars/publish", h.PublishStar)
	router.POST("/stars/delete", h.DeleteStar)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, errMsg string) {
	logrus.Error(errMsg)
	ctx.HTML(errorStatusCode, "error.html", gin.H{
		"error": errMsg,
	})
}
