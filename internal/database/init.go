package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

const (
	dataFolder                  = "data"
	dbFileName                  = "yagi.sqlite3"
	migrationsFolder            = "migrations"
	ownerReadWriteOthersReadDir = 0755
)

func Initialize() (*Queries, error) {
	db, err := initialize(dataFolder, dbFileName, migrationsFolder)
	if err != nil {
		return nil, err
	}

	queries := New(db)
	return queries, nil
}

func initialize(dataFolder string, dbFileName string, migrationsFolder string) (*sql.DB, error) {
	if _, err := os.Stat(dataFolder); os.IsNotExist(err) {
		slog.Debug("data folder does not exist, creating it", "dataFolder", dataFolder)
		if err := os.Mkdir(dataFolder, ownerReadWriteOthersReadDir); err != nil {
			return nil, err
		}
	}

	slog.Debug("initializing database", "dataFolder", dataFolder, "dbFileName", dbFileName)
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s/%s", dataFolder, dbFileName))
	if err != nil {
		slog.Debug("failed to open database", "error", err)
		return nil, err
	}

	// Database driver for migrations
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		slog.Debug("failed to create driver", "error", err)
		return nil, err
	}

	// Source for migrations -> embedded filesystem
	ms, err := iofs.New(migrations, migrationsFolder)
	if err != nil {
		slog.Debug("failed to create migrations source", "error", err)
		return nil, err
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		ms,
		"yagi",
		driver,
	)

	if err != nil {
		slog.Debug("failed to create migrate instance", "error", err)
		return nil, err
	}

	slog.Debug("running up migrations", "migrationsFolder", migrationsFolder)
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Debug("failed to run up migrations", "error", err)
		return nil, err
	}

	return db, nil
}
