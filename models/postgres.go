package models

import (
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgresConfig) String() string {
	// fmt.Sprintf is used to format a string and return it without
	// printing it. It works like fmt.Printf, but instead of
	// displaying the output, it returns the formatted string.
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

// Open will open a SQL connection with the provided
// Postgres database. Callers of Open need to ensure
// that the connection is eventually closed via the
// db.Close() method.
func Open(config PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", config.String())
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	return db, nil
}

// Just for development
func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "baloo",
		Password: "junglebook",
		Database: "lenslocked",
		SSLMode:  "disable",
	}
}

func Migrate(db *sql.DB, dir string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	err = goose.Up(db, dir)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}

func MigrateFS(db *sql.DB, migrationFS fs.FS, dir string) error {
	if dir == "" {
		dir = "."
	}

	goose.SetBaseFS(migrationFS)
	// - Imagine you’re using an embedded filesystem (migrationFS) to read
	// migration files from somewhere other than the disk (e.g., from a .zip
	// or .tar file, or from resources bundled with your app).
	// - You set goose.SetBaseFS(migrationFS) to make goose use that custom
	// source for migrations.
	// - Once the migrations are done, you reset the base filesystem to nil
	// to avoid future code using the custom filesystem when it's not needed.
	defer func() {
		goose.SetBaseFS(nil)
	}()

	return Migrate(db, dir)
}
