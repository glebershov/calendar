APP_NAME := calendar
BIN_DIR := bin
CMD_DIR := ./cmd/calendar
DB_DSN := postgres://calendar_user:calendar_pass@localhost:5432/calendar_db?sslmode=disable
MIGRATIONS_DIR := ./migrations

.PHONY: build run test docker-up docker-down fmt migrate-up migrate-down migrate-force

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)

run:
	go run $(CMD_DIR)

test:
	go test ./...

fmt:
	gofmt -w cmd internal

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" -verbose up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" -verbose down


