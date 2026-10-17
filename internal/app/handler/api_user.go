package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"stellar-measurements-backend/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

// RegisterUserAPI godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя в системе с ролью создателя услуг (Astronomer)
// @Tags users
// @Accept json
// @Accept mpfd
// @Produce json
// @Param request body ds.UserRegisterRequest true "Данные регистрации"
// @Success 201 {object} ds.UserSerializer
// @Failure 400 {object} ds.ErrorResponse
// @Router /api/users/register [post]
func (h *Handler) RegisterUserAPI(ctx *gin.Context) {
	var req ds.UserRegisterRequest
	if strings.Contains(ctx.GetHeader("Content-Type"), "application/json") {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
				Error: "Некорректные данные: " + err.Error(),
			})
			return
		}
	} else {
		req.Login = strings.TrimSpace(ctx.PostForm("login"))
		req.Password = strings.TrimSpace(ctx.PostForm("password"))
	}

	if req.Login == "" || req.Password == "" {
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: "Поля 'login' и 'password' обязательны для заполнения",
		})
		return
	}

	newUser, err := h.Repository.RegisterUser(req.Login, req.Password)
	if err != nil {
		logrus.Warnf("Ошибка регистрации: %v", err)
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, ds.ToUserSerializer(newUser))
}

// LoginUserAPI godoc
// @Summary Аутентификация пользователя
// @Description Проверяет логин и пароль, возвращает JWT токен и выставляет HTTP-only cookie
// @Tags users
// @Accept json
// @Accept mpfd
// @Produce json
// @Param request body ds.UserLoginRequest true "Учетные данные пользователя"
// @Success 200 {object} ds.LoginResponse
// @Failure 400 {object} ds.ErrorResponse
// @Failure 401 {object} ds.ErrorResponse
// @Failure 500 {object} ds.ErrorResponse
// @Router /api/users/login [post]
func (h *Handler) LoginUserAPI(ctx *gin.Context) {
	var req ds.UserLoginRequest
	if strings.Contains(ctx.GetHeader("Content-Type"), "application/json") {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
				Error: "Некорректные данные: " + err.Error(),
			})
			return
		}
	} else {
		req.Login = strings.TrimSpace(ctx.PostForm("login"))
		req.Password = strings.TrimSpace(ctx.PostForm("password"))
	}

	if req.Login == "" || req.Password == "" {
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: "Поля 'login' и 'password' обязательны",
		})
		return
	}

	user, err := h.Repository.AuthenticateUser(req.Login, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ds.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	secret := "stellar-secret-token-key-2026"
	expiresIn := 24 * time.Hour
	if h.Config != nil {
		if h.Config.JWT.Token != "" {
			secret = h.Config.JWT.Token
		}
		if h.Config.JWT.ExpiresIn > 0 {
			expiresIn = h.Config.JWT.ExpiresIn
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "stellar-measurements-backend",
		},
		UserID: user.ID,
		Login:  user.Login,
		Role:   user.Role,
	})

	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ds.ErrorResponse{
			Error: "Ошибка генерации JWT токена: " + err.Error(),
		})
		return
	}

	maxAge := int(expiresIn.Seconds())
	ctx.SetCookie("access_token", tokenStr, maxAge, "/", "", false, true)
	ctx.SetCookie("token", tokenStr, maxAge, "/", "", false, true)

	ctx.JSON(http.StatusOK, ds.LoginResponse{
		ExpiresIn:   int64(maxAge),
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		User:        ds.ToUserSerializer(user),
	})
}

// LogoutUserAPI godoc
// @Summary Деавторизация пользователя
// @Description Отзывает текущий JWT токен и заносит его в черный список (blacklist) Redis
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ds.MessageResponse
// @Failure 400 {object} ds.ErrorResponse
// @Failure 401 {object} ds.ErrorResponse
// @Router /api/users/logout [post]
func (h *Handler) LogoutUserAPI(ctx *gin.Context) {
	jwtStr := h.extractToken(ctx)
	if jwtStr == "" {
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: "Токен авторизации не передан",
		})
		return
	}

	ttl := 24 * time.Hour
	if h.Config != nil && h.Config.JWT.ExpiresIn > 0 {
		ttl = h.Config.JWT.ExpiresIn
	}

	userLogin := ctx.GetString("userLogin")

	if h.Redis != nil {
		if err := h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), jwtStr, userLogin, ttl); err != nil {
			logrus.Errorf("Ошибка записи токена в черный список Redis: %v", err)
		} else {
			logrus.Infof("Токен успешно добавлен в черный список Redis (user: %s): %s...", userLogin, jwtStr[:min(len(jwtStr), 15)])
		}
	}

	// Очищаем куки
	ctx.SetCookie("access_token", "", -1, "/", "", false, true)
	ctx.SetCookie("token", "", -1, "/", "", false, true)

	ctx.JSON(http.StatusOK, ds.MessageResponse{
		Message: "Деавторизация выполнена успешно: токен помещен в черный список",
	})
}

// GetUserByIDAPI godoc
// @Summary Получение информации о пользователе по ID
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} ds.UserSerializer
// @Failure 404 {object} ds.ErrorResponse
// @Router /api/users/{id} [get]
func (h *Handler) GetUserByIDAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, ds.ErrorResponse{
			Error: "Некорректный ID пользователя",
		})
		return
	}

	user, err := h.Repository.GetUserByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ds.ErrorResponse{
			Error: "Пользователь не найден",
		})
		return
	}

	ctx.JSON(http.StatusOK, ds.ToUserSerializer(user))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
