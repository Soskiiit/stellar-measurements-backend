package ds

import (
	"stellar-measurements-backend/internal/app/role"
)

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Login       string    `gorm:"type:varchar(50);unique;not null" json:"login"`
	Password    string    `gorm:"type:varchar(100);not null" json:"-"`
	Role        role.Role `gorm:"type:int;not null;default:0" json:"role"`
	IsModerator bool      `gorm:"type:boolean;not null;default:false" json:"is_moderator"`
}

func (User) TableName() string {
	return "users"
}
