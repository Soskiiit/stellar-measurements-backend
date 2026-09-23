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

// RegisterHandler регистрирует роуты согласно требованиям ЛР2:
// Всего 6 HTTP методов:
// - 3 GET (feed, add/draft, catalog)
// - 1 POST добавления новой карточки через ORM
// - 1 POST публикации карточки через ORM
// - 1 POST логического удаления услуги через SQL курсор (SQL UPDATE, без ORM)
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// 3 GET метода
	router.GET("/", h.GetFeed)
	router.GET("/feed/:id", h.GetFeed)
	router.GET("/add", h.GetDraft)
	router.GET("/catalog", h.GetCatalog)

	// 3 POST метода
	router.POST("/stars", h.CreateDraft)         // 1. Добавление новой карточки через ORM
	router.POST("/stars/publish", h.PublishStar) // 2. Публикация карточки через ORM
	router.POST("/stars/delete", h.DeleteStar)   // 3. Логическое удаление услуги через SQL курсор (без ORM)
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
