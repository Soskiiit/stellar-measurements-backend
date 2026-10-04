package repository

import (
	"errors"
	"fmt"

	"stellar-measurements-backend/internal/app/ds"

	"gorm.io/gorm"
)

func (r *Repository) RegisterUser(login string, password string) (*ds.User, error) {
	if login == "" || password == "" {
		return nil, fmt.Errorf("логин и пароль обязательны")
	}

	var existing ds.User
	err := r.db.Where("login = ?", login).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("пользователь с логином '%s' уже существует", login)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	newUser := ds.User{
		Login:       login,
		Password:    password,
		IsModerator: false,
	}

	if err := r.db.Create(&newUser).Error; err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return &newUser, nil
}

func (r *Repository) AuthenticateUser(login string, password string) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ? AND password = ?", login, password).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("неверный логин или пароль")
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(id uint) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("пользователь с id=%d не найден", id)
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserWithStars(userID uint) (*ds.UserWithStarsSerializer, error) {
	user, err := r.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	var stars []ds.ParallaxStar
	err = r.db.Preload("Likes").
		Where("creator_id = ? AND status != ?", userID, ds.StatusDeleted).
		Order("id ASC").
		Find(&stars).Error
	if err != nil {
		return nil, err
	}

	starSerializers := make([]ds.ParallaxStarCatalogItemSerializer, 0, len(stars))
	for _, star := range stars {
		starSerializers = append(starSerializers, ds.ParallaxStarCatalogItemSerializer{
			ID:          star.ID,
			Name:        star.Name,
			Description: star.Description,
			Status:      star.Status,
			ImageURL:    star.ImageURL,
			VideoURL:    star.VideoURL,
			Parallax:    star.Parallax,
			Distance:    star.Distance,
			LikesCount:  len(star.Likes),
			IsCreator:   1,
		})
	}

	return &ds.UserWithStarsSerializer{
		User:  ds.ToUserSerializer(user),
		Stars: starSerializers,
	}, nil
}
