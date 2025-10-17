package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Open открывает БД.
// Миграция при открытии.
func Open() (*sqlx.DB, error) {
	dsn := os.Getenv("TODO_DBFILE")
	if dsn == "" {
		dsn = "scheduler.db"
	}

	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err = db.DB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	// Автомиграция
	if err = Migrate(db); err != nil {
		return nil, fmt.Errorf("migrate schema: %w", err)
	}

	return db, nil
}

// Migrate создаёт таблицу задачек.
func Migrate(db *sqlx.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,
  date    TEXT    NOT NULL,
  title   TEXT    NOT NULL,
  comment TEXT    NOT NULL DEFAULT '',
  repeat  TEXT    NOT NULL DEFAULT ''
);`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
