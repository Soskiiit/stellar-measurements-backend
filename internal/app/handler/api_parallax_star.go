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

// GetParallaxStarsAPI godoc
// @Summary Список опубликованных звезд каталога
// @Description Возвращает опубликованные звезды с поддержкой серверной фильтрации по расстоянию. Для авторизованных пользователей признак is_creator (1/0) указывает авторство текущего пользователя.
// @Tags parallax_stars
// @Produce json
// @Param distance query number false "Максимальное расстояние до звезды (пк)"
// @Success 200 {array} ds.ParallaxStarCatalogItemSerializer
// @Failure 400 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/parallax_stars [get]
func (h *Handler) GetParallaxStarsAPI(ctx *gin.Context) {
	currentUserID := GetContextUserID(ctx)
	distanceStr := ctx.Query("distance")

	var maxDistance *float64
	if distanceStr != "" {
		d, err := strconv.ParseFloat(distanceStr, 64)
		if err != nil || d < 0 {
			ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
				Error: "Некорректное значение фильтра distance (должно быть неотрицательным числом)",
			})
			return
		}
		maxDistance = &d
	}

	stars, err := h.Repository.GetPublishedParallaxStarsFiltered(maxDistance)
	if err != nil {
		logrus.Errorf("Ошибка получения каталога звезд: %v", err)
		ctx.JSON(http.StatusInternalServerError, ds.ErrorResponse{
			Error: "Не удалось загрузить список звезд",
		})
		return
	}

	result := make([]ds.ParallaxStarCatalogItemSerializer, 0, len(stars))
	for _, star := range stars {
		isCreator := 0
		if currentUserID > 0 && star.CreatorID == currentUserID {
			isCreator = 1
		}

		result = append(result, ds.ParallaxStarCatalogItemSerializer{
			ID:         star.ID,
			Name:       star.Name,
			ImageURL:   star.ImageURL,
			Distance:   star.Distance,
			LikesCount: len(star.Likes),
			IsCreator:  isCreator,
		})
	}

	ctx.JSON(http.StatusOK, result)
}

// GetParallaxStarFeedAPI godoc
// @Summary Лента опубликованных звезд (Vibes)
// @Description Возвращает опубликованную звезду для вертикальной ленты с признаком is_liked. При параметрах ?id=...&next=true возвращается следующая звезда.
// @Tags parallax_stars
// @Produce json
// @Param id query int false "Текущий ID звезды"
// @Param next query bool false "Перейти к следующей звезде"
// @Success 200 {object} ds.ParallaxStarFeedSerializer
// @Failure 400 {object} ds.ErrorResponse
// @Failure 404 {object} ds.ErrorResponse
// @Router /api/parallax_stars/feed [get]
func (h *Handler) GetParallaxStarFeedAPI(ctx *gin.Context) {
	currentUserID := GetContextUserID(ctx)
	idStr := ctx.Query("id")
	next := ctx.Query("next") == "true"

	var starID *int
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
				Error: "Неверный ID звезды",
			})
			return
		}
		starID = &id
	}

	star, err := h.Repository.GetParallaxStarFeed(starID, next)
	if err != nil || star == nil {
		ctx.JSON(http.StatusNotFound, ds.ErrorResponse{
			Error: "Опубликованная звезда не найдена",
		})
		return
	}

	serializer := ds.ToParallaxStarFeedSerializer(star, currentUserID)
	ctx.JSON(http.StatusOK, serializer)
}

// GetDraftAPI godoc
// @Summary Черновик текущего пользователя
// @Description Возвращает активный черновик услуги-звезды авторизованного создателя (синглтон, не более 1 на пользователя)
// @Tags parallax_stars
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ds.ParallaxStarDraftSerializer
// @Failure 401 {object} ds.ErrorResponse
// @Failure 404 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/parallax_stars/draft [get]
func (h *Handler) GetDraftAPI(ctx *gin.Context) {
	currentUserID := GetContextUserID(ctx)
	if currentUserID == 0 {
		ctx.JSON(http.StatusUnauthorized, ds.ErrorResponse{
			Error: "Требуется авторизация",
		})
		return
	}

	draft, err := h.Repository.GetDraftByCreator(currentUserID)
	if err != nil {
		logrus.Errorf("Ошибка поиска черновика: %v", err)
		ctx.JSON(http.StatusInternalServerError, ds.ErrorResponse{
			Error: "Ошибка сервера при поиске черновика",
		})
		return
	}

	if draft == nil {
		ctx.JSON(http.StatusNotFound, ds.ErrorResponse{
			Error: "Активный черновик у пользователя отсутствует",
		})
		return
	}

	serializer := ds.ToParallaxStarDraftSerializer(draft)
	ctx.JSON(http.StatusOK, serializer)
}

