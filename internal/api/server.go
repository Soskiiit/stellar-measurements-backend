package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"stellar-measurements-backend/internal/app/handler"
	"stellar-measurements-backend/internal/app/repository"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/", h.GetFeed)
	r.GET("/feed/:id", h.GetFeed)
	r.GET("/catalog", h.GetTileList)
	r.GET("/add", h.GetDraft)

	r.Run()
	log.Println("Server down")
}
