# Развёртывание

Весь сервис поднимается из `docker-compose.yml` в этой папке. Снаружи открыты только порты 80 и 443 контейнера Caddy:

```
Телефон студента ──┐
                   ├─► https://домен ─► caddy
Браузер препода ───┘                     ├─ /api/*     ─► backend ─► postgres
                                         └─ остальное  ─► frontend (Next.js)
```

Сайт и API работают на одном адресе, поэтому CORS на сервере не участвует. `migrate` накатывает новые миграции перед каждым запуском бэкенда и завершается.

## Что нужно заранее

- Сервер с Docker и публичным IPv4, открытые порты 80 и 443.
- Домен с A-записью на IP сервера. Без неё Caddy не получит сертификат.

## Первый запуск

Рядом должны лежать две ветки репозитория:

```bash
git clone -b backend  https://github.com/Psychoz42/TPUAttendanceControl.git TPUAttendanceControlBackend
git clone -b frontend https://github.com/Psychoz42/TPUAttendanceControl.git TPUAttendanceControlFrontend
cd TPUAttendanceControlBackend/deploy

cp .env.example .env     # задать DOMAIN и секреты: openssl rand -hex 32
nano users.txt           # только демо-аккаунты: пароли лежат открытым текстом
chmod 644 users.txt      # контейнер читает файл не от root
chmod 600 .env

docker compose up -d --build
docker compose ps -a     # migrate должен быть exited (0), остальные running
```

Проверка:

```bash
curl -s https://<домен>/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"x","password":"x"}'
# ожидается {"success":false,"code":"LOGIN_ERROR",...}
```

Открытая страница `https://<домен>/` должна показать вход преподавателя.

## Обновление

```bash
git -C ../../TPUAttendanceControlFrontend pull
git pull
docker compose up -d --build
```

После правки `users.txt`: `docker compose restart backend`. После правки `Caddyfile`: `docker compose restart caddy`.

## Важно

- Смена `POSTGRES_PASSWORD` в `.env` не меняет пароль в уже созданной базе, его нужно менять через `ALTER USER`.
- Смена `JWT_SECRET` разлогинивает всех.
- `EVENT_RETENTION_DAYS=0` хранит занятия бессрочно. Положительное значение удаляет занятия и списки безвозвратно.
- Перед `docker compose down -v` подумайте: флаг `-v` удаляет базу вместе со всеми данными.
- Логи ограничены пятью файлами по 50 МБ на контейнер.

## Сервер в России

Docker Hub, npm и прокси модулей Go могут быть недоступны или медленны. Если сборка или `docker pull` падают по сети, варианты такие: зеркало реестра в `/etc/docker/daemon.json`, `GOPROXY` при сборке бэкенда, сборка образов на другой машине и перенос через `docker save | ssh <сервер> docker load`.
