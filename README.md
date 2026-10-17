# stellar-measurements-backend

## Лабораторная 3 — веб-сервис REST API

### Домен «Услуга» (параллаксная звезда)

- `GET /api/parallax_stars` — список опубликованных услуг; необязательный query-фильтр `distance` (максимальное расстояние в парсеках по теме). Ответ 200: массив объектов с признаками `is_creator` (1, если создатель — текущий пользователь) и количеством лайков `likes_count`.
- `GET /api/parallax_stars/feed` — лента: первая опубликованная услуга, id не указывается. Ответ 200: объект с признаками `is_creator` и `is_liked` (1, если текущий пользователь поставил лайк).
- `GET /api/parallax_stars/feed/:id` — опубликованная услуга по id; 404, если не существует или удалена.
- `GET /api/parallax_stars/feed/:id?next=true` — следующая опубликованная после id, в конце ленты заворот на первую.
- `GET /api/parallax_stars/draft` — единственный черновик текущего пользователя, id с клиента не передаётся; 404, если черновика нет.
- `POST /api/parallax_stars` — создание черновика; принимает `name` (текст), опциональные `description`, `parallax`, `distance` и файлы `image`, `video` (multipart/form-data, именно файлы, а не URL). Файлы сохраняются в MinIO под сгенерированными латинскими именами, URL пишутся в поля БД. Ответ 201 + объект черновика.
- `PUT /api/parallax_stars/:id/publish` — публикация своего черновика (смена статуса); JSON-тело: `description`, `parallax`, `distance`. Ответ 200 + обновлённый объект; 404, если черновик чужой или уже опубликован.
- `DELETE /api/parallax_stars/:id` — логическое удаление своей услуги (status = 'deleted', SQL UPDATE через курсор). Ответ 200; чужая услуга — 403.
- `POST /api/parallax_stars/:id/like` — лайк текущего пользователя; JSON `{"like": 1}` ставит, `{"like": 0}` отменяет. Ответ 200 + `{"is_liked": ..., "likes_count": ...}`.

### Домен «Пользователь»

- `POST /api/users/register` — регистрация; JSON `{login, password}`. Ответ 201 + `{id, login}` (пароль не сериализуется); 409, если логин занят.
- `POST /api/users/login` — заглушка аутентификации для ЛР4. Ответ 200.
- `POST /api/users/logout` — заглушка деавторизации для ЛР4. Ответ 200.

### Соглашения ответов

- Набор полей JSON всегда одинаковый; незаполненные значения приходят как `null` / `""`.
- Статус услуги в ответах не передаётся (хранится только в БД); коды состояния в тело не дублируются.
- Успех: 200 / 201; ошибки: 400, 403, 404, 409 без избыточных обёрток.

## Таблицы базы данных

### users — зарегистрированные пользователи
| Поле | Тип | Описание |
|---|---|---|
| id | SERIAL, PK | первичный ключ |
| login | VARCHAR(50), UNIQUE, NOT NULL | логин |
| password | VARCHAR(100), NOT NULL | пароль (в ЛР4 — хэш) |
| is_moderator | BOOLEAN, NOT NULL, DEFAULT false | признак роли модератора |

### parallax_stars — черновики, опубликованные и удалённые услуги
| Поле | Тип | Описание |
|---|---|---|
| id | SERIAL, PK | первичный ключ |
| status | VARCHAR(20), NOT NULL, DEFAULT 'draft' | draft / published / deleted |
| name | VARCHAR(100), NOT NULL | наименование звезды |
| description | VARCHAR(255), DEFAULT '' | астрофизическое описание |
| image_url | VARCHAR(255), DEFAULT '' | ссылка/имя изображения в MinIO |
| video_url | VARCHAR(255), DEFAULT '' | ссылка/имя видео в MinIO |
| parallax | NUMERIC(4,3), NULL | годичный параллакс p в секундах дуги (поле по теме) |
| distance | NUMERIC(5,2), NULL | расстояние d = 1 / p в парсеках (поле по теме) |
| creator_id | INTEGER, NOT NULL, FK → users.id | создатель (ON DELETE RESTRICT) |
| date_create | TIMESTAMPTZ, NOT NULL | дата создания черновика |
| date_finish | TIMESTAMPTZ, NULL | дата формирования / публикации |

### parallax_star_likes — связь многие-ко-многим пользователи ↔ услуги
| Поле | Тип | Описание |
|---|---|---|
| id | SERIAL, PK | первичный ключ |
| user_id | INTEGER, FK → users.id | кто лайкнул (ON DELETE RESTRICT) |
| parallax_star_id | INTEGER, FK → parallax_stars.id | что лайкнули (ON DELETE RESTRICT) |
| — | UNIQUE(user_id, parallax_star_id) | один лайк на пару |