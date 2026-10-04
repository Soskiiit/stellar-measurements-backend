package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"mime/multipart"
	"time"

	"stellar-measurements-backend/internal/app/ds"

	"gorm.io/gorm"
)

const (
	DefaultImageURL = "/static/img/parallax_star_default_404.jpg"
	DefaultVideoURL = "/static/videos/parallax_star_default.mp4"
)

func (r *Repository) GetPublishedParallaxStars() ([]ds.ParallaxStar, error) {
	return r.GetPublishedParallaxStarsFiltered(nil)
}

func (r *Repository) GetPublishedStars() ([]ds.ParallaxStar, error) {
	return r.GetPublishedParallaxStars()
}

func (r *Repository) GetPublishedParallaxStarsFiltered(maxDistance *float64) ([]ds.ParallaxStar, error) {
	var stars []ds.ParallaxStar
	query := r.db.Preload("Likes").Preload("Creator").
		Where("status = ?", ds.StatusPublished)

	if maxDistance != nil && *maxDistance > 0 {
		query = query.Where("distance <= ?", *maxDistance)
	}

	err := query.Order("id ASC").Find(&stars).Error
	if err != nil {
		return nil, err
	}
	return stars, nil
}

func (r *Repository) GetFirstPublishedParallaxStar() (*ds.ParallaxStar, error) {
	var star ds.ParallaxStar
	err := r.db.Preload("Likes").Preload("Creator").
		Where("status = ?", ds.StatusPublished).
		Order("id ASC").
		First(&star).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &star, nil
}

func (r *Repository) GetFirstPublishedStar() (*ds.ParallaxStar, error) {
	return r.GetFirstPublishedParallaxStar()
}

func (r *Repository) GetNextPublishedParallaxStarID(currentID int) (uint, error) {
	var star ds.ParallaxStar
	err := r.db.Select("id").
		Where("status = ? AND id > ?", ds.StatusPublished, currentID).
		Order("id ASC").
		First(&star).Error

	if err == nil {
		return star.ID, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	err = r.db.Select("id").
		Where("status = ?", ds.StatusPublished).
		Order("id ASC").
		First(&star).Error
	if err != nil {
		return 0, err
	}

	return star.ID, nil
}

func (r *Repository) GetNextPublishedStarID(currentID int) (uint, error) {
	return r.GetNextPublishedParallaxStarID(currentID)
}

func (r *Repository) GetParallaxStarFeed(starID *int, next bool) (*ds.ParallaxStar, error) {
	if starID == nil {
		return r.GetFirstPublishedParallaxStar()
	}

	targetID := *starID
	if next {
		nextID, err := r.GetNextPublishedParallaxStarID(targetID)
		if err != nil {
			return nil, err
		}
		targetID = int(nextID)
	}

	return r.GetParallaxStarByID(targetID)
}

func (r *Repository) GetParallaxStarsByDistance(maxDistance float64) ([]ds.ParallaxStar, error) {
	return r.GetPublishedParallaxStarsFiltered(&maxDistance)
}

func (r *Repository) GetStarsByDistance(maxDistance float64) ([]ds.ParallaxStar, error) {
	return r.GetParallaxStarsByDistance(maxDistance)
}

func (r *Repository) GetParallaxStarByID(id int) (*ds.ParallaxStar, error) {
	var star ds.ParallaxStar
	err := r.db.Preload("Likes").Preload("Creator").
		Where("id = ? AND status != ?", id, ds.StatusDeleted).
		First(&star).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &star, nil
}

func (r *Repository) GetStarByID(id int) (*ds.ParallaxStar, error) {
	return r.GetParallaxStarByID(id)
}

func (r *Repository) GetParallaxStarByIDCursor(id int) (*ds.ParallaxStar, error) {
	query := "SELECT id, name, description, status, image_url, video_url, parallax, distance, date_create, creator_id FROM parallax_stars WHERE id = $1 AND status = 'published'"
	row := r.db.Raw(query, id).Row()

	var star ds.ParallaxStar
	err := row.Scan(
		&star.ID,
		&star.Name,
		&star.Description,
		&star.Status,
		&star.ImageURL,
		&star.VideoURL,
		&star.Parallax,
		&star.Distance,
		&star.DateCreate,
		&star.CreatorID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &star, nil
}

func (r *Repository) GetStarByIDCursor(id int) (*ds.ParallaxStar, error) {
	return r.GetParallaxStarByIDCursor(id)
}

func (r *Repository) GetDraftByCreator(creatorID uint) (*ds.ParallaxStar, error) {
	var star ds.ParallaxStar
	err := r.db.Preload("Likes").Preload("Creator").
		Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).
		First(&star).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &star, nil
}

func (r *Repository) CreateDraft(creatorID uint, name string) (*ds.ParallaxStar, error) {
	draft := ds.ParallaxStar{
		Name:        name,
		Status:      ds.StatusDraft,
		CreatorID:   creatorID,
		DateCreate:  time.Now(),
		Description: "",
		ImageURL:    DefaultImageURL,
		VideoURL:    DefaultVideoURL,
		Parallax:    0,
		Distance:    0,
	}

	err := r.db.Create(&draft).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка создания черновика: %w", err)
	}

	return &draft, nil
}

