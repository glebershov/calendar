# Инструкция по тестированию Calendar API

## Быстрый старт

### 1. Запуск приложения через Docker Compose (рекомендуется)

```bash
make docker-up
```

Это запустит:
- PostgreSQL на порту 5432
- Calendar API на порту 8080

### 2. Запуск приложения локально

Если PostgreSQL уже запущен локально:

```bash
# Убедитесь, что PostgreSQL запущен и доступен
# Затем запустите приложение:
make run
```

### 3. Автоматическое тестирование

Запустите скрипт тестирования:

```bash
./test_api.sh
```

Скрипт проверит:
- ✅ Health check endpoint
- ✅ Создание события (POST /api/events)
- ✅ Получение списка событий (GET /api/events)
- ✅ Обновление события (PATCH /api/events/{id})
- ✅ Удаление события (DELETE /api/events/{id})
- ✅ Валидацию данных

## Ручное тестирование с curl

### 1. Health check
```bash
curl http://localhost:8080/health
```

Ожидается: `OK` со статусом 200

### 2. Создание события
```bash
curl -X POST http://localhost:8080/api/events \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Встреча",
    "description": "Важная встреча",
    "start_time": "2024-12-20T10:00:00Z",
    "end_time": "2024-12-20T12:00:00Z",
    "owner_id": "user-123"
  }'
```

Ожидается: JSON с созданным событием (статус 201)

### 3. Получение списка событий
```bash
curl "http://localhost:8080/api/events?owner_id=user-123"
```

Ожидается: JSON массив со списком событий (статус 200)

### 4. Обновление события
```bash
# Замените {EVENT_ID} на реальный ID события
curl -X PATCH http://localhost:8080/api/events/{EVENT_ID} \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Обновленная встреча"
  }'
```

Ожидается: статус 204 (No Content)

### 5. Удаление события
```bash
# Замените {EVENT_ID} на реальный ID события
curl -X DELETE http://localhost:8080/api/events/{EVENT_ID}
```

Ожидается: статус 204 (No Content)

## Проверка логов

Если приложение запущено через Docker:
```bash
docker logs calendar-app
```

## Остановка приложения

```bash
make docker-down
```

## Запуск unit-тестов (если есть)

```bash
make test
```

