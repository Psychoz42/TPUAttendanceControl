# TPUAttendanceControlBackend

Бэкенд системы учёта посещаемости ТПУ на QR-кодах: Go, Fiber v2, PostgreSQL.

## Запуск

```bash
cp .env.example .env        # задать POSTGRES_PASSWORD
docker compose up -d        # PostgreSQL 16 на 127.0.0.1:5431
migrate -path migrations -database "postgresql://postgres:<пароль>@localhost:5431/postgres?sslmode=disable" up
DATABASE_URL="postgresql://postgres:<пароль>@localhost:5431/postgres" go run .
```

Сервер читает пользователей из `users.txt` (формат: `username password email Фамилия-Имя-Отчество группа`) и раздаёт собранный фронтенд из `./dist`. Оба файла в репозиторий не входят.

## Переменные окружения

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `DATABASE_URL` | `postgresql://postgres:2367@localhost:5431/postgres` | строка подключения к БД |
| `LISTEN_ADDR` | `:80` | адрес HTTP-сервера |
| `USERS_FILE` | `users.txt` | файл пользователей |
| `TPU_API_URL` | — | адрес реального API ТПУ (сейчас не используется) |

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
