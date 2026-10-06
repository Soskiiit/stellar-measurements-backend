package session

import (
	"stellar-measurements-backend/internal/app/ds"
)

const CurrentUserConstantID uint = 1

var CurrentUser = &ds.User{
	ID:          CurrentUserConstantID,
	Login:       "user1",
	IsModerator: false,
}

func GetCurrentUser() *ds.User {
	return CurrentUser
}

func GetCurrentUserID() uint {
	return CurrentUser.ID
}
