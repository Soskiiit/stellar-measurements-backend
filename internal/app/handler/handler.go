package handler

import (
	"net/http"
	"strconv"

	"stellar-measurements-backend/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Неверный ID"})
		return
	}

	if ctx.Query("next") == "true" {
		stars, _ := h.Repository.GetPublishedStars()
		for i, s := range stars {
			if s.ID == id && i+1 < len(stars) {
				ctx.Redirect(http.StatusFound, "/feed/"+strconv.Itoa(stars[i+1].ID))
				return
			}
		}
		ctx.Redirect(http.StatusFound, "/feed/"+idStr)
		return
	}

	star, err := h.Repository.GetStarByID(id)
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Звезда не найдена"})
		return
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"star": star,
	})
}

func (h *Handler) GetDraft(ctx *gin.Context) {
	star, err := h.Repository.GetDraft()
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Черновик не найден"})
		return
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"star": star,
	})
}

func (h *Handler) GetTileList(ctx *gin.Context) {
	parallaxStr := ctx.Query("parallax")

	var stars []repository.Star
	var err error

	if parallaxStr == "" {
		stars, err = h.Repository.GetPublishedStars()
	} else {
		parallax, parseErr := strconv.ParseFloat(parallaxStr, 64)
		if parseErr != nil {
			logrus.Error(parseErr)
			ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Неверное значение параллакса"})
			return
		}
		stars, err = h.Repository.GetStarsByParallax(parallax)
	}

	if err != nil {
		logrus.Error(err)
	}

	for i := range stars {
		stars[i].Distance = 1.0 / stars[i].Parallax
	}

	ctx.HTML(http.StatusOK, "tile.html", gin.H{
		"stars":    stars,
		"parallax": parallaxStr,
	})
}
