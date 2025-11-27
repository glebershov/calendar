APP_NAME := calendar
BIN_DIR := bin
CMD_DIR := ./cmd/calendar

.PHONY: build run test docker-up docker-down fmt

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

