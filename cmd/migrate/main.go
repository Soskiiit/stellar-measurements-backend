package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"stellar-measurements-backend/internal/app/ds"
	"stellar-measurements-backend/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	dsnStr := dsn.FromEnv()
	logrus.Infof("Connecting to DB: %s", dsnStr)

	db, err := gorm.Open(postgres.Open(dsnStr), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("Failed to connect database: %v", err)
	}

	// Миграция схем таблиц (каскадное удаление запрещено)
	err = db.AutoMigrate(
		&ds.User{},
		&ds.Star{},
		&ds.StarLike{},
	)
	if err != nil {
		logrus.Fatalf("Cannot migrate DB: %v", err)
	}

	// Явное приведение типов колонок при наличии существующих данных
	db.Exec("ALTER TABLE stars ALTER COLUMN description TYPE varchar(255);")
	db.Exec("ALTER TABLE stars ALTER COLUMN parallax TYPE numeric(4,3) USING parallax::numeric(4,3);")
	db.Exec("ALTER TABLE stars ALTER COLUMN distance TYPE numeric(5,2) USING distance::numeric(5,2);")

	// Ограничения целостности на неотрицательные значения (CHECK constraints)
	db.Exec("ALTER TABLE stars DROP CONSTRAINT IF EXISTS chk_stars_parallax_non_negative;")
	db.Exec("ALTER TABLE stars ADD CONSTRAINT chk_stars_parallax_non_negative CHECK (parallax >= 0);")
	db.Exec("ALTER TABLE stars DROP CONSTRAINT IF EXISTS chk_stars_distance_non_negative;")
	db.Exec("ALTER TABLE stars ADD CONSTRAINT chk_stars_distance_non_negative CHECK (distance >= 0);")

	// Создаём частичный уникальный индекс: не более одной услуги в статусе черновик (draft) у каждого пользователя
	db.Exec("DROP INDEX IF EXISTS idx_stars_creator_draft;")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_stars_creator_draft ON stars (creator_id) WHERE status = 'draft';")

	logrus.Info("Database schema migrated successfully")

	// Наполнение начальными данными при их отсутствии
	seedData(db)
}

