# My Files

Небольшое веб-приложение для загрузки, хранения, скачивания и удаления файлов.

Основная часть проекта - Go-сервер. Клиентская часть собирается в статические файлы и отдается через nginx.

## Структура проекта

```text
.
├── client/              # React/Vite клиент, в Docker отдается через nginx
├── server/              # Go backend
├── docker-compose.yaml  # запуск всего приложения
└── README.md
```

## Сервер

Сервер написан на Go и слушает порт `8000` внутри Docker-сети.

Основные возможности сервера:

- авторизация по паролю;
- хранение сессии в cookie `session_id`;
- загрузка файлов через `POST /files`;
- список файлов через `GET /files`;
- скачивание файла через `GET /files/{id}`;
- удаление файла через `DELETE /files/{id}`;
- хранение метаданных в SQLite;
- хранение содержимого файлов кусками на диске.

В Docker backend не публикуется наружу через `ports`. Он доступен только другим контейнерам внутри Compose-сети. Наружу смотрит только nginx-контейнер клиента.

## Авторизация

Перед обычным запуском нужно один раз задать пароль:

```powershell
docker compose run --rm setup-auth
```

Эта команда запускает тот же backend-образ, но с флагом:

```text
./file-uploader -setup-auth
```

Пароль вводится интерактивно. В базе сохраняется не исходный пароль, а bcrypt-хеш.

После этого приложение запускается обычной командой:

```powershell
docker compose up --build
```

Если пароль нужно изменить, можно снова выполнить:

```powershell
docker compose run --rm setup-auth
```

Команда использует тот же Docker volume, что и основной сервер, поэтому обновляет пароль в той же SQLite базе.

## Хранение данных

Путь для данных сервера задается переменной окружения:

```text
APP_DATA_DIR
```

В Docker Compose она установлена так:

```yaml
APP_DATA_DIR: /var/lib/my-files
```

Этот путь внутри контейнера подключен к Docker volume:

```yaml
volumes:
  - file-data:/var/lib/my-files
```

То есть данные лежат не внутри временной файловой системы контейнера, а в постоянном Docker volume `my-files_file-data`.

Внутри volume структура такая:

```text
/var/lib/my-files/
├── files.db
└── chunks/
    └── <file-id>/
        ├── chunk_000000
        ├── chunk_000001
        └── ...
```

`files.db` - SQLite база с метаданными, паролем и сессиями.

`chunks/` - содержимое загруженных файлов. Каждый файл хранится в отдельной папке по UUID, разбитый на куски.

Данные сохраняются после:

```powershell
docker compose down
docker compose up --build
docker compose restart
```

Данные будут удалены, если явно удалить volume:

```powershell
docker compose down -v
docker volume rm my-files_file-data
docker system prune --volumes
```

Если нужно посмотреть файлы внутри volume:

```powershell
docker compose exec server ls -R /var/lib/my-files
```

Если сервер сейчас не запущен:

```powershell
docker compose run --rm server ls -R /var/lib/my-files
```

## Docker Compose

В Compose описаны три сервиса:

```text
server      основной Go backend
setup-auth  одноразовая команда для настройки пароля
client      nginx + собранный frontend
```

Обычный запуск:

```powershell
docker compose up --build
```

Запуск в фоне:

```powershell
docker compose up -d --build
```

Остановка:

```powershell
docker compose down
```

Важно: не добавляйте `-v`, если хотите сохранить базу и загруженные файлы.

## Порты

Наружу опубликован только frontend/nginx:

```yaml
client:
  ports:
    - "5173:80"
```

Открыть приложение:

```text
http://localhost:5173
```

Backend в Compose использует:

```yaml
server:
  expose:
    - "8000"
```

`expose` не открывает порт наружу. Он нужен только для связи контейнеров внутри Docker-сети.

nginx проксирует API-запросы:

```text
/login -> http://server:8000/login
/files -> http://server:8000/files
```

Поэтому браузер ходит на `localhost:5173`, а nginx уже передает API-запросы backend-контейнеру.

## Локальный запуск сервера без Docker

Из папки `server`:

```powershell
go run . -setup-auth
go run .
```

По умолчанию, если `APP_DATA_DIR` не задан, сервер использует локальную папку:

```text
server/data
```

Можно указать свой путь:

```powershell
$env:APP_DATA_DIR="C:\my-files-data"
go run .
```

## API

```text
POST   /login       авторизация, принимает поле password
GET    /files       список файлов
POST   /files       загрузка файла, multipart поле file
GET    /files/{id}  скачивание файла
DELETE /files/{id}  удаление файла
```

Все маршруты `/files` требуют cookie `session_id`.

## Безопасность

Что уже сделано:

- пароль хранится как bcrypt-хеш;
- SQL-запросы используют параметры, а не конкатенацию строк;
- backend не опубликован наружу отдельным портом;
- SQLite база не имеет сетевого порта;
- cookie сессии имеет `HttpOnly` и `SameSite=Lax`.

Что важно учитывать для реального сервера:

- для публичного доступа нужен HTTPS;
- при HTTPS cookie стоит выставлять с `Secure=true`;
- сейчас нет rate limit на `/login`;
- nginx ограничивает размер загрузки через `client_max_body_size`;
- Docker volume нужно бэкапить отдельно;
- не публикуйте backend или базу через `ports`, если в этом нет необходимости.
