package server

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const (
	dbSchema        = "workflows"
	migrationsTable = "schema_migrations_manager"
)

func upMigrations(connStr, migrationsDir string) error {
	slog.Info("up migrations...")

	connStringURL, err := url.Parse(connStr)
	if err != nil {
		return fmt.Errorf("parsing connection string: %w", err)
	}

	values := connStringURL.Query()
	values.Set("search_path", dbSchema) // set db schema
	//values.Set("x-migrations-table", migrationsTable) // set migrations table
	connStringURL.RawQuery = values.Encode()
	connStr = connStringURL.String()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("open postgres connection: %w", err)
	}

	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS " + dbSchema)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	pgMigrate, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrations: %w", err)
	}
	if err := pgMigrate.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("up migrations: no changes")

			return nil
		}

		return fmt.Errorf("up: %w", err)
	}

	slog.Info("up migrations: done")

	return nil
}
