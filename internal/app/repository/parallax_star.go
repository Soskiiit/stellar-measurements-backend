package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"stellar-measurements-backend/internal/app/ds"

	"gorm.io/gorm"
)

func (r *Repository) GetPublishedParallaxStars() ([]ds.ParallaxStar, error) {
	var stars []ds.ParallaxStar
	err := r.db.Preload("Likes").
		Where("status = ?", ds.StatusPublished).
		Order("id ASC").
		Find(&stars).Error
	if err != nil {
		return nil, err
	}
	return stars, nil
}

func (r *Repository) GetPublishedStars() ([]ds.ParallaxStar, error) {
	return r.GetPublishedParallaxStars()
}

func (r *Repository) GetFirstPublishedParallaxStar() (*ds.ParallaxStar, error) {
	var star ds.ParallaxStar
	err := r.db.Preload("Likes").
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

func (r *Repository) GetParallaxStarsByDistance(maxDistance float64) ([]ds.ParallaxStar, error) {
	var stars []ds.ParallaxStar
	err := r.db.Preload("Likes").
		Where("status = ? AND distance <= ?", ds.StatusPublished, maxDistance).
		Order("distance ASC").
		Find(&stars).Error
	if err != nil {
		return nil, err
	}
	return stars, nil
}

func (r *Repository) GetStarsByDistance(maxDistance float64) ([]ds.ParallaxStar, error) {
	return r.GetParallaxStarsByDistance(maxDistance)
}

func (r *Repository) GetParallaxStarByID(id int) (*ds.ParallaxStar, error) {
	var star ds.ParallaxStar
	err := r.db.Preload("Likes").
		Where("id = ? AND status = ?", id, ds.StatusPublished).
		First(&star).Error
	if err != nil {
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
	var stars []ds.ParallaxStar
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).Limit(1).Find(&stars).Error
	if err != nil {
		return nil, err
	}
	if len(stars) == 0 {
		return nil, nil
	}
	return &stars[0], nil
}

func (r *Repository) CreateDraft(creatorID uint, name string) (*ds.ParallaxStar, error) {
	draft := ds.ParallaxStar{
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