// AddParallaxStarAPI godoc
// @Summary Добавление новой звезды (черновик) с медиафайлами
// @Description Создает черновик услуги с автозаполнением создателя из JWT токена и загрузкой фото/видео в MinIO
// @Tags parallax_stars
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Название звезды"
// @Param description formData string false "Краткое описание"
// @Param parallax formData number false "Годичный параллакс (угловые секунды)"
// @Param distance formData number false "Расстояние до звезды (пк)"
// @Param image formData file false "Изображение звезды"
// @Param pic formData file false "Альтернативное имя поля изображения"
// @Param video formData file false "Видеоролик звезды"
// @Success 201 {object} ds.ParallaxStarDraftSerializer
// @Failure 400 {object} ds.ErrorResponse
// @Failure 401 {object} ds.ErrorResponse
// @Router /api/parallax_stars [post]
func (h *Handler) AddParallaxStarAPI(ctx *gin.Context) {
	currentUserID := GetContextUserID(ctx)
	if currentUserID == 0 {
		ctx.JSON(http.StatusUnauthorized, ds.ErrorResponse{
			Error: "Требуется авторизация",
		})
		return
	}

	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: "Ошибка чтения multipart формы: " + err.Error(),
		})
		return
	}

	name := strings.TrimSpace(ctx.Request.FormValue("name"))
	if name == "" {
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: "Поле 'name' обязательно для заполнения",
		})
		return
	}

	description := strings.TrimSpace(ctx.Request.FormValue("description"))

	var parallax float64
	if pStr := ctx.Request.FormValue("parallax"); pStr != "" {
		p, err := strconv.ParseFloat(pStr, 64)
		if err != nil || p < 0 || p > 9.999 {
			ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
				Error: "Параллакс должен быть числом от 0 до 9.999",
			})
			return
		}
		parallax = math.Round(p*1000) / 1000
	}

	var distance float64
	if dStr := ctx.Request.FormValue("distance"); dStr != "" {
		d, err := strconv.ParseFloat(dStr, 64)
		if err != nil || d < 0 || d > 999.99 {
			ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
				Error: "Расстояние должно быть числом от 0 до 999.99",
			})
			return
		}
		distance = math.Round(d*100) / 100
	}

	imageHeader, _ := ctx.FormFile("image")
	if imageHeader == nil {
		imageHeader, _ = ctx.FormFile("pic")
	}

	videoHeader, _ := ctx.FormFile("video")

	star, err := h.Repository.CreateParallaxStarWithMedia(
		currentUserID,
		name,
		description,
		parallax,
		distance,
		imageHeader,
		videoHeader,
	)
	if err != nil {
		logrus.Errorf("Ошибка создания звезды: %v", err)
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	serializer := ds.ToParallaxStarDraftSerializer(star)
	ctx.JSON(http.StatusCreated, serializer)
}

