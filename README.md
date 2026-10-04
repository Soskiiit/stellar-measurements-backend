# Определение расстояния до звезды в Млечном пути по ее годовому параллаксу

**Веб-сервис бэкенда для заявочной системы измерения параметров звезд.**  
Выполнено в рамках Лабораторной работы №3 по курсу «Разработка веб-сервисов».

---

## 1. Предметная область и сущности

- **Предметная область:** Определение расстояния до звезды в Млечном пути по ее годовому параллаксу ($d = 1 / p$, где $p$ — параллакс в секундах дуги, $d$ — расстояние в парсеках).
- **Услуги (Services):** Звёзды (`parallax_stars`).
- **Заявка:** Расчёт расстояния с указанием годичного параллакса по каждой звезде.
- **Пользователи (Users):** Учетные записи исследователей и астрономов (`users`).
- **Лайки / Закладки (Likes):** Отметки пользователей для звёзд (`parallax_star_likes`).

---

## 2. Структура базы данных (PostgreSQL)

### Таблица `users` (Пользователи)
| Поле | Тип данных | Ограничения | Описание |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` / `uint` | `PRIMARY KEY` | Уникальный идентификатор пользователя |
| `login` | `VARCHAR(50)` | `NOT NULL, UNIQUE` | Логин пользователя |
| `password` | `VARCHAR(100)` | `NOT NULL` | Хеш пароля (скрыт из сериализации `json:"-"`) |
| `is_moderator` | `BOOLEAN` | `NOT NULL, DEFAULT false` | Признак роли модератора |

### Таблица `parallax_stars` (Услуги — Звёзды)
| Поле | Тип данных | Ограничения | Описание |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` / `uint` | `PRIMARY KEY` | Уникальный идентификатор карточки звезды |
| `name` | `VARCHAR(100)` | `NOT NULL` | Наименование звезды (напр., *Проксима Центавра*) |
| `description` | `VARCHAR(255)` | `DEFAULT ''` | Краткое астрофизическое описание звезды |
| `status` | `VARCHAR(20)` | `NOT NULL, DEFAULT 'draft'` | Статус услуги: `draft` (черновик), `published` (опубликован), `deleted` (удалён) |
| `image_url` | `VARCHAR(255)` | `NOT NULL, DEFAULT ''` | Ссылка на изображение звезды в MinIO / статике |
| `video_url` | `VARCHAR(255)` | `NOT NULL, DEFAULT ''` | Ссылка на видеоролик звезды в MinIO / статике |
| `parallax` | `NUMERIC(4,3)` | `CHECK (parallax >= 0)` | Годичный параллакс звезды ($p$, в секундах дуги) |
| `distance` | `NUMERIC(5,2)` | `CHECK (distance >= 0)` | Расстояние до звезды ($d = 1 / p$, в парсеках) |
| `date_create` | `TIMESTAMPTZ` | `NOT NULL` | Дата и время создания записи (вычисляется сервером) |
| `date_finish` | `TIMESTAMPTZ` | `DEFAULT NULL` | Дата и время публикации/формирования заявки |
| `creator_id` | `BIGINT` | `NOT NULL, REFERENCES users(id)` | Внешний ключ: создатель услуги |

*Ограничения:*
- Каскадное удаление запрещено (`RESTRICT`).
- Для каждого пользователя разрешено не более 1 активного черновика (`idx_parallax_stars_creator_draft ON parallax_stars(creator_id) WHERE status = 'draft'`).

### Таблица `parallax_star_likes` (Связь M-M: Лайки звезд)
| Поле | Тип данных | Ограничения | Описание |
| :--- | :--- | :--- | :--- |
| `id` | `BIGSERIAL` / `uint` | `PRIMARY KEY` | Идентификатор записи лайка |
| `user_id` | `BIGINT` | `NOT NULL, REFERENCES users(id)` | Внешний ключ на пользователя |
| `parallax_star_id` | `BIGINT` | `NOT NULL, REFERENCES parallax_stars(id)` | Внешний ключ на звезду |

*Ограничения:*
- Уникальный составной индекс `idx_user_parallax_star_like (user_id, parallax_star_id)` исключает повторные лайки.

---

## 3. Функция-Singleton текущего пользователя

В соответствии с методическими указаниями к ЛР3, пользователь-создатель зафиксирован константой через потокобезопасную функцию-синглтон (`sync.Once`).

Файл: `internal/app/session/session.go`:
```go
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
```

