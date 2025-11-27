package databases

import (
    "database/sql"
    "fmt"

    _ "github.com/lib/pq"

    "calendar/internal/config"
)

type Postgres struct {
    DB *sql.DB
}

func NewPostgres(cfg *config.Config) (*Postgres, error) {
    db, err := sql.Open("postgres", cfg.Postgres.DSN)
    if err != nil {
        return nil, fmt.Errorf("failed to open postgres: %w", err)
    }

    if err := db.Ping(); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to ping postgres: %w", err)
    }

    // Настройки пула соединений (опционально)
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)

    return &Postgres{DB: db}, nil
}

func (p *Postgres) Close() error {
    if p.DB != nil {
        return p.DB.Close()
    }
    return nil
}

