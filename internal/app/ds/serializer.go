package ds

import (
	"time"
)

type UserSerializer struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

func ToUserSerializer(user *User) UserSerializer {
	if user == nil {
		return UserSerializer{}
	}
	return UserSerializer{
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	}
}

type ParallaxStarCatalogItemSerializer struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	ImageURL    string  `json:"image_url"`
	VideoURL    string  `json:"video_url"`
	Parallax    float64 `json:"parallax"`
	Distance    float64 `json:"distance"`
	LikesCount  int     `json:"likes_count"`
	IsCreator   int     `json:"is_creator"` // 0 или 1
}

type FullParallaxStarSerializer struct {
	ID          uint           `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	ImageURL    string         `json:"image_url"`
	VideoURL    string         `json:"video_url"`
	Parallax    float64        `json:"parallax"`
	Distance    float64        `json:"distance"`
	DateCreate  time.Time      `json:"date_create"`
	DateFinish  *time.Time     `json:"date_finish"`
	CreatorID   uint           `json:"creator_id"`
	Creator     UserSerializer `json:"creator"`
	LikesCount  int            `json:"likes_count"`
	IsCreator   int            `json:"is_creator"` // 0 или 1
	IsLiked     int            `json:"is_liked"`   // 0 или 1
}

func ToFullParallaxStarSerializer(star *ParallaxStar, currentUserID uint) FullParallaxStarSerializer {
	var dateFinish *time.Time
	if star.DateFinish.Valid {
		dateFinish = &star.DateFinish.Time
	}

	isCreator := 0
	if star.CreatorID == currentUserID {
		isCreator = 1
	}

	isLiked := 0
	for _, l := range star.Likes {
		if l.UserID == currentUserID {
			isLiked = 1
			break
		}
	}

	return FullParallaxStarSerializer{
		ID:          star.ID,
		Name:        star.Name,
		Description: star.Description,
		Status:      star.Status,
		ImageURL:    star.ImageURL,
		VideoURL:    star.VideoURL,
		Parallax:    star.Parallax,
		Distance:    star.Distance,
		DateCreate:  star.DateCreate,
		DateFinish:  dateFinish,
		CreatorID:   star.CreatorID,
		Creator:     ToUserSerializer(&star.Creator),
		LikesCount:  len(star.Likes),
		IsCreator:   isCreator,
		IsLiked:     isLiked,
	}
}

type CreateParallaxStarRequest struct {
	Name        string  `form:"name" json:"name" binding:"required"`
	Description string  `form:"description" json:"description"`
	Parallax    float64 `form:"parallax" json:"parallax"`
	Distance    float64 `form:"distance" json:"distance"`
}

type PublishParallaxStarRequest struct {
	Description string   `form:"description" json:"description"`
	Parallax    *float64 `form:"parallax" json:"parallax"`
	Distance    *float64 `form:"distance" json:"distance"`
}

type LikeParallaxStarRequest struct {
	Like *int `form:"like" json:"like" binding:"required"` // 1 - поставить, 0 - отменить
}

type UserRegisterRequest struct {
	Login    string `form:"login" json:"login" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type UserLoginRequest struct {
	Login    string `form:"login" json:"login" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}
