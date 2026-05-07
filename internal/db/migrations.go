package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS _migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("create _migrations table: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, file := range files {
		var applied bool
		pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM _migrations WHERE filename=$1)`, file).Scan(&applied)
		if applied {
			continue
		}

		path := filepath.Join(dir, file)
		sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("run migration %s: %w", file, err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO _migrations (filename) VALUES ($1)`, file); err != nil {
			return fmt.Errorf("record migration %s: %w", file, err)
		}
	}
	return nil
}
