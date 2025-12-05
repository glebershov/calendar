#!/bin/bash

# Скрипт для тестирования Calendar API
# Использование: ./test_api.sh

BASE_URL="http://localhost:8080"
OWNER_ID="test-user-123"

echo "🧪 Тестирование Calendar API"
echo "================================"
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для проверки ответа
check_response() {
    local status=$1
    local expected=$2
    local name=$3
    
    if [ "$status" -eq "$expected" ]; then
        echo -e "${GREEN}✓${NC} $name"
        return 0
    else
        echo -e "${RED}✗${NC} $name (ожидался статус $expected, получен $status)"
        return 1
    fi
}

# 1. Проверка health endpoint
echo "1. Проверка health endpoint..."
HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
check_response "$HEALTH_STATUS" 200 "Health check"
echo ""

# 2. Создание события
echo "2. Создание события..."
CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/events" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Тестовое событие",
    "description": "Описание тестового события",
    "start_time": "2024-12-20T10:00:00Z",
    "end_time": "2024-12-20T12:00:00Z",
    "owner_id": "'"$OWNER_ID"'"
  }')

CREATE_HTTP_CODE=$(echo "$CREATE_RESPONSE" | tail -n1)
CREATE_BODY=$(echo "$CREATE_RESPONSE" | sed '$d')

if check_response "$CREATE_HTTP_CODE" 201 "Создание события"; then
    EVENT_ID=$(echo "$CREATE_BODY" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "   Создано событие с ID: $EVENT_ID"
else
    echo "   Ответ: $CREATE_BODY"
    echo -e "${RED}Ошибка: не удалось создать событие. Проверьте, что сервер запущен.${NC}"
    exit 1
fi
echo ""

# 3. Получение списка событий
echo "3. Получение списка событий..."
LIST_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/events?owner_id=$OWNER_ID")
LIST_HTTP_CODE=$(echo "$LIST_RESPONSE" | tail -n1)
LIST_BODY=$(echo "$LIST_RESPONSE" | sed '$d')

if check_response "$LIST_HTTP_CODE" 200 "Получение списка событий"; then
    EVENT_COUNT=$(echo "$LIST_BODY" | grep -o '"id"' | wc -l | tr -d ' ')
    echo "   Найдено событий: $EVENT_COUNT"
else
    echo "   Ответ: $LIST_BODY"
fi
echo ""

# 4. Обновление события
if [ -n "$EVENT_ID" ]; then
    echo "4. Обновление события (ID: $EVENT_ID)..."
    UPDATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH "$BASE_URL/api/events/$EVENT_ID" \
      -H "Content-Type: application/json" \
      -d '{
        "title": "Обновленное событие",
        "description": "Обновленное описание"
      }')
    
    UPDATE_HTTP_CODE=$(echo "$UPDATE_RESPONSE" | tail -n1)
    check_response "$UPDATE_HTTP_CODE" 204 "Обновление события"
    echo ""
fi

# 5. Удаление события
if [ -n "$EVENT_ID" ]; then
    echo "5. Удаление события (ID: $EVENT_ID)..."
    DELETE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/api/events/$EVENT_ID")
    check_response "$DELETE_STATUS" 204 "Удаление события"
    echo ""
fi

# 6. Проверка, что событие удалено
if [ -n "$EVENT_ID" ]; then
    echo "6. Проверка удаления события..."
    LIST_AFTER_DELETE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/events?owner_id=$OWNER_ID")
    LIST_AFTER_DELETE_CODE=$(echo "$LIST_AFTER_DELETE" | tail -n1)
    LIST_AFTER_DELETE_BODY=$(echo "$LIST_AFTER_DELETE" | sed '$d')
    
    if [ "$LIST_AFTER_DELETE_CODE" -eq 200 ]; then
        EVENT_COUNT_AFTER=$(echo "$LIST_AFTER_DELETE_BODY" | grep -o '"id"' | wc -l | tr -d ' ')
        if [ "$EVENT_COUNT_AFTER" -eq 0 ]; then
            echo -e "${GREEN}✓${NC} Событие успешно удалено"
        else
            echo -e "${YELLOW}⚠${NC} Событие все еще присутствует в списке"
        fi
    fi
    echo ""
fi

# 7. Тест валидации (создание с невалидными данными)
echo "7. Тест валидации (создание без обязательных полей)..."
VALIDATION_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/events" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Событие без title и owner_id"
  }')

VALIDATION_HTTP_CODE=$(echo "$VALIDATION_RESPONSE" | tail -n1)
check_response "$VALIDATION_HTTP_CODE" 400 "Валидация (должна вернуть 400)"
echo ""

echo "================================"
echo -e "${GREEN}Тестирование завершено!${NC}"