func (r *Repository) CreateParallaxStarWithMedia(
	creatorID uint,
	name string,
	description string,
	parallax float64,
	distance float64,
	imageHeader *multipart.FileHeader,
	videoHeader *multipart.FileHeader,
) (*ds.ParallaxStar, error) {
	existingDraft, err := r.GetDraftByCreator(creatorID)
	if err != nil {
		return nil, err
	}
	if existingDraft != nil {
		return nil, fmt.Errorf("у вас уже есть активный черновик (id=%d), завершите публикацию или удалите его", existingDraft.ID)
	}

	imageURL := DefaultImageURL
	if imageHeader != nil {
		uploadedImageURL, uploadErr := r.UploadMedia(imageHeader, "star_img")
		if uploadErr != nil {
			return nil, fmt.Errorf("ошибка загрузки изображения в MinIO: %w", uploadErr)
		}
		imageURL = uploadedImageURL
	}

	videoURL := DefaultVideoURL
	if videoHeader != nil {
		uploadedVideoURL, uploadErr := r.UploadMedia(videoHeader, "star_video")
		if uploadErr != nil {
			return nil, fmt.Errorf("ошибка загрузки видео в MinIO: %w", uploadErr)
		}
		videoURL = uploadedVideoURL
	}

	if distance == 0 && parallax > 0 {
		distance = math.Round((1.0/parallax)*100) / 100
	}

	newStar := ds.ParallaxStar{
		Name:        name,
		Description: description,
		Status:      ds.StatusDraft,
		ImageURL:    imageURL,
		VideoURL:    videoURL,
		Parallax:    parallax,
		Distance:    distance,
		DateCreate:  time.Now(),
		CreatorID:   creatorID,
	}

	if err := r.db.Create(&newStar).Error; err != nil {
		return nil, fmt.Errorf("ошибка сохранения звезды в БД: %w", err)
	}

	_ = r.db.Preload("Creator").Preload("Likes").First(&newStar, newStar.ID)
	return &newStar, nil
}

func (r *Repository) PublishParallaxStar(starID uint, description string, parallax float64, distance float64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"description": description,
		"parallax":    parallax,
		"distance":    distance,
		"status":      ds.StatusPublished,
		"date_finish": sql.NullTime{Time: now, Valid: true},
	}

	err := r.db.Model(&ds.ParallaxStar{}).Where("id = ?", starID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("ошибка публикации звезды: %w", err)
	}

	return nil
}

func (r *Repository) PublishStar(starID uint, description string, parallax float64, distance float64) error {
	return r.PublishParallaxStar(starID, description, parallax, distance)
}

