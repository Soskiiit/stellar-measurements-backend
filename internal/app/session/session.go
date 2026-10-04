package session

import (
	"sync"

	"stellar-measurements-backend/internal/app/ds"
)

const CurrentUserConstantID uint = 1

type UserSessionSingleton struct {
	User *ds.User
}

var (
	instance *UserSessionSingleton
	once     sync.Once
)

func GetCurrentUser() *ds.User {
	once.Do(func() {
		instance = &UserSessionSingleton{
			User: &ds.User{
				ID:          CurrentUserConstantID,
				Login:       "user1",
				IsModerator: false,
			},
		}
	})
	return instance.User
}

func GetCurrentUserID() uint {
	return GetCurrentUser().ID
}
