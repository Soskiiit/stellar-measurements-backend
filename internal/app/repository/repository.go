package repository

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type RepositorySettings struct {
	PostgresDSN     string
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string
	MinioUseSSL     bool
	MinioPublicURL  string
}

type Repository struct {
	db             *gorm.DB
	minio          *minio.Client
	minioBucket    string
	minioPublicURL string
}

func New(settings *RepositorySettings) (*Repository, error) {
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(settings.PostgresDSN), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	bucketName := settings.MinioBucketName
	if bucketName == "" {
		bucketName = "parallax-stars"
	}

	endpoint := settings.MinioEndpoint
	if endpoint == "" {
		endpoint = "localhost:9002"
	}

	publicURL := settings.MinioPublicURL
	if publicURL == "" {
		scheme := "http"
		if settings.MinioUseSSL {
			scheme = "https"
		}
		publicURL = fmt.Sprintf("%s://%s", scheme, endpoint)
	}
	publicURL = strings.TrimRight(publicURL, "/")

	var minioClient *minio.Client
	if settings.MinioAccessKey != "" || settings.MinioSecretKey != "" {
		client, errMinio := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(settings.MinioAccessKey, settings.MinioSecretKey, ""),
			Secure: settings.MinioUseSSL,
		})
		if errMinio != nil {
			logrus.Warnf("Ошибка создания клиента MinIO: %v", errMinio)
		} else {
			minioClient = client
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			exists, errExists := minioClient.BucketExists(ctx, bucketName)
			if errExists != nil {
				logrus.Warnf("Не удалось проверить бакет MinIO '%s': %v", bucketName, errExists)
			} else if !exists {
				errMake := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
				if errMake != nil {
					logrus.Warnf("Не удалось создать бакет MinIO '%s': %v", bucketName, errMake)
				} else {
					logrus.Infof("Создан бакет MinIO '%s'", bucketName)
				}
			}
		}
	}

	return &Repository{
		db:             db,
		minio:          minioClient,
		minioBucket:    bucketName,
		minioPublicURL: publicURL,
	}, nil
}

func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

// UploadMedia загружает файл в MinIO с латинским сгенерированным именем и возвращает URL
func (r *Repository) UploadMedia(header *multipart.FileHeader, prefix string) (string, error) {
	if header == nil {
		return "", nil
	}

	if r.minio == nil {
		return "", fmt.Errorf("minio клиент не инициализирован")
	}

	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка открытия загружаемого файла: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("ошибка чтения заголовка файла: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("ошибка перемотки файла: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		if strings.HasPrefix(contentType, "image/png") {
			ext = ".png"
		} else if strings.HasPrefix(contentType, "image/jpeg") {
			ext = ".jpg"
		} else if strings.HasPrefix(contentType, "image/webp") {
			ext = ".webp"
		} else if strings.HasPrefix(contentType, "video/mp4") {
			ext = ".mp4"
		} else {
			ext = ".bin"
		}
	}

	cleanPrefix := "media"
	if prefix != "" {
		cleanPrefix = strings.ToLower(strings.ReplaceAll(prefix, " ", "_"))
	}
	// Генерация уникального имени файла строго на латинице
	latinFileName := fmt.Sprintf("%s_%d_%s%s", cleanPrefix, time.Now().Unix(), uuid.New().String()[:8], ext)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = r.minio.PutObject(ctx, r.minioBucket, latinFileName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка отправки объекта в MinIO: %w", err)
	}

	fileURL := fmt.Sprintf("%s/%s/%s", r.minioPublicURL, r.minioBucket, latinFileName)
	return fileURL, nil
}
