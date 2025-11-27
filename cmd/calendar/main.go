package main

import (
    "log"

    "calendar/internal/application"
    "calendar/internal/config"
)

func main() {
    // 1. Загружаем конфиг
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // 2. Инициализируем приложение (логгер, БД, HTTP‑сервер)
    app, err := application.NewApp(cfg)
    if err != nil {
        log.Fatalf("failed to init application: %v", err)
    }

    // 3. Запускаем приложение с graceful shutdown
    if err := app.Run(); err != nil {
        log.Fatalf("application stopped with error: %v", err)
    }
}