// PublishParallaxStarAPI godoc
// @Summary Публикация черновика звезды
// @Description Переводит черновик текущего создателя в статус published, рассчитывает расстояние и выставляет дату завершения
// @Tags parallax_stars
// @Security BearerAuth
// @Accept json
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID звезды"
// @Param request body ds.PublishParallaxStarRequest true "Параметры публикации"
// @Success 200 {object} ds.ParallaxStarFeedSerializer
// @Failure 400 {object} ds.ErrorResponse
// @Failure 401 {object} ds.ErrorResponse
// @Failure 403 {object} ds.ErrorResponse
// @Router /api/parallax_stars/{id}/publish [put]
func (h *Handler) PublishParallaxStarAPI(ctx *gin.Context) {
	currentUserID := GetContextUserID(ctx)
	if currentUserID == 0 {
		ctx.JSON(http.StatusUnauthorized, ds.ErrorResponse{
			Error: "Требуется авторизация",
		})
		return
	}

	idStr := ctx.Param("id")
	starID, err := strconv.Atoi(idStr)
	if err != nil || starID <= 0 {
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: "Некорректный ID звезды",
		})
		return
	}

	var req ds.PublishParallaxStarRequest
	if strings.Contains(ctx.GetHeader("Content-Type"), "application/json") {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
				Error: "Некорректный JSON: " + err.Error(),
			})
			return
		}
	} else {
		req.Description = strings.TrimSpace(ctx.PostForm("description"))
		if pStr := ctx.PostForm("parallax"); pStr != "" {
			if p, err := strconv.ParseFloat(pStr, 64); err == nil {
				req.Parallax = &p
			}
		}
		if dStr := ctx.PostForm("distance"); dStr != "" {
			if d, err := strconv.ParseFloat(dStr, 64); err == nil {
				req.Distance = &d
			}
		}
	}

	if req.Parallax != nil {
		if *req.Parallax <= 0 || *req.Parallax > 9.999 {
			ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
				Error: "Годичный параллакс должен быть в диапазоне от 0.001 до 9.999",
			})
			return
		}
		p := math.Round((*req.Parallax)*1000) / 1000
		req.Parallax = &p
	}

	if req.Distance != nil {
		if *req.Distance < 0.01 || *req.Distance > 999.99 {
			ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
				Error: "Расстояние должно быть в диапазоне от 0.01 до 999.99",
			})
			return
		}
		d := math.Round((*req.Distance)*100) / 100
		req.Distance = &d
	}

	publishedStar, err := h.Repository.PublishParallaxStarAPI(
		uint(starID),
		currentUserID,
		req.Description,
		req.Parallax,
		req.Distance,
	)
	if err != nil {
		logrus.Errorf("Ошибка публикации звезды: %v", err)
		if strings.Contains(err.Error(), "только создатель") {
			ctx.JSON(http.StatusForbidden, ds.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	serializer := ds.ToParallaxStarFeedSerializer(publishedStar, currentUserID)
	ctx.JSON(http.StatusOK, serializer)
}

// DeleteParallaxStarAPI godoc
// @Summary Логическое удаление звезды (soft delete)
// @Description Переводит статус услуги в deleted с помощью SQL UPDATE. Доступно только создателю.
// @Tags parallax_stars
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID звезды"
// @Success 200 {object} ds.MessageResponse
// @Failure 400 {object} ds.ErrorResponse
// @Failure 401 {object} ds.ErrorResponse
// @Failure 403 {object} ds.ErrorResponse
// @Router /api/parallax_stars/{id} [delete]
func (h *Handler) DeleteParallaxStarAPI(ctx *gin.Context) {
	currentUserID := GetContextUserID(ctx)
	if currentUserID == 0 {
		ctx.JSON(http.StatusUnauthorized, ds.ErrorResponse{
			Error: "Требуется авторизация",
		})
		return
	}

	idStr := ctx.Param("id")
	starID, err := strconv.Atoi(idStr)
	if err != nil || starID <= 0 {
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: "Некорректный ID звезды",
		})
		return
	}

	err = h.Repository.SoftDeleteParallaxStarByCreator(uint(starID), currentUserID)
	if err != nil {
		logrus.Errorf("Ошибка логического удаления: %v", err)
		if strings.Contains(err.Error(), "только создатель") {
			ctx.JSON(http.StatusForbidden, ds.ErrorResponse{
				Error: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, ds.MessageResponse{
		Message: "Звезда успешно удалена (soft delete)",
	})
}

// LikeParallaxStarAPI godoc
// @Summary Установка или отмена лайка
// @Description Добавляет (like=1) или удаляет (like=0) лайк от текущего авторизованного пользователя
// @Tags parallax_stars
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID звезды"
// @Param request body ds.LikeParallaxStarRequest true "Параметр like (1 или 0)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ds.ErrorResponse
// @Failure 401 {object} ds.ErrorResponse
// @Router /api/parallax_stars/{id}/like [post]
func (h *Handler) LikeParallaxStarAPI(ctx *gin.Context) {
	currentUserID := GetContextUserID(ctx)
	if currentUserID == 0 {
		ctx.JSON(http.StatusUnauthorized, ds.ErrorResponse{
			Error: "Требуется авторизация для управления лайками",
		})
		return
	}

	starID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || starID <= 0 {
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{Error: "Некорректный ID звезды"})
		return
	}

	var req ds.LikeParallaxStarRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{Error: "Поле 'like' должно быть 1 или 0"})
		return
	}

	isLiked, likesCount, err := h.Repository.ToggleStarLike(uint(starID), currentUserID, req.Like)
	if err != nil {
		logrus.Errorf("Ошибка обработки лайка: %v", err)
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"is_liked":    isLiked,
		"likes_count": likesCount,
	})
}
