package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	_ "modernc.org/sqlite"
)

func OpenSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}
	if _, err = db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return db, nil
}

func RunMigrations(ctx context.Context, db *sql.DB, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".sql" {
			files = append(files, filepath.Join(migrationsDir, e.Name()))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("migration %s failed: %w", filepath.Base(f), err)
		}
	}

	return nil
}
