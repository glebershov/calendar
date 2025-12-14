# Инструкция по запуску и проверке приложения

## Способ 1: Запуск через Docker Compose (Рекомендуется)

### Шаги:

1. **Запустите все сервисы:**

   ```bash
   make docker-up
   # или
   docker compose up -d --build
   ```

2. **Проверьте, что все контейнеры запущены:**

   ```bash
   docker compose ps
   ```

   Должны быть запущены:

   - `calendar-postgres` (PostgreSQL)
   - `zookeeper` (Zookeeper для Kafka)
   - `kafka` (Kafka брокер)
   - `calendar-app` (Ваше приложение)

3. **Проверьте логи приложения:**

   ```bash
   docker compose logs -f calendar
   ```

   Вы должны увидеть:

   - `connected to postgres`
   - `kafka producer started`
   - `kafka consumer started`
   - `starting http server`

4. **Проверьте healthcheck:**

   ```bash
   curl http://localhost:8080/health
   ```

   Должен вернуть: `OK`

5. **Создайте тестовое событие через API:**

   ```bash
   curl -X POST http://localhost:8080/api/events \
     -H "Content-Type: application/json" \
     -d '{
       "id": "test-event-1",
       "title": "Тестовое событие",
       "description": "Описание события",
       "start_time": "2024-12-25T10:00:00Z",
       "end_time": "2024-12-25T11:00:00Z",
       "owner_id": "user-1"
     }'
   ```

6. **Проверьте логи Kafka Consumer:**

   ```bash
   docker compose logs -f calendar | grep "received event from kafka"
   ```

   Через 10 секунд после создания события producer должен отправить его в Kafka, а consumer должен получить и залогировать.

7. **Проверьте список событий:**
   ```bash
   curl http://localhost:8080/api/events?owner_id=user-1
   ```

### Остановка:

```bash
make docker-down
# или
docker compose down
```

