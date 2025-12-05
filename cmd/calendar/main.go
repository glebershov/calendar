package main

import (
	"log"
	// "time"

	"calendar/internal/application"
	"calendar/internal/config"
	"calendar/internal/migrations"
)

func main() {
	// 1. Загружаем конфиг
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Авто‑миграции (только если включены в конфиге)
	if cfg.Postgres.AutoMigrate {
		if err := migrations.Up(cfg.Postgres.DSN); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
	}

	// 3. Инициализируем приложение (логгер, БД, HTTP‑сервер, хендлеры)
	app, err := application.NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to init application: %v", err)
	}

	//  Фоновое задание с периодическим запуском
    // if cfg.BackgroundTask.Enabled {
    //     ticker := time.NewTicker(cfg.BackgroundTask.Period)
    //     stop := make(chan bool)

    //     go func() {
    //         for {
    //             select {
    //             case <-ticker.C:
    //                 // бизнес-логика фоновой задачи
    //                 log.Println("Выполняется фоновая задача...")
    //             case <-stop:
    //                 ticker.Stop()
    //                 return
    //             }
    //         }
    //     }()

    //     defer func() {
    //         stop <- true
    //     }()

	// 4. Запускаем приложение с graceful shutdown
	if err := app.Run(); err != nil {
		log.Fatalf("application stopped with error: %v", err)
	}
}
