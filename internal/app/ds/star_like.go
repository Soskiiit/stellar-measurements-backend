package ds

type StarLike struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"not null;uniqueIndex:idx_user_star_like" json:"user_id"`
	StarID uint `gorm:"not null;uniqueIndex:idx_user_star_like" json:"star_id"`

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user,omitempty"`
	Star Star `gorm:"foreignKey:StarID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"star,omitempty"`
}

func (StarLike) TableName() string {
	return "star_likes"
}
