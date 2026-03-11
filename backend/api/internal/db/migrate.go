package db

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"streamforge/api/migrations"
)

// RunMigrations executes database migrations using the provided database URL
func RunMigrations(databaseURL string) error {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs wrapper for migrations: %w", err)
	}

	// Prepare pgx connection string format for golang-migrate
	// pgx/v5 driver expects "pgx5://" scheme instead of "postgres://"
	// This is a simple replacement but handle with care
	pgxURL := "pgx5" + databaseURL[8:] // Replace 'postgres' with 'pgx5'

	m, err := migrate.NewWithSourceInstance("iofs", d, pgxURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate instance: %w", err)
	}

	// Run migrations
	err = m.Up()

	if err != nil && errors.Is(err, migrate.ErrNoChange) {
		log.Println("Database migrations: no changes to apply")
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Logging complete
	version, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	log.Printf("Successfully applied migrations. Current version: %d, Dirty: %v\n", version, dirty)

	return nil
}
