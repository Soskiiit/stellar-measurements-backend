package handler

import (
	"net/http"
	"strconv"
	"strings"

	"stellar-measurements-backend/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterUserAPI - POST /api/users/register
// Регистрация нового пользователя
func (h *Handler) RegisterUserAPI(ctx *gin.Context) {
	var req ds.UserRegisterRequest
	if strings.Contains(ctx.GetHeader("Content-Type"), "application/json") {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Некорректные данные: " + err.Error(),
			})
			return
		}
	} else {
		req.Login = strings.TrimSpace(ctx.PostForm("login"))
		req.Password = strings.TrimSpace(ctx.PostForm("password"))
	}

	if req.Login == "" || req.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Поля 'login' и 'password' обязательны для заполнения",
		})
		return
	}

	newUser, err := h.Repository.RegisterUser(req.Login, req.Password)
	if err != nil {
		logrus.Warnf("Ошибка регистрации: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, ds.ToUserSerializer(newUser))
}

// LoginUserAPI - POST /api/users/login
// Аутентификация пользователя (заглушка для 4-й лабораторной работы)
func (h *Handler) LoginUserAPI(ctx *gin.Context) {
	var req ds.UserLoginRequest
	if strings.Contains(ctx.GetHeader("Content-Type"), "application/json") {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Некорректные данные: " + err.Error(),
			})
			return
		}
	} else {
		req.Login = strings.TrimSpace(ctx.PostForm("login"))
		req.Password = strings.TrimSpace(ctx.PostForm("password"))
	}

	if req.Login == "" || req.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Поля 'login' и 'password' обязательны",
		})
		return
	}

	user, err := h.Repository.AuthenticateUser(req.Login, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Заглушка токена авторизации для 4-й лабораторной работы
	mockToken := "mock-jwt-token-for-lab4-" + user.Login

	ctx.JSON(http.StatusOK, gin.H{
		"token": mockToken,
		"user":  ds.ToUserSerializer(user),
	})
}

// LogoutUserAPI - POST /api/users/logout
// Деавторизация пользователя (заглушка для 4-й лабораторной работы)
func (h *Handler) LogoutUserAPI(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Деавторизация выполнена успешно",
	})
}

// GetUserWithStarsAPI - GET /api/users/:id
// Получение профиля пользователя со списком его созданных звезд (демонстрация вложенной сериализации)
func (h *Handler) GetUserWithStarsAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil || userID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректный ID пользователя",
		})
		return
	}

	data, err := h.Repository.GetUserWithStars(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