func (r *Repository) PublishParallaxStarAPI(
	starID uint,
	creatorID uint,
	description string,
	parallax *float64,
	distance *float64,
) (*ds.ParallaxStar, error) {
	var star ds.ParallaxStar
	err := r.db.Where("id = ?", starID).First(&star).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("звезда с id=%d не найдена", starID)
		}
		return nil, err
	}

	if star.CreatorID != creatorID {
		return nil, fmt.Errorf("только создатель может опубликовать свою звезду")
	}

	if star.Status == ds.StatusPublished {
		return nil, fmt.Errorf("звезда уже опубликована")
	}
	if star.Status == ds.StatusDeleted {
		return nil, fmt.Errorf("нельзя опубликовать удаленную звезду")
	}

	updates := map[string]interface{}{
		"status":      ds.StatusPublished,
		"date_finish": sql.NullTime{Time: time.Now(), Valid: true},
	}

	if description != "" {
		updates["description"] = description
	}
	if parallax != nil && *parallax > 0 {
		updates["parallax"] = *parallax
		if distance == nil || *distance == 0 {
			calcDistance := math.Round((1.0/(*parallax))*100) / 100
			updates["distance"] = calcDistance
		}
	}
	if distance != nil && *distance > 0 {
		updates["distance"] = *distance
	}

	if err := r.db.Model(&ds.ParallaxStar{}).Where("id = ?", starID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("ошибка публикации: %w", err)
	}

	var updated ds.ParallaxStar
	_ = r.db.Preload("Likes").Preload("Creator").First(&updated, starID)
	return &updated, nil
}

func (r *Repository) DeleteParallaxStar(starID uint) error {
	query := "UPDATE parallax_stars SET status = $1 WHERE id = $2 RETURNING id"

	rows, err := r.db.Raw(query, ds.StatusDeleted, starID).Rows()
	if err != nil {
		return fmt.Errorf("ошибка открытия курсора при удалении звезды с id %d: %w", starID, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return fmt.Errorf("звезда с id %d не найдена", starID)
	}

	var deletedID uint
	if err := rows.Scan(&deletedID); err != nil {
		return fmt.Errorf("ошибка чтения из SQL курсора: %w", err)
	}

	return nil
}

func (r *Repository) DeleteStar(starID uint) error {
	return r.DeleteParallaxStar(starID)
}

func (r *Repository) SoftDeleteParallaxStarByCreator(starID uint, creatorID uint) error {
	var star ds.ParallaxStar
	err := r.db.Where("id = ?", starID).First(&star).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("звезда с id=%d не найдена", starID)
		}
		return err
	}

	if star.CreatorID != creatorID {
		return fmt.Errorf("удалять разрешено только свои услуги")
	}

	if star.Status == ds.StatusDeleted {
		return fmt.Errorf("звезда уже удалена")
	}

	return r.db.Model(&ds.ParallaxStar{}).Where("id = ?", starID).Update("status", ds.StatusDeleted).Error
}

func (r *Repository) ToggleStarLike(starID uint, userID uint, shouldLike int) (bool, int64, error) {
	var star ds.ParallaxStar
	err := r.db.Where("id = ? AND status = ?", starID, ds.StatusPublished).First(&star).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, 0, fmt.Errorf("опубликованная звезда с id=%d не найдена", starID)
		}
		return false, 0, err
	}

	var existingLike ds.ParallaxStarLike
	findErr := r.db.Where("user_id = ? AND parallax_star_id = ?", userID, starID).First(&existingLike).Error

	if shouldLike == 1 {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			newLike := ds.ParallaxStarLike{
				UserID:         userID,
				ParallaxStarID: starID,
			}
			if err := r.db.Create(&newLike).Error; err != nil {
				return false, 0, fmt.Errorf("ошибка добавления лайка: %w", err)
			}
		}
	} else if shouldLike == 0 {
		if findErr == nil {
			if err := r.db.Delete(&existingLike).Error; err != nil {
				return false, 0, fmt.Errorf("ошибка удаления лайка: %w", err)
			}
		}
	}

	var count int64
	r.db.Model(&ds.ParallaxStarLike{}).Where("parallax_star_id = ?", starID).Count(&count)
	isLiked := shouldLike == 1
	return isLiked, count, nil
}
