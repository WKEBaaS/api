package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"i3s-service/internal/configs"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed postgres/*.sql
var postgresMigrations embed.FS

func MigrateI3S(config *configs.Config) error {
	// db, err := pgx.Connect(context.Background(), config.DatabaseURL)
	db, err := sql.Open("pgx", config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("failed to close database: %v\n", err)
		}
	}()

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}

	source, err := iofs.New(postgresMigrations, "postgres")
	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	mig, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("migrations applied successfully")
	return nil
}
