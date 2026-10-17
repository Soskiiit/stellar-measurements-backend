package redis

import (
	"context"
	"time"
)

const jwtPrefix = "jwt."

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

// WriteJWTToBlacklist добавляет JWT токен в черный список на время оставшейся жизни токена,
// сохраняя информацию о пользователе (для демонстрации содержимого в Redis)
func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, userLogin string, jwtTTL time.Duration) error {
	val := "revoked"
	if userLogin != "" {
		val = "user:" + userLogin
	}
	return c.client.Set(ctx, getJWTKey(jwtStr), val, jwtTTL).Err()
}

// CheckJWTInBlacklist проверяет наличие JWT токена в черном списке
// Если токен отсутствует, возвращает redis.Nil
func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) error {
	return c.client.Get(ctx, getJWTKey(jwtStr)).Err()
}
