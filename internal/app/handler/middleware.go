package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"stellar-measurements-backend/internal/app/ds"
	"stellar-measurements-backend/internal/app/role"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

const jwtPrefix = "Bearer "

// extractToken извлекает токен из заголовка Authorization или куки
func (h *Handler) extractToken(ctx *gin.Context) string {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, jwtPrefix) {
			return strings.TrimSpace(authHeader[len(jwtPrefix):])
		}
		// Поддержка передачи токена без префикса "Bearer "
		return strings.TrimSpace(authHeader)
	}

	if cookie, err := ctx.Cookie("access_token"); err == nil && cookie != "" {
		return cookie
	}
	if cookie, err := ctx.Cookie("token"); err == nil && cookie != "" {
		return cookie
	}

	// Поддержка ручной передачи заголовка Cookie (например, из insomnia/postman)
	cookieHeader := ctx.GetHeader("Cookie")
	if cookieHeader != "" {
		for _, part := range strings.Split(cookieHeader, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "token=") {
				return strings.TrimSpace(strings.TrimPrefix(part, "token="))
			}
			if strings.HasPrefix(part, "access_token=") {
				return strings.TrimSpace(strings.TrimPrefix(part, "access_token="))
			}
		}
		if strings.HasPrefix(cookieHeader, "eyJ") {
			return strings.TrimSpace(cookieHeader)
		}
	}

	return ""
}

// WithAuthCheck проверяет JWT токен, его отсутствие в Redis blacklist и права доступа по ролям
func (h *Handler) WithAuthCheck(assignedRoles ...role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jwtStr := h.extractToken(ctx)
		if jwtStr == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется авторизация: токен не передан",
			})
			return
		}

		// Проверка токена в черном списке Redis
		if h.Redis != nil {
			err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr)
			if err == nil {
				// Токен найден в черном списке (после logout)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Токен недействителен: находится в черном списке",
				})
				return
			}
			if !errors.Is(err, redis.Nil) {
				logrus.Warnf("Ошибка обращения к Redis blacklist: %v", err)
			}
		}

		secret := "stellar-secret-token-key-2026"
		if h.Config != nil && h.Config.JWT.Token != "" {
			secret = h.Config.JWT.Token
		}

		claims := &ds.JWTClaims{}
		token, err := jwt.ParseWithClaims(jwtStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Недействительный или истекший токен авторизации",
			})
			return
		}

		// Проверка роли пользователя
		if len(assignedRoles) > 0 {
			roleAllowed := false
			for _, r := range assignedRoles {
				if claims.Role == r {
					roleAllowed = true
					break
				}
			}
			if !roleAllowed {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": fmt.Sprintf("Роли %s запрещен доступ к данному ресурсу", claims.Role),
				})
				return
			}
		}

		ctx.Set("userID", claims.UserID)
		ctx.Set("userRole", claims.Role)
		ctx.Set("userLogin", claims.Login)
		ctx.Next()
	}
}

// WithOptionalAuth извлекает пользователя из токена, если он передан, но не блокирует гостевой доступ
func (h *Handler) WithOptionalAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jwtStr := h.extractToken(ctx)
		if jwtStr == "" {
			ctx.Next()
			return
		}

		if h.Redis != nil {
			if err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr); err == nil {
				ctx.Next()
				return
			}
		}

		secret := "stellar-secret-token-key-2026"
		if h.Config != nil && h.Config.JWT.Token != "" {
			secret = h.Config.JWT.Token
		}

		claims := &ds.JWTClaims{}
		token, err := jwt.ParseWithClaims(jwtStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err == nil && token.Valid {
			ctx.Set("userID", claims.UserID)
			ctx.Set("userRole", claims.Role)
			ctx.Set("userLogin", claims.Login)
		}

		ctx.Next()
	}
}

// GetContextUserID возвращает ID текущего пользователя из контекста или 0 для гостя
func GetContextUserID(ctx *gin.Context) uint {
	if val, exists := ctx.Get("userID"); exists {
		if id, ok := val.(uint); ok {
			return id
		}
	}
	return 0
}
