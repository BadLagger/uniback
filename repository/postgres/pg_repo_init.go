package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"uniback/utils"
)

//go:embed migrations/*
var migarionsFS embed.FS

func createMigrationTable(ctx context.Context, db *sql.DB) error {

	_, err := db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS migrations (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) UNIQUE NOT NULL,
				applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`)

	return err
}

func readMigrationFiles() (map[string]string, error) {
	migrations := make(map[string]string)

	files, err := fs.ReadDir(migarionsFS, "migrations")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		content, err := fs.ReadFile(migarionsFS, "migrations/"+file.Name())
		if err != nil {
			return nil, err
		}

		migrations[file.Name()] = string(content)
	}

	return migrations, nil
}

func applyMigrations(ctx context.Context, db *sql.DB, migrations map[string]string) error {
	log := utils.GlobalLogger()

	keys := make([]string, 0, len(migrations))
	for name := range migrations {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {

		sql := migrations[name]

		var applied bool
		err := db.QueryRowContext(ctx,
			"SELECT EXISTS (SELECT 1 FROM migrations WHERE name = $1)", name).Scan(&applied)

		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if !applied {
			log.Debug("%s try to apply...", name)
			tx, err := db.BeginTx(ctx, nil)

			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			if _, err := tx.ExecContext(ctx, sql); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %w", name, err)
			}

			if _, err := tx.ExecContext(ctx,
				"INSERT INTO migrations (name) VALUES ($1)", name); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration %s: %w", name, err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration %s: %w", name, err)
			}

			log.Debug("%s applied OK!!!", name)
		} else {
			log.Debug("%s already applied...", name)
		}
	}

	return nil
}

func initSchema(ctx context.Context, db *sql.DB) error {

	log := utils.GlobalLogger()

	log.Info("Try to Init DB...")

	if err := createMigrationTable(ctx, db); err != nil {
		log.Error("Check migration table FAIL!")
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	log.Debug("Check migration table OK!")

	migrations, err := readMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	return applyMigrations(ctx, db, migrations)
}
