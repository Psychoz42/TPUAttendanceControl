# Развёртывание

Весь сервис поднимается из `docker-compose.yml` в этой папке. Снаружи открыты только порты 80 и 443 контейнера Caddy:

```
Телефон студента ──┐
                   ├─► https://домен ─► caddy
Браузер препода ───┘                     ├─ /api/*     ─► backend ─► postgres
                                         └─ остальное  ─► frontend (Next.js)
```

Сайт и API работают на одном адресе, поэтому CORS на сервере не участвует. `migrate` накатывает новые миграции перед каждым запуском бэкенда и завершается.

Образы бэкенда и фронтенда собираются на компьютере разработчика и загружаются на сервер готовыми, вместе около 70 МБ. На сервере с 2 ГБ памяти сборка рискованна: `next build` занимает до 1,8 ГБ. Заодно серверу не нужны npm и прокси модулей Go, которые из России могут быть недоступны.

## Как идти по этапам

Этапы выполняются строго по одному. После каждого сверяем вывод с разделом «Готово, если» и только потом идём дальше. Если что-то не сошлось, сначала разбираемся.

Секреты не должны попадать в чат: пароль сервера, файл ключа `tpu_server`, содержимое `.env`. Команды ниже их не выводят.

Перед началом всё должно быть запушено: ветки `backend` и `frontend`, включая эту инструкцию.

Значения для подстановки: `<IP>` — адрес сервера, `<домен>` — домен сервиса.

## Этап 0. Доступ агента к серверу (на компьютере, PowerShell)

Отдельная настройка ключей у хостера не нужна: ключ кладётся на сервер обычной командой, достаточно входа по паролю. Пароль вводит только владелец.

Ключ `$HOME\.ssh\tpu_server` уже создан. Команда спросит пароль сервера:

```powershell
Get-Content "$HOME\.ssh\tpu_server.pub" | ssh root@<IP> "mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
```

Если входа по SSH с паролем нет, только веб-консоль хостера: выведите ключ через `Get-Content "$HOME\.ssh\tpu_server.pub"` и в консоли выполните

```bash
mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '<строка ключа>' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys
```

Создайте файл `$HOME\.ssh\config` через VS Code или Блокнот (через PowerShell 5.1 не надо: появится BOM, и ssh не прочитает файл):

```
Host tpu
    HostName <IP>
    User root
    IdentityFile ~/.ssh/tpu_server
    IdentitiesOnly yes
```

Проверка, на вопрос ответить `yes`:

```powershell
ssh tpu "uname -a"
```

**Готово, если** система вывелась без запроса пароля. Дальше агент сам выполняет команды через `ssh tpu`, и каждую из них показывает на подтверждение.

**Если ключ не заработал**, работаем вручную: владелец вставляет команды этапов в терминал сервера и присылает вывод. Этап 2.2 тогда пропускается, а `scp` будет спрашивать пароль.

## Этап 1. Осмотр сервера

Ничего не меняет.

```bash
. /etc/os-release; echo "$PRETTY_NAME"; uname -m; nproc; free -h; df -h /
ss -tlnp
getent hosts <домен>
for u in https://registry-1.docker.io/v2/ https://mirror.gcr.io/v2/ https://acme-v02.api.letsencrypt.org/directory https://github.com; do
  printf '%s ' "$u"; curl -s -o /dev/null -m 10 -w '%{http_code}\n' "$u" || echo "недоступен"
done
```

**Готово, если:**
- `uname -m` выводит `x86_64`, иначе образы с компьютера не подойдут;
- `getent` выводит IP сервера, иначе A-запись домена ещё не настроена или не разошлась;
- Let's Encrypt и GitHub отвечают. `401` у реестров Docker значит, что они доступны;
- порты 80 и 443 никем не заняты.

Инструкция рассчитана на Ubuntu 22.04 или 24.04. На другой системе команды этапов 2 и 3 нужно скорректировать.

## Этап 2. Подготовка системы

### 2.1. Обновления и swap

Если `apt-get update` падает или идёт очень медленно, переключитесь на зеркало и повторите:

```bash
grep -rlE 'archive\.ubuntu\.com' /etc/apt | xargs -r sed -i -E 's#https?://([a-z]+\.)?archive\.ubuntu\.com/ubuntu#http://mirror.yandex.ru/ubuntu#'
```

Основные команды:

