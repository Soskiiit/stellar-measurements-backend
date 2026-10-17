package handler

import (
	"net/http"

	"stellar-measurements-backend/internal/app/config"
	"stellar-measurements-backend/internal/app/redis"
	"stellar-measurements-backend/internal/app/repository"
	"stellar-measurements-backend/internal/app/role"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	Config     *config.Config
}

func NewHandler(r *repository.Repository, red *redis.Client, cfg *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Redis:      red,
		Config:     cfg,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	// Swagger UI и редиректы
	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.PersistAuthorization(true), // сохраняет Bearer токен при перезагрузке страницы
	))
	router.GET("/swagger", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/docs", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Корневые алиасы (согласно методичке):
	router.POST("/login", h.LoginUserAPI)
	router.POST("/register", h.RegisterUserAPI)
	router.POST("/logout", h.WithAuthCheck(role.Astronomer, role.Moderator), h.LogoutUserAPI)

	// ==================== REST API МАРШРУТЫ (/api) ====================
	api := router.Group("/api")
	{
		// Аутентификация и регистрация пользователей (публичные):
		api.POST("/users/register", h.RegisterUserAPI)
		api.POST("/users/login", h.LoginUserAPI)

		// Публичные методы получения услуг (работают как для гостей, так и для авторизованных):
		api.GET("/parallax_stars", h.WithOptionalAuth(), h.GetParallaxStarsAPI)
		api.GET("/parallax_stars/feed", h.WithOptionalAuth(), h.GetParallaxStarFeedAPI)

		// Защищенные маршруты (требуют валидного JWT и разрешенной роли):
		authorized := api.Group("")
		authorized.Use(h.WithAuthCheck(role.Astronomer, role.Moderator))
		{
			// Деавторизация (с занесением токена в черный список Redis):
			authorized.POST("/users/logout", h.LogoutUserAPI)
			authorized.GET("/users/:id", h.GetUserByIDAPI)

			// Методы создателя услуги:
			authorized.GET("/parallax_stars/draft", h.GetDraftAPI)
			authorized.POST("/parallax_stars", h.AddParallaxStarAPI)
			authorized.PUT("/parallax_stars/:id/publish", h.PublishParallaxStarAPI)
			authorized.DELETE("/parallax_stars/:id", h.DeleteParallaxStarAPI)
			authorized.POST("/parallax_stars/:id/like", h.LikeParallaxStarAPI)
		}
	}

	// Fallback для несуществующих маршрутов
	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Маршрут не найден",
		})
	})
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, errMsg string) {
	logrus.Error(errMsg)
	ctx.JSON(errorStatusCode, gin.H{
		"error": errMsg,
	})
}
