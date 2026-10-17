package ds

import (
	"stellar-measurements-backend/internal/app/role"

	"github.com/golang-jwt/jwt"
)

// JWTClaims структура claims для токена авторизации
type JWTClaims struct {
	jwt.StandardClaims
	UserID uint      `json:"user_id"`
	Login  string    `json:"login"`
	Role   role.Role `json:"role"`
}

// LoginResponse - ответ при успешной аутентификации
type LoginResponse struct {
	ExpiresIn   int64          `json:"expires_in" example:"86400"`
	AccessToken string         `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string         `json:"token_type" example:"Bearer"`
	User        UserSerializer `json:"user"`
}

// MessageResponse - типовой ответ с текстовым сообщением
type MessageResponse struct {
	Message string `json:"message" example:"Операция выполнена успешно"`
}

// ErrorResponse - типовой ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error" example:"Описание ошибки"`
}
