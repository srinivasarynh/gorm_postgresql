package migrations

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

var migrationsFS embed.FS

func RunMigrations(dsn string) error {
	d, err := iofs.New(migrationsFS, "sql")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", &err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", &err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
