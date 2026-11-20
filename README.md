# Сокращатель ссылок

## Запуск через Docker Compose

Требуются Docker и плагин Docker Compose.

```powershell
cp .env.example .env
docker compose up --build
```

API доступно по адресу
`http://localhost:8084`, Swagger UI — по адресу `http://localhost:8081`.


```powershell
Copy-Item .env.example .env
```

Все маршруты защищены HTTP Basic Auth. Для локального запуска используются
значения из `.env`; замените их перед публикацией сервиса в сети.

Миграции применяются при старте приложения, если `AUTO_MIGRATE=true`.
Данные PostgreSQL сохраняются в Docker volume `postgres_data`.

## API

- `POST /url/` — создать короткую ссылку
- `PATCH /url/{alias}` — изменить alias
- `DELETE /url/{alias}` — удалить короткую ссылку
- `GET /{alias}` — перенаправить на исходный URL

## Тесты

```powershell
go test ./...
```

Тесты Basic Auth берут `AUTH_USERNAME` и `AUTH_PASSWORD` из корневого `.env`
или из окружения процесса.
