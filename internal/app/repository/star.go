package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"stellar-measurements-backend/internal/app/ds"
)

// GetPublishedStars возвращает все опубликованные звёзды через ORM
func (r *Repository) GetPublishedStars() ([]ds.Star, error) {
	var stars []ds.Star
	err := r.db.Preload("Likes").
		Where("status = ?", ds.StatusPublished).
		Order("id ASC").
		Find(&stars).Error
	if err != nil {
		return nil, err
	}
	return stars, nil
}

// GetStarsByDistance выполняет фильтрацию опубликованных звёзд по расстоянию через ORM
func (r *Repository) GetStarsByDistance(maxDistance float64) ([]ds.Star, error) {
	var stars []ds.Star
	err := r.db.Preload("Likes").
		Where("status = ? AND distance <= ?", ds.StatusPublished, maxDistance).
		Order("distance ASC").
		Find(&stars).Error
	if err != nil {
		return nil, err
	}
	return stars, nil
}

// GetStarByID возвращает опубликованную звезду по ID через ORM.
// Удалённые звёзды и черновики через этот метод недоступны для просмотра.
func (r *Repository) GetStarByID(id int) (*ds.Star, error) {
	var star ds.Star
	err := r.db.Preload("Likes").
		Where("id = ? AND status = ?", id, ds.StatusPublished).
		First(&star).Error
	if err != nil {
		return nil, err
	}
	return &star, nil
}

// GetStarByIDCursor демонстрирует получение записи через курсор (Raw SQL)
func (r *Repository) GetStarByIDCursor(id int) (*ds.Star, error) {
	query := "SELECT id, name, description, status, image_url, video_url, parallax, distance, date_create, creator_id FROM stars WHERE id = $1 AND status = 'published'"
	row := r.db.Raw(query, id).Row()

	var star ds.Star
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

// GetDraftByCreator находит черновик конкретного пользователя через ORM
func (r *Repository) GetDraftByCreator(creatorID uint) (*ds.Star, error) {
	var star ds.Star
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&star).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &star, nil
}

// CreateDraft создаёт новую карточку в статусе 'draft' через ORM
func (r *Repository) CreateDraft(creatorID uint, name string) (*ds.Star, error) {
	draft := ds.Star{
		Name:        name,
		Status:      ds.StatusDraft,
		CreatorID:   creatorID,
		DateCreate:  time.Now(),
		Description: "",
		ImageURL:    "",
		VideoURL:    "",
		Parallax:    0,
		Distance:    0,
	}

	err := r.db.Create(&draft).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка создания черновика: %w", err)
	}

	return &draft, nil
}

// PublishStar публикует карточку через ORM, обновляя описание, параллакс, расстояние и дату формирования
func (r *Repository) PublishStar(starID uint, description string, parallax float64, distance float64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"description": description,
		"parallax":    parallax,
		"distance":    distance,
		"status":      ds.StatusPublished,
		"date_finish": sql.NullTime{Time: now, Valid: true},
	}

	err := r.db.Model(&ds.Star{}).Where("id = ?", starID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("ошибка публикации звезды: %w", err)
	}

	return nil
}

// DeleteStar выполняет логическое удаление услуги (статус меняется на 'deleted')
// с помощью выполнения SQL запроса UPDATE, БЕЗ ORM.
func (r *Repository) DeleteStar(starID uint) error {
	query := "UPDATE stars SET status = $1 WHERE id = $2"
	result := r.db.Exec(query, ds.StatusDeleted, starID)
	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении звезды с id %d: %w", starID, result.Error)
	}
	return nil
}