```bash
export DEBIAN_FRONTEND=noninteractive
apt-get update && apt-get -y upgrade
apt-get install -y curl git ufw

fallocate -l 2G /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab
echo 'vm.swappiness=10' > /etc/sysctl.d/90-tpu.conf && sysctl -p /etc/sysctl.d/90-tpu.conf
```

**Готово, если** `free -h` показывает Swap 2.0Gi.

### 2.2. Защита входа

Только если `ssh tpu` работает по ключу. Текущее подключение не закрывайте, пока не проверите вход из нового окна.

```bash
cat > /etc/ssh/sshd_config.d/00-tpu.conf <<'EOF'
PasswordAuthentication no
PermitRootLogin prohibit-password
EOF
sshd -t && systemctl reload ssh
sshd -T | grep -E '^(passwordauthentication|permitrootlogin)'
```

**Готово, если** выведено `passwordauthentication no` и `ssh tpu "echo ok"` из нового окна отвечает `ok`.

### 2.3. Файрвол

Если в этапе 1 sshd слушает не порт 22, подставьте свой порт, иначе доступ пропадёт. В панели хостера порты 22, 80 и 443 тоже должны быть открыты.

```bash
ufw allow 22/tcp && ufw allow 80/tcp && ufw allow 443/tcp && ufw --force enable && ufw status
```

**Готово, если** `ssh tpu "echo ok"` из нового окна всё ещё отвечает.

## Этап 3. Docker

```bash
apt-get install -y docker.io docker-compose-v2
systemctl enable --now docker
docker --version; docker compose version
docker pull caddy:2.10-alpine
```

Если `docker pull` падает, подключите зеркало и повторите:

```bash
echo '{"registry-mirrors": ["https://mirror.gcr.io"]}' > /etc/docker/daemon.json && systemctl restart docker
```

Если зеркало тоже не помогло, загрузите базовые образы с компьютера (PowerShell) и на этапе 6 они уже будут на сервере:

```powershell
docker pull caddy:2.10-alpine; docker pull postgres:16; docker pull migrate/migrate:v4.18.3
docker save -o "$env:TEMP\tpu-base.tar" caddy:2.10-alpine postgres:16 migrate/migrate:v4.18.3
scp "$env:TEMP\tpu-base.tar" tpu:/root/
ssh tpu "docker load -i /root/tpu-base.tar && rm /root/tpu-base.tar"
```

**Готово, если** `docker compose version` выводит v2 и образ `caddy` скачан или загружен.

## Этап 4. Код и настройки

Репозиторий публичный, авторизация не нужна.

```bash
mkdir -p /opt/tpu && cd /opt/tpu
git clone -b backend  --depth 1 https://github.com/Psychoz42/TPUAttendanceControl.git TPUAttendanceControlBackend
git clone -b frontend --depth 1 https://github.com/Psychoz42/TPUAttendanceControl.git TPUAttendanceControlFrontend
cd TPUAttendanceControlBackend/deploy

cp .env.example .env
sed -i "s|^DOMAIN=.*|DOMAIN=<домен>|; s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=$(openssl rand -hex 32)|; s|^JWT_SECRET=.*|JWT_SECRET=$(openssl rand -hex 32)|" .env
chmod 600 .env
grep -c change-me .env; grep ^DOMAIN= .env
```

Создайте `users.txt`. Только демо-аккаунты со своими паролями: пароли лежат открытым текстом, настоящие пароли людей использовать нельзя. Формат строки: `логин пароль почта Фамилия-Имя-Отчество группа роль`. Этот файл заполняет владелец, пароли в чат не присылать.

```bash
nano users.txt
chmod 644 users.txt     # контейнер читает файл не от root
awk 'NF<5' users.txt | wc -l; awk '{print $6}' users.txt | sort | uniq -c
```

Пример строк:

```
ivanov    <пароль> ivanov@tpu.ru    Иванов-Иван-Иванович   8К52Т   student
petrova   <пароль> petrova@tpu.ru   Петрова-Анна-Сергеевна -       teacher
```

**Готово, если** `grep -c change-me` выводит `0`, домен верный, `awk 'NF<5'` выводит `0`, и в подсчёте есть и `student`, и `teacher`.

## Этап 5. Сборка и загрузка образов (на компьютере, PowerShell)

Запускать из папки бэкенда, рядом должна быть папка фронтенда с актуальной веткой:

