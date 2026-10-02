package ds

type ParallaxStarLike struct {
	ID             uint `gorm:"primaryKey" json:"id"`
	UserID         uint `gorm:"not null;uniqueIndex:idx_user_parallax_star_like" json:"user_id"`
	ParallaxStarID uint `gorm:"not null;uniqueIndex:idx_user_parallax_star_like" json:"parallax_star_id"`

	User         User         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user,omitempty"`
	ParallaxStar ParallaxStar `gorm:"foreignKey:ParallaxStarID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"parallax_star,omitempty"`
}

func (ParallaxStarLike) TableName() string {
	return "parallax_star_likes"
}

type StarLike = ParallaxStarLike
