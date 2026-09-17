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
	var id int
	if idStr == "" {
		stars, err := h.Repository.GetPublishedStars()
		if err != nil || len(stars) == 0 {
			logrus.Error("Звёзды не найдены")
			ctx.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Звёзды не найдены"})
			return
		}
		id = stars[0].ID
	} else {
		var err error
		id, err = strconv.Atoi(idStr)
		if err != nil {
			logrus.Error(err)
			ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Неверный ID"})
			return
		}
	}

	if ctx.Query("next") == "true" {
		stars, _ := h.Repository.GetPublishedStars()
		if len(stars) > 0 {
			minID := stars[0].ID
			for _, s := range stars {
				if s.ID < minID {
					minID = s.ID
				}
			}
			for i, s := range stars {
				if s.ID == id {
					if i+1 < len(stars) {
						ctx.Redirect(http.StatusFound, "/feed/"+strconv.Itoa(stars[i+1].ID))
					} else {
						ctx.Redirect(http.StatusFound, "/feed/"+strconv.Itoa(minID))
					}
					return
				}
			}
			ctx.Redirect(http.StatusFound, "/feed/"+strconv.Itoa(minID))
			return
		}
	}

	star, err := h.Repository.GetStarByID(id)
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Звезда не найдена"})
		return
	}

	if star.Parallax > 0 {
		star.Distance = 1.0 / star.Parallax
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
	distanceStr := ctx.Query("distance")

	var stars []repository.Star
	var err error

	if distanceStr == "" {
		stars, err = h.Repository.GetPublishedStars()
	} else {
		distance, parseErr := strconv.ParseFloat(distanceStr, 64)
		if parseErr != nil {
			logrus.Error(parseErr)
			ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Неверное значение расстояния"})
			return
		}
		stars, err = h.Repository.GetStarsByDistance(distance)
	}

	if err != nil {
		logrus.Error(err)
	}

	for i := range stars {
		stars[i].Distance = 1.0 / stars[i].Parallax
	}

	ctx.HTML(http.StatusOK, "tile.html", gin.H{
		"stars":    stars,
		"distance": distanceStr,
	})
}
