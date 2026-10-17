package main

import (
	"context"
	"os"

	"stellar-measurements-backend/internal/app/config"
	"stellar-measurements-backend/internal/app/dsn"
	"stellar-measurements-backend/internal/app/handler"
	"stellar-measurements-backend/internal/app/redis"
	"stellar-measurements-backend/internal/app/repository"
	"stellar-measurements-backend/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	_ "stellar-measurements-backend/docs"
)

// @title Stellar Measurements API
// @version 1.0
// @description REST API веб-сервиса по теме «Определение расстояния до звезды в Млечном пути по ее годовому параллаксу» с JWT-авторизацией.

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен с префиксом Bearer: "Bearer <ваш_токен>"

func main() {
	_ = godotenv.Load()
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	logrus.Infof("Connecting to DB: %s", postgresString)

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9002"
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "root"
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "rootpassword"
	}
	minioBucketName := os.Getenv("MINIO_BUCKET_NAME")
	if minioBucketName == "" {
		minioBucketName = "parallax-stars"
	}
	minioPublicURL := os.Getenv("MINIO_PUBLIC_URL")
	if minioPublicURL == "" {
		minioPublicURL = "http://" + minioEndpoint
	}

	rep, errRep := repository.New(&repository.RepositorySettings{
		PostgresDSN:     postgresString,
		MinioEndpoint:   minioEndpoint,
		MinioAccessKey:  minioAccessKey,
		MinioSecretKey:  minioSecretKey,
		MinioBucketName: minioBucketName,
		MinioPublicURL:  minioPublicURL,
		MinioUseSSL:     false,
	})
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// Инициализация Redis клиента для черного списка токенов (blacklist)
	ctx := context.Background()
	redisClient, errRedis := redis.New(ctx, conf.Redis)
	if errRedis != nil {
		logrus.Warnf("Could not connect to Redis: %v. Running with fallback (blacklist unavailable).", errRedis)
	} else {
		logrus.Infof("Connected to Redis at %s:%d", conf.Redis.Host, conf.Redis.Port)
		defer redisClient.Close()
	}

	hand := handler.NewHandler(rep, redisClient, conf)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
