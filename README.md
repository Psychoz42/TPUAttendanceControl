# TPUAttendanceControlBackend

Бэкенд системы учёта посещаемости ТПУ на QR-кодах: Go, Fiber v2, PostgreSQL.

## Запуск

```bash
cp .env.example .env        # задать POSTGRES_PASSWORD
docker compose up -d        # PostgreSQL 16 на 127.0.0.1:5431
migrate -path migrations -database "postgresql://postgres:<пароль>@localhost:5431/postgres?sslmode=disable" up
JWT_SECRET="$(openssl rand -base64 48)" DATABASE_URL="postgresql://postgres:<пароль>@localhost:5431/postgres" go run .
```

Сервер читает пользователей из `users.txt` (формат: `username password email Фамилия-Имя-Отчество группа [роль]`, роль `student` или `teacher`, без неё `student`). Файл в репозиторий не входит. Развёртывание на сервере описано в [deploy/README.md](deploy/README.md).

## Переменные окружения

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `JWT_SECRET` | — | секрет подписи токенов, не короче 32 байт. Без него сервер не стартует. При смене все выданные токены перестают действовать |
| `AUTH_REQUIRED` | `false` | `true`: все запросы к `/attendance/*` требуют токен |
| `DATABASE_URL` | `postgresql://postgres:2367@localhost:5431/postgres` | строка подключения к БД |
| `LISTEN_ADDR` | `:80` | адрес HTTP-сервера |
| `USERS_FILE` | `users.txt` | файл пользователей |
| `PROXY_HEADER` | — | заголовок с IP клиента, например `X-Forwarded-For`. Задавать, только если сервер доступен исключительно через прокси, иначе клиент подделает свой IP |
| `EVENT_RETENTION_DAYS` | `0` | через сколько дней удалять занятия вместе со списками, `0` значит хранить бессрочно |
| `TPU_API_URL` | — | адрес реального API ТПУ (сейчас не используется) |

## Авторизация

`POST /api/v1/auth/login` принимает `{"email", "password"}`. К прежним полям ответа добавились `role` (`student` или `teacher`) и `token`:

```json
{"username": "...", "fullName": "...", "email": "...", "group": "...", "role": "teacher", "token": "<JWT>"}
```

Токен живёт 12 часов, обновления нет: после истечения нужно войти заново. Его нужно отправлять во всех запросах к `/api/v1/attendance/*`:

```
Authorization: Bearer <token>
```

Пути и остальные поля запросов не меняются. С токеном сервер берёт пользователя из него, а не из тела запроса:

| Эндпоинт | Роль | Что меняет токен |
|---|---|---|
| `scan` | `student` | отмечается владелец токена, `username` в теле не нужен и игнорируется |
| `createEvent` | `teacher` | владелец события из токена, `event_owner` в теле игнорируется |
| `getOwnerEvents` | `teacher` | события владельца токена, `owner` в теле игнорируется |
| `getEvent`, `disableEventUpdating`, `deleteEvent` | `teacher` | только свои события, на чужое `EVENT_NOT_FOUND` |
| `updateCode` | `teacher` | только свои события, на чужое `INVALID_QR` |

Новые ошибки в формате `{"success": false, "code", "message"}`. В отличие от остальных ошибок API, они приходят с HTTP-статусом не 200:

| HTTP | `code` | Когда | Что делать клиенту |
|---|---|---|---|
| 401 | `UNAUTHORIZED` | токен неверный или просрочен; при `AUTH_REQUIRED=true` ещё и если его нет | войти заново |
| 403 | `FORBIDDEN` | у роли нет доступа к эндпоинту | показать сообщение |

Пока `AUTH_REQUIRED=false`, запрос без токена работает по-старому: данные берутся из тела, права не проверяются. Это нужно только на время перехода клиентов на токены. Неверный токен отклоняется при любом значении флага.

## Занятия

Закрытые занятия и списки отметившихся остаются на сервере на срок `EVENT_RETENTION_DAYS`. Если сайт перестал обновлять код, занятие закрывается автоматически через 10–20 минут после истечения кода, и продолжить его уже нельзя.

`getOwnerEvents` отдаёт занятия от новых к старым:

```json
[{"event_id": 1, "event_name": "...", "event_time": "08:30", "created_at": "2026-11-07T08:25:00Z", "active": true, "noted_count": 17}]
```

`POST /api/v1/attendance/getEvent` принимает `{"event_id"}` и отдаёт текущее состояние занятия. Он нужен, чтобы вернуться к занятию после обновления страницы или потери ответа `updateCode` и чтобы получить итоговый список после закрытия. Поле `code` есть только у открытого занятия, `noted_students` всегда массив:

```json
{"event_id": 1, "event_name": "...", "event_time": "08:30", "created_at": "2026-11-07T08:25:00Z", "active": true, "code": "...", "noted_students": ["Иванов Иван Иванович"]}
```

## Тесты

```bash
go test ./...
TEST_DATABASE_URL="postgresql://postgres:<пароль>@localhost:5431/postgres" go test ./...
```

Без `TEST_DATABASE_URL` интеграционные тесты из `main_test.go` пропускаются. Они создают события с владельцем `integration-test-owner` и удаляют их за собой.

## Сборка для сервера

```bash
GOOS=linux GOARCH=amd64 go build -o tpu_att .
```
