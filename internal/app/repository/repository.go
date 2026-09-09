package repository

import "fmt"

type Star struct {
	ID          int
	Name        string
	Parallax    float64
	Distance    float64
	Description string
	Status      string
	Likes       []string
	ImageURL    string
	VideoURL    string
}

type Repository struct {
	stars []Star
}

func NewRepository() (*Repository, error) {
	stars := []Star{
		{
			ID:          1,
			Name:        "Проксима Центавра",
			Parallax:    0.768,
			Distance:    1.30,
			Description: "Ближайшая к Солнцу звезда. Красный карлик спектрального класса M5.5V. Часть тройной системы Альфа Центавра.",
			Status:      "published",
			Likes:       []string{"user1", "user4"},
			ImageURL:    "http://localhost:9002/stars/Proxima%20Centauri.jpg",
			VideoURL:    "http://localhost:9002/stars/Proxima%20Centauri.mp4",
		},
		{
			ID:          2,
			Name:        "Бетельгейзе",
			Parallax:    0.0045,
			Distance:    222.22,
			Description: "Красный сверхгигант в созвездии Ориона. Одна из самых крупных известных звёзд. Спектральный класс M1-2.",
			Status:      "published",
			Likes:       []string{"user2", "user5", "user6"},
			ImageURL:    "http://localhost:9002/stars/Betelgeuse.jpg",
			VideoURL:    "http://localhost:9002/stars/Betelgeuse.mp4",
		},
		{
			ID:          3,
			Name:        "Альфа Центавра A",
			Parallax:    0.754,
			Distance:    1.33,
			Description: "Главная звезда тройной системы Альфа Центавра. Спектральный класс G2V, аналог Солнца. Расстояние около 4.37 световых лет.",
			Status:      "published",
			Likes:       []string{"user1", "user2"},
			ImageURL:    "http://localhost:9002/stars/AlphaCentauri.jpg",
			VideoURL:    "http://localhost:9002/stars/AlphaCentauri.mp4",
		},
		{
			ID:          4,
			Name:        "Вега",
			Parallax:    0.130,
			Distance:    7.68,
			Description: "Ярчайшая звезда созвездия Лиры. Спектральный класс A0V. Используется как стандарт нуля цветности.",
			Status:      "draft",
			Likes:       []string{},
			ImageURL:    "http://localhost:9002/stars/Vega.jpg",
			VideoURL:    "http://localhost:9002/stars/Vega.mp4",
		},
		{
			ID:          5,
			Name:        "Полярная звезда",
			Parallax:    0.0076,
			Distance:    131.58,
			Description: "Осьми переменная цефеида в созвездии Малой Медведицы. Указывает направление на северный полюс мира. Спектральный класс F7Ib.",
			Status:      "published",
			Likes:       []string{"user3", "user7", "user8", "user9"},
			ImageURL:    "http://localhost:9002/stars/Polaris.jpg",
			VideoURL:    "http://localhost:9002/stars/Polaris.mp4",
		},
		{
			ID:          6,
			Name:        "Альтаир",
			Parallax:    0.199,
			Distance:    5.03,
			Description: "Звезда в созвездии Орла. Спектральный класс A7V. Быстро вращается вокруг оси. Одна из ближайших видимых невооружённым глазом звёзд.",
			Status:      "deleted",
			Likes:       []string{},
			ImageURL:    "http://localhost:9002/stars/Altair.jpg",
			VideoURL:    "http://localhost:9002/stars/Altair.mp4",
		},
	}

	return &Repository{stars: stars}, nil
}

func (r *Repository) GetStars() ([]Star, error) {
	if len(r.stars) == 0 {
		return nil, fmt.Errorf("коллекция звёзд пуста")
	}
	return r.stars, nil
}

func (r *Repository) GetStarByID(id int) (Star, error) {
	for _, s := range r.stars {
		if s.ID == id {
			return s, nil
		}
	}
	return Star{}, fmt.Errorf("услуга с id %d не найдена", id)
}

func (r *Repository) GetDraft() (Star, error) {
	for _, s := range r.stars {
		if s.Status == "draft" {
			return s, nil
		}
	}
	return Star{}, fmt.Errorf("черновик не найден")
}

func (r *Repository) GetPublishedStars() ([]Star, error) {
	var result []Star
	for _, s := range r.stars {
		if s.Status != "deleted" {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *Repository) GetStarsByDistance(maxDistance float64) ([]Star, error) {
	var result []Star
	for _, s := range r.stars {
		if s.Status != "deleted" && s.Distance <= maxDistance {
			result = append(result, s)
		}
	}
	return result, nil
}
