package database

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

func Initialize() (*sql.DB, error) {
	if err := os.MkdirAll(".yagi", 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", "file:.yagi/yagi.db")
	if err != nil {
		slog.Debug("failed to open database", "error", err)
		return nil, err
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		slog.Debug("failed to create driver", "error", err)
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"yagi",
		driver,
	)

	if err != nil {
		slog.Debug("failed to create migrate instance", "error", err)
		return nil, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Debug("failed to run migrations", "error", err)
		return nil, err
	}

	return db, nil
}
