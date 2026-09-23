package ds

import (
	"database/sql"
	"time"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusDeleted   = "deleted"
)

type Star struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(100);not null" json:"name"`
	Description string       `gorm:"type:varchar(255);default:''" json:"description"`
	Status      string       `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	ImageURL    string       `gorm:"type:varchar(255);not null;default:''" json:"image_url"`
	VideoURL    string       `gorm:"type:varchar(255);not null;default:''" json:"video_url"`
	Parallax    float64      `gorm:"type:numeric(4,3);default:0;check:parallax >= 0" json:"parallax"`
	Distance    float64      `gorm:"type:numeric(5,2);default:0;check:distance >= 0" json:"distance"`
	DateCreate  time.Time    `gorm:"not null" json:"date_create"`
	DateFinish  sql.NullTime `gorm:"default:null" json:"date_finish"`
	CreatorID   uint         `gorm:"not null" json:"creator_id"`

	Creator User       `gorm:"foreignKey:CreatorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"creator,omitempty"`
	Likes   []StarLike `gorm:"foreignKey:StarID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"likes,omitempty"`
}

func (Star) TableName() string {
	return "stars"
}