### Использование синглтона в методах:
- `GetParallaxStarsAPI`: вычисляет признак `is_creator` (0 или 1), сравнивая `star.CreatorID == session.GetCurrentUserID()`.
- `GetParallaxStarFeedAPI`: передает ID текущего пользователя для вычисления `is_creator` и `is_liked`.
- `GetDraftAPI`: находит черновик именно для пользователя `session.GetCurrentUserID()`.
- `AddParallaxStarAPI`: устанавливает `CreatorID` новой звезды равным `session.GetCurrentUserID()`.
- `PublishParallaxStarAPI`: запрещает публикацию чужой звезды (`star.CreatorID != session.GetCurrentUserID()`).
- `DeleteParallaxStarAPI`: разрешает удаление только звезд текущего пользователя (`star.CreatorID != session.GetCurrentUserID()`).
- `LikeParallaxStarAPI`: привязывает установку/снятие лайка к `session.GetCurrentUserID()`.

---

## 4. Описание REST API методов (`/api`)

Все методы сервиса расположены под префиксом `/api` и возвращают ответы в формате JSON.

| № | Метод | URL | Назначение |
| :-: | :--- | :--- | :--- |
| 1 | `GET` | `/api/parallax_stars` | Список опубликованных звезд (с фильтром по `distance` и полем `is_creator`) |
| 2 | `GET` | `/api/parallax_stars/feed` | Лента опубликованных звезд (первая звезда) |
| 3 | `GET` | `/api/parallax_stars/feed/:id?next=true` | Переход к следующей опубликованной звезде в ленте |
| 4 | `GET` | `/api/parallax_stars/draft` | Получение активного черновика текущего пользователя |
| 5 | `POST` | `/api/parallax_stars` | Добавление новой звезды (черновик) с загрузкой фото и видео в MinIO |
| 6 | `PUT` | `/api/parallax_stars/:id/publish` | Публикация черновика (смена статуса на `published`) |
| 7 | `DELETE` | `/api/parallax_stars/:id` | Логическое удаление услуги (soft delete) только своего черновика/звезды |
| 8 | `POST` | `/api/parallax_stars/:id/like` | Установка (`like=1`) или снятие (`like=0`) лайка текущим пользователем |
| 9 | `POST` | `/api/users/register` | Регистрация нового пользователя |
| 10 | `POST` | `/api/users/login` | Аутентификация (заглушка для ЛР4, выдача JWT токена) |
| 11 | `POST` | `/api/users/logout` | Деавторизация (заглушка для ЛР4) |
| 12 | `GET` | `/api/users/:id` | Профиль пользователя со списком его звезд (вложенная сериализация) |

---

### Детальная спецификация эндпоинтов

#### 1. GET `/api/parallax_stars`
- **Параметры query:**
  - `distance` (float, опционально) — максимальное расстояние в парсеках для фильтрации.
- **Статус ответа:** `200 OK`.
- **Пример ответа:**
```json
[
  {
    "id": 1,
    "name": "Проксима Центавра",
    "description": "Ближайшая к Солнцу звезда.",
    "status": "published",
    "image_url": "http://localhost:9002/parallax-stars/star1.jpg",
    "video_url": "http://localhost:9002/parallax-stars/star1.mp4",
    "parallax": 0.768,
    "distance": 1.30,
    "likes_count": 2,
    "is_creator": 1
  }
]
```

#### 2. POST `/api/parallax_stars`
- **Content-Type:** `multipart/form-data`.
- **Поля формы:**
  - `name` (string, обязательно) — наименование звезды.
  - `description` (string) — описание.
  - `parallax` (float) — параллакс в секундах дуги.
  - `distance` (float) — расстояние в парсеках.
  - `image` (file) — файл изображения (загружается в бакет `parallax-stars` в MinIO).
  - `video` (file) — файл видеоролика (загружается в бакет `parallax-stars` в MinIO).
- **Статусы ответа:** `201 Created` при успехе, `400 Bad Request` при ошибке валидации.

#### 3. GET `/api/parallax_stars/draft`
- **Статус ответа:** `200 OK` (если есть активный черновик), `404 Not Found` (если черновика нет).
- Возвращает полную модель со вложенным создателем и списком лайков.

#### 4. PUT `/api/parallax_stars/:id/publish`
- **Тело запроса (JSON):**
```json
{
  "description": "Уточненные астрометрические данные Gaia DR3",
  "parallax": 0.768,
  "distance": 1.30
}
```
- **Правило перехода статусов:** Переход разрешен только из `draft` в `published`. Из `deleted` публикация запрещена. Устанавливается `date_finish = now()`.

#### 5. DELETE `/api/parallax_stars/:id`
- **Логика:** Выполняется soft delete (статус изменяется на `deleted`, физического удаления из базы нет).
- Разрешено только создателю услуги (`creator_id == session.GetCurrentUserID()`).

#### 6. POST `/api/parallax_stars/:id/like`
- **Тело запроса (JSON):**
```json
{
  "like": 1
}
```
- `like: 1` — поставить лайк; `like: 0` — отменить лайк.

#### 7. POST `/api/users/register`
- **Тело запроса (JSON):**
```json
{
  "login": "astronomer1",
  "password": "secretPassword123"
}
```
- **Статус ответа:** `201 Created`. Пароль исключен из ответа (`json:"-"`).

---