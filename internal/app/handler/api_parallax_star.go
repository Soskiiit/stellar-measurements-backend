package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"stellar-measurements-backend/internal/app/ds"
	"stellar-measurements-backend/internal/app/session"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetParallaxStarsAPI - GET /api/parallax_stars
// Получение списка опубликованных звезд с возможностью фильтрации по расстоянию.
// Для каждой звезды возвращается признак is_creator (0/1), если создатель совпадает с текущим пользователем.
func (h *Handler) GetParallaxStarsAPI(ctx *gin.Context) {
	currentUserID := session.GetCurrentUserID()
	distanceStr := ctx.Query("distance")

	var maxDistance *float64
	if distanceStr != "" {
		d, err := strconv.ParseFloat(distanceStr, 64)
		if err != nil || d < 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Некорректное значение фильтра distance (должно быть неотрицательным числом)",
			})
			return
		}
		maxDistance = &d
	}

	stars, err := h.Repository.GetPublishedParallaxStarsFiltered(maxDistance)
	if err != nil {
		logrus.Errorf("Ошибка получения каталога звезд: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось загрузить список звезд",
		})
		return
	}

	result := make([]ds.ParallaxStarCatalogItemSerializer, 0, len(stars))
	for _, star := range stars {
		isCreator := 0
		if star.CreatorID == currentUserID {
			isCreator = 1
		}

		result = append(result, ds.ParallaxStarCatalogItemSerializer{
			ID:          star.ID,
			Name:        star.Name,
			Description: star.Description,
			Status:      star.Status,
			ImageURL:    star.ImageURL,
			VideoURL:    star.VideoURL,
			Parallax:    star.Parallax,
			Distance:    star.Distance,
			LikesCount:  len(star.Likes),
			IsCreator:   isCreator,
		})
	}

	ctx.JSON(http.StatusOK, result)
}

// GetParallaxStarFeedAPI - GET /api/parallax_stars/feed
// Лента опубликованных звезд (только опубликованные). Возвращает признак is_liked (0/1),
// если текущий пользователь лайкнул эту услугу. С параметрами ?id=...&next=true переходит к следующей звезде.
func (h *Handler) GetParallaxStarFeedAPI(ctx *gin.Context) {
	currentUserID := session.GetCurrentUserID()
	idStr := ctx.Query("id")
	next := ctx.Query("next") == "true"

	var starID *int
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Неверный ID звезды",
			})
			return
		}
		starID = &id
	}

	star, err := h.Repository.GetParallaxStarFeed(starID, next)
	if err != nil || star == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Опубликованная звезда не найдена",
		})
		return
	}

	serializer := ds.ToFullParallaxStarSerializer(star, currentUserID)
	ctx.JSON(http.StatusOK, serializer)
}

// GetDraftAPI - GET /api/parallax_stars/draft
// Получение черновика текущего пользователя (не более 1 записи, ID не указывается в URL).
func (h *Handler) GetDraftAPI(ctx *gin.Context) {
	currentUserID := session.GetCurrentUserID()
	draft, err := h.Repository.GetDraftByCreator(currentUserID)
	if err != nil {
		logrus.Errorf("Ошибка поиска черновика: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка сервера при поиске черновика",
		})
		return
	}

	if draft == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Активный черновик у пользователя отсутствует",
		})
		return
	}

	serializer := ds.ToFullParallaxStarSerializer(draft, currentUserID)
	ctx.JSON(http.StatusOK, serializer)
}

// AddParallaxStarAPI - POST /api/parallax_stars
// Добавление новой услуги-звезды в статусе черновик с загрузкой файлов изображения и видео в MinIO.
// Названия файлов генерируются на латинице. Системные поля рассчитываются бэкендом.
func (h *Handler) AddParallaxStarAPI(ctx *gin.Context) {
	currentUserID := session.GetCurrentUserID()

	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Ошибка чтения multipart формы: " + err.Error(),
		})
		return
	}

	name := strings.TrimSpace(ctx.Request.FormValue("name"))
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Поле 'name' обязательно для заполнения",
		})
		return
	}

	description := strings.TrimSpace(ctx.Request.FormValue("description"))

	var parallax float64
	if pStr := ctx.Request.FormValue("parallax"); pStr != "" {
		p, err := strconv.ParseFloat(pStr, 64)
		if err != nil || p < 0 || p > 9.999 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Параллакс должен быть числом от 0 до 9.999",
			})
			return
		}
		parallax = math.Round(p*1000) / 1000
	}

	var distance float64
	if dStr := ctx.Request.FormValue("distance"); dStr != "" {
		d, err := strconv.ParseFloat(dStr, 64)
		if err != nil || d < 0 || d > 999.99 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Расстояние должно быть числом от 0 до 999.99",
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
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	serializer := ds.ToFullParallaxStarSerializer(star, currentUserID)
	ctx.JSON(http.StatusCreated, serializer)
}

// PublishParallaxStarAPI - PUT /api/parallax_stars/:id/publish
// Публикация черновика (смена статуса на 'published'): фиксация даты формирования (date_finish),
// заполнение описания, годичного параллакса и расчет расстояния.
func (h *Handler) PublishParallaxStarAPI(ctx *gin.Context) {
	currentUserID := session.GetCurrentUserID()
	idStr := ctx.Param("id")
	starID, err := strconv.Atoi(idStr)
	if err != nil || starID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректный ID звезды",
		})
		return
	}

	var req ds.PublishParallaxStarRequest
	if strings.Contains(ctx.GetHeader("Content-Type"), "application/json") {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Некорректный JSON: " + err.Error(),
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
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Годичный параллакс должен быть в диапазоне от 0.001 до 9.999",
			})
			return
		}
		p := math.Round((*req.Parallax)*1000) / 1000
		req.Parallax = &p
	}

	if req.Distance != nil {
		if *req.Distance < 0.01 || *req.Distance > 999.99 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Расстояние должно быть в диапазоне от 0.01 до 999.99",
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
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	serializer := ds.ToFullParallaxStarSerializer(publishedStar, currentUserID)
	ctx.JSON(http.StatusOK, serializer)
}

// DeleteParallaxStarAPI - DELETE /api/parallax_stars/:id
// Логическое удаление услуги (только soft delete, статус меняется на 'deleted').
// Разрешено только создателю услуги.
func (h *Handler) DeleteParallaxStarAPI(ctx *gin.Context) {
	currentUserID := session.GetCurrentUserID()
	idStr := ctx.Param("id")
	starID, err := strconv.Atoi(idStr)
	if err != nil || starID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректный ID звезды",
		})
		return
	}

	err = h.Repository.SoftDeleteParallaxStarByCreator(uint(starID), currentUserID)
	if err != nil {
		logrus.Errorf("Ошибка логического удаления: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Звезда успешно удалена (soft delete)",
	})
}

// LikeParallaxStarAPI - POST /api/parallax_stars/:id/like
// Поставить (like=1) или отменить (like=0) лайк от текущего пользователя.
func (h *Handler) LikeParallaxStarAPI(ctx *gin.Context) {
	starID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || starID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID звезды"})
		return
	}

	var req ds.LikeParallaxStarRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Поле 'like' должно быть 1 или 0"})
		return
	}

	isLiked, likesCount, err := h.Repository.ToggleStarLike(uint(starID), session.GetCurrentUserID(), req.Like)
	if err != nil {
		logrus.Errorf("Ошибка обработки лайка: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"is_liked":    isLiked,
		"likes_count": likesCount,
	})
}
