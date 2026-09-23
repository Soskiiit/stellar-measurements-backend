package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"stellar-measurements-backend/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const DefaultImageURL = "/static/img/star_default_404.jpg"
const DefaultVideoURL = "/static/videos/star_default.mp4"

func populateDefaults(star *ds.Star) {
	if star.ImageURL == "" {
		star.ImageURL = DefaultImageURL
	}
	if star.VideoURL == "" {
		star.VideoURL = DefaultVideoURL
	}
	if star.Distance == 0 && star.Parallax > 0 {
		star.Distance = 1.0 / star.Parallax
	}
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	var star *ds.Star
	var err error

	if idStr == "" {
		star, err = h.Repository.GetFirstPublishedStar()
		if err != nil || star == nil {
			h.errorHandler(ctx, http.StatusNotFound, "Опубликованные звёзды не найдены")
			return
		}
	} else {
		id, parseErr := strconv.Atoi(idStr)
		if parseErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, "Неверный ID звезды")
			return
		}

		if ctx.Query("next") == "true" {
			nextID, nextErr := h.Repository.GetNextPublishedStarID(id)
			if nextErr != nil {
				h.errorHandler(ctx, http.StatusNotFound, "Опубликованные звёзды не найдены")
				return
			}
			ctx.Redirect(http.StatusFound, "/feed/"+strconv.Itoa(int(nextID)))
			return
		}

		star, err = h.Repository.GetStarByID(id)
		if err != nil || star == nil {
			h.errorHandler(ctx, http.StatusNotFound, "Звезда не найдена или удалена")
			return
		}
	}

	populateDefaults(star)

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"star": star,
	})
}

func (h *Handler) GetDraft(ctx *gin.Context) {
	const currentUserID = 1
	draft, err := h.Repository.GetDraftByCreator(currentUserID)
	if err != nil {
		logrus.Error("Ошибка получения черновика: ", err)
	}

	hasDraft := draft != nil
	if hasDraft {
		populateDefaults(draft)
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"hasDraft": hasDraft,
		"star":     draft,
	})
}

func (h *Handler) CreateDraft(ctx *gin.Context) {
	const currentUserID = 1
	name := strings.TrimSpace(ctx.PostForm("name"))
	if name == "" {
		name = "Новая звезда"
	}
	if len([]rune(name)) > 100 {
		h.errorHandler(ctx, http.StatusBadRequest, "Длина названия звезды не может превышать 100 символов")
		return
	}

	existingDraft, _ := h.Repository.GetDraftByCreator(currentUserID)
	if existingDraft == nil {
		_, err := h.Repository.CreateDraft(currentUserID, name)
		if err != nil {
			logrus.Error("Ошибка создания черновика: ", err)
			h.errorHandler(ctx, http.StatusInternalServerError, "Не удалось создать черновик")
			return
		}
	}

	ctx.Redirect(http.StatusFound, "/add")
}

func (h *Handler) PublishStar(ctx *gin.Context) {
	const currentUserID = 1

	starIDStr := ctx.PostForm("star_id")
	var starID uint
	if starIDStr != "" {
		id, err := strconv.Atoi(starIDStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, "Неверный ID звезды")
			return
		}
		starID = uint(id)
	} else {
		draft, err := h.Repository.GetDraftByCreator(currentUserID)
		if err != nil || draft == nil {
			h.errorHandler(ctx, http.StatusBadRequest, "Черновик для публикации не найден")
			return
		}
		starID = draft.ID
	}

	description := strings.TrimSpace(ctx.PostForm("description"))
	if len([]rune(description)) > 255 {
		h.errorHandler(ctx, http.StatusBadRequest, "Длина описания не может превышать 255 символов")
		return
	}

	parallaxStr := ctx.PostForm("parallax")
	distanceStr := ctx.PostForm("distance")

	parallax, err := strconv.ParseFloat(parallaxStr, 64)
	if err != nil || parallax < 0.001 || parallax > 9.999 {
		h.errorHandler(ctx, http.StatusBadRequest, "Годичный параллакс должен быть в диапазоне от 0.001 до 9.999")
		return
	}
	parallax = math.Round(parallax*1000) / 1000

	var distance float64
	if distanceStr != "" {
		d, parseErr := strconv.ParseFloat(distanceStr, 64)
		if parseErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, "Некорректное значение расстояния")
			return
		}
		distance = d
	}

	if distance == 0 && parallax > 0 {
		distance = 1.0 / parallax
	}
	distance = math.Round(distance*100) / 100

	if distance < 0.01 || distance > 999.99 {
		h.errorHandler(ctx, http.StatusBadRequest, "Расстояние должно быть в диапазоне от 0.01 до 999.99")
		return
	}

	err = h.Repository.PublishStar(starID, description, parallax, distance)
	if err != nil {
		logrus.Error("Ошибка публикации карточки: ", err)
		h.errorHandler(ctx, http.StatusInternalServerError, "Не удалось опубликовать звезду")
		return
	}

	ctx.Redirect(http.StatusFound, "/catalog")
}

func (h *Handler) GetCatalog(ctx *gin.Context) {
	distanceStr := ctx.Query("distance")
	var stars []ds.Star
	var err error

	if distanceStr == "" {
		stars, err = h.Repository.GetPublishedStars()
	} else {
		distance, parseErr := strconv.ParseFloat(distanceStr, 64)
		if parseErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, "Неверное значение фильтра расстояния")
			return
		}
		stars, err = h.Repository.GetStarsByDistance(distance)
	}

	if err != nil {
		logrus.Error(err)
	}

	for i := range stars {
		populateDefaults(&stars[i])
	}

	ctx.HTML(http.StatusOK, "tile.html", gin.H{
		"stars":    stars,
		"distance": distanceStr,
	})
}

func (h *Handler) DeleteStar(ctx *gin.Context) {
	starIDStr := ctx.PostForm("star_id")
	starID, err := strconv.Atoi(starIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Неверный ID звезды для удаления")
		return
	}

	err = h.Repository.DeleteStar(uint(starID))
	if err != nil {
		logrus.Error("Ошибка удаления звезды: ", err)
		h.errorHandler(ctx, http.StatusInternalServerError, "Ошибка при удалении звезды")
		return
	}

	ctx.Redirect(http.StatusFound, "/catalog")
}
