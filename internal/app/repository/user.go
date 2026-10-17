package repository

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"

	"stellar-measurements-backend/internal/app/ds"
	"stellar-measurements-backend/internal/app/role"

	"gorm.io/gorm"
)

// GenerateHashString хэширует строку с помощью sha1 согласно методическим указаниям
func GenerateHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

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
		Password:    GenerateHashString(password),
		Role:        role.Astronomer,
		IsModerator: false,
	}

	if err := r.db.Create(&newUser).Error; err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return &newUser, nil
}

func (r *Repository) AuthenticateUser(login string, password string) (*ds.User, error) {
	var user ds.User
	hashedPass := GenerateHashString(password)
	err := r.db.Where("login = ? AND (password = ? OR password = ?)", login, password, hashedPass).First(&user).Error
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

func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("пользователь с логином '%s' не найден", login)
		}
		return nil, err
	}
	return &user, nil
}