func seedData(db *gorm.DB) {
	var userCount int64
	db.Model(&ds.User{}).Count(&userCount)
	if userCount == 0 {
		users := []ds.User{
			{ID: 1, Login: "user1", Password: "password1", IsModerator: false},
			{ID: 2, Login: "user2", Password: "password2", IsModerator: true},
			{ID: 3, Login: "user3", Password: "password3", IsModerator: false},
			{ID: 4, Login: "user4", Password: "password4", IsModerator: false},
			{ID: 5, Login: "user5", Password: "password5", IsModerator: false},
			{ID: 6, Login: "user6", Password: "password6", IsModerator: false},
			{ID: 7, Login: "user7", Password: "password7", IsModerator: false},
			{ID: 8, Login: "user8", Password: "password8", IsModerator: false},
			{ID: 9, Login: "user9", Password: "password9", IsModerator: false},
		}
		if err := db.Create(&users).Error; err != nil {
			logrus.Errorf("Error seeding users: %v", err)
		} else {
			logrus.Info("Users seeded successfully")
		}
		// Обновляем sequence для users_id_seq
		db.Exec("SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));")
	}

	var starCount int64
	db.Model(&ds.Star{}).Count(&starCount)
	if starCount == 0 {
		now := time.Now()
		past := now.Add(-48 * time.Hour)

		stars := []ds.Star{
			{
				ID:          1,
				Name:        "Проксима Центавра",
				Parallax:    0.768,
				Distance:    1.30,
				Description: "Ближайшая к Солнцу звезда. Красный карлик спектрального класса M5.5V. Часть тройной системы Альфа Центавра.",
				Status:      ds.StatusPublished,
				ImageURL:    "http://localhost:9002/stars/Proxima%20Centauri.jpg",
				VideoURL:    "http://localhost:9002/stars/Proxima%20Centauri.mp4",
				DateCreate:  past,
				DateFinish:  sql.NullTime{Time: past.Add(time.Hour), Valid: true},
				CreatorID:   1,
			},
			{
				ID:          2,
				Name:        "Бетельгейзе",
				Parallax:    0.004,
				Distance:    222.22,
				Description: "Красный сверхгигант в созвездии Ориона. Одна из самых крупных известных звёзд. Спектральный класс M1-2.",
				Status:      ds.StatusPublished,
				ImageURL:    "http://localhost:9002/stars/Betelgeuse.jpg",
				VideoURL:    "http://localhost:9002/stars/Betelgeuse.mp4",
				DateCreate:  past,
				DateFinish:  sql.NullTime{Time: past.Add(2 * time.Hour), Valid: true},
				CreatorID:   1,
			},
			{
				ID:          3,
				Name:        "Альфа Центавра A",
				Parallax:    0.754,
				Distance:    1.33,
				Description: "Главная звезда тройной системы Альфа Центавра. Спектральный класс G2V, аналог Солнца. Расстояние около 4.37 световых лет.",
				Status:      ds.StatusPublished,
				ImageURL:    "http://localhost:9002/stars/AlphaCentauri.jpg",
				VideoURL:    "http://localhost:9002/stars/AlphaCentauri.mp4",
				DateCreate:  past,
				DateFinish:  sql.NullTime{Time: past.Add(3 * time.Hour), Valid: true},
				CreatorID:   1,
			},
			{
				ID:          4,
				Name:        "Вега",
				Parallax:    0.130,
				Distance:    7.68,
				Description: "Ярчайшая звезда созвездия Лиры. Спектральный класс A0V. Используется как стандарт нуля цветности.",
				Status:      ds.StatusDraft,
				ImageURL:    "http://localhost:9002/stars/Vega.jpg",
				VideoURL:    "http://localhost:9002/stars/Vega.mp4",
				DateCreate:  now,
				DateFinish:  sql.NullTime{Valid: false},
				CreatorID:   1,
			},
			{
				ID:          5,
				Name:        "Полярная звезда",
				Parallax:    0.008,
				Distance:    131.58,
				Description: "Осьми переменная цефеида в созвездии Малой Медведицы. Указывает направление на северный полюс мира. Спектральный класс F7Ib.",
				Status:      ds.StatusPublished,
				ImageURL:    "http://localhost:9002/stars/Polaris.jpg",
				VideoURL:    "http://localhost:9002/stars/Polaris.mp4",
				DateCreate:  past,
				DateFinish:  sql.NullTime{Time: past.Add(4 * time.Hour), Valid: true},
				CreatorID:   2,
			},
			{
				ID:          6,
				Name:        "Альтаир",
				Parallax:    0.199,
				Distance:    5.03,
				Description: "Звезда в созвездии Орла. Спектральный класс A7V. Быстро вращается вокруг оси. Одна из ближайших видимых невооружённым глазом звёзд.",
				Status:      ds.StatusDeleted,
				ImageURL:    "http://localhost:9002/stars/Altair.jpg",
				VideoURL:    "http://localhost:9002/stars/Altair.mp4",
				DateCreate:  past,
				DateFinish:  sql.NullTime{Time: past.Add(5 * time.Hour), Valid: true},
				CreatorID:   1,
			},
		}

		if err := db.Create(&stars).Error; err != nil {
			logrus.Errorf("Error seeding stars: %v", err)
		} else {
			logrus.Info("Stars seeded successfully")
		}
		// Обновляем sequence для stars_id_seq
		db.Exec("SELECT setval('stars_id_seq', (SELECT MAX(id) FROM stars));")
	}

	var likeCount int64
	db.Model(&ds.StarLike{}).Count(&likeCount)
	if likeCount == 0 {
		likes := []ds.StarLike{
			// Проксима Центавра (user1, user4)
			{ID: 1, UserID: 1, StarID: 1},
			{ID: 2, UserID: 4, StarID: 1},

			// Бетельгейзе (user2, user5, user6)
			{ID: 3, UserID: 2, StarID: 2},
			{ID: 4, UserID: 5, StarID: 2},
			{ID: 5, UserID: 6, StarID: 2},

			// Альфа Центавра A (user1, user2)
			{ID: 6, UserID: 1, StarID: 3},
			{ID: 7, UserID: 2, StarID: 3},

			// Полярная звезда (user3, user7, user8, user9)
			{ID: 8, UserID: 3, StarID: 5},
			{ID: 9, UserID: 7, StarID: 5},
			{ID: 10, UserID: 8, StarID: 5},
			{ID: 11, UserID: 9, StarID: 5},
		}
		if err := db.Create(&likes).Error; err != nil {
			logrus.Errorf("Error seeding star likes: %v", err)
		} else {
			logrus.Info("Star likes seeded successfully")
		}
		// Обновляем sequence для star_likes_id_seq
		db.Exec("SELECT setval('star_likes_id_seq', (SELECT MAX(id) FROM star_likes));")
	}

	fmt.Println("Database migration and initial seed completed.")
}