```powershell
git pull
git -C ..\TPUAttendanceControlFrontend pull
docker build -t tpu-attendance-backend .
docker build -t tpu-attendance-frontend --build-arg NEXT_PUBLIC_API_URL=/api/v1 ..\TPUAttendanceControlFrontend
docker save -o "$env:TEMP\tpu-images.tar" tpu-attendance-backend tpu-attendance-frontend
scp "$env:TEMP\tpu-images.tar" tpu:/root/
ssh tpu "docker load -i /root/tpu-images.tar && rm /root/tpu-images.tar"
Remove-Item "$env:TEMP\tpu-images.tar"
```

Первая сборка фронтенда может занять минут 15: долго идёт `npm ci`.

**Готово, если** `ssh tpu "docker images"` показывает `tpu-attendance-backend` и `tpu-attendance-frontend`.

## Этап 6. Запуск

Перед запуском A-запись должна указывать на сервер, а порты 80 и 443 должны быть открыты. Иначе Caddy не получит сертификат, а повторные неудачные попытки могут упереться в лимит Let's Encrypt.

```bash
cd /opt/tpu/TPUAttendanceControlBackend/deploy
docker compose up -d --no-build
sleep 20
docker compose ps -a
docker compose logs migrate --tail 3
docker compose logs backend --tail 5
docker compose logs caddy --tail 20 | grep -iE 'certificate obtained|error' 
```

**Готово, если** `migrate` завершён с кодом 0, остальные контейнеры `running`, в логах Caddy есть `certificate obtained`.

## Этап 7. Проверка

С компьютера или сервера:

```bash
curl -s https://<домен>/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"x","password":"x"}'
# {"success":false,"code":"LOGIN_ERROR",...}

curl -s -o /dev/null -w '%{http_code}\n' https://<домен>/api/v1/attendance/getOwnerEvents -X POST -H 'Content-Type: application/json' -d '{}'
# 401: AUTH_REQUIRED включён

curl -s -o /dev/null -m 5 -w '%{http_code}\n' http://<IP>:5432/ ; curl -s -o /dev/null -m 5 -w '%{http_code}\n' http://<IP>:8080/
# 000 и 000: база и бэкенд снаружи недоступны

free -h; docker stats --no-stream
```

В браузере на `https://<домен>/`:
1. Вход преподавателем, создание занятия, на экране QR и таймер.
2. F5 на экране занятия: QR снова рабочий.
3. «Завершить приём»: занятие появляется в архиве, ведомость выгружается.

**Готово, если** всё сошлось, а `free -h` показывает запас памяти.

## Этап 8. Мобилка

1. Передать сокоманднику домен. Он собирает release: `-Ptpu.releaseBaseUrl=https://<домен>`. Нужен его файл `keystore.properties`.
2. Выдать ему тестового студента из `users.txt`. Сквозные тесты мобилки на этом сервере не запускать: они используют аккаунты с паролями из плана.
3. На телефоне: вход, скан QR с экрана, студент появляется в списке у преподавателя. Отдельно проверить QR с проектора с задних рядов.

## Обновление

На компьютере повторить этап 5. Если менялись `deploy/` или `migrations/`, на сервере сначала обновить код:

```bash
cd /opt/tpu/TPUAttendanceControlBackend/deploy
git pull
docker compose up -d --no-build
```

После правки `users.txt`: `docker compose restart backend`. После правки `Caddyfile`: `docker compose restart caddy`.

## Важно

- Смена `POSTGRES_PASSWORD` в `.env` не меняет пароль в уже созданной базе, его нужно менять через `ALTER USER`.
- Смена `JWT_SECRET` разлогинивает всех.
- `EVENT_RETENTION_DAYS=0` хранит занятия бессрочно. Положительное значение удаляет занятия и списки безвозвратно.
- `docker compose down -v` удаляет базу вместе со всеми данными. Обычная остановка: `docker compose down`, без `-v`.
- Логи ограничены пятью файлами по 50 МБ на контейнер.
- Бэкапы базы пока не настроены. Если данные станут важны: `docker compose exec -T postgres pg_dump -U postgres -Fc postgres > backup.dump`, и копию хранить не на этом сервере.
- Лимит 100 логинов в минуту с одного IP оставлен намеренно на время демо. Перед реальным использованием его нужно поднять, см. `routes/limiters.go`: группа за одним NAT упрётся в него.

Нагрузочный тест описан в [loadtest/README.md](loadtest/README.md).
