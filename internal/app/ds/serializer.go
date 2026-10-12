package ds

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
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	ImageURL   string  `json:"image_url"`
	Distance   float64 `json:"distance"`
	LikesCount int     `json:"likes_count"`
	IsCreator  int     `json:"is_creator"`
}

type ParallaxStarFeedSerializer struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	VideoURL    string  `json:"video_url"`
	Parallax    float64 `json:"parallax"`
	Distance    float64 `json:"distance"`
	LikesCount  int     `json:"likes_count"`
	IsLiked     int     `json:"is_liked"`
}

func ToParallaxStarFeedSerializer(star *ParallaxStar, currentUserID uint) ParallaxStarFeedSerializer {
	isLiked := 0
	for _, l := range star.Likes {
		if l.UserID == currentUserID {
			isLiked = 1
			break
		}
	}

	return ParallaxStarFeedSerializer{
		ID:          star.ID,
		Name:        star.Name,
		Description: star.Description,
		ImageURL:    star.ImageURL,
		VideoURL:    star.VideoURL,
		Parallax:    star.Parallax,
		Distance:    star.Distance,
		LikesCount:  len(star.Likes),
		IsLiked:     isLiked,
	}
}

type ParallaxStarDraftSerializer struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	VideoURL    string  `json:"video_url"`
	Parallax    float64 `json:"parallax"`
	Distance    float64 `json:"distance"`
}

func ToParallaxStarDraftSerializer(star *ParallaxStar) ParallaxStarDraftSerializer {
	return ParallaxStarDraftSerializer{
		ID:          star.ID,
		Name:        star.Name,
		Description: star.Description,
		ImageURL:    star.ImageURL,
		VideoURL:    star.VideoURL,
		Parallax:    star.Parallax,
		Distance:    star.Distance,
	}
}

type FullParallaxStarSerializer = ParallaxStarFeedSerializer

func ToFullParallaxStarSerializer(star *ParallaxStar, currentUserID uint) ParallaxStarFeedSerializer {
	return ToParallaxStarFeedSerializer(star, currentUserID)
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
	Like int `json:"like" binding:"oneof=0 1"` // 1 - поставить, 0 - отменить
}

type UserRegisterRequest struct {
	Login    string `form:"login" json:"login" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type UserLoginRequest struct {
	Login    string `form:"login" json:"login" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}
