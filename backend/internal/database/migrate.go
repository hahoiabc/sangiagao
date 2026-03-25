package database

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations executes all pending SQL migrations from the embedded filesystem.
// Migrations are tracked in a schema_migrations table.
// Files must be named like: 002_description.sql (number prefix determines order).
func RunMigrations(pool *pgxpool.Pool, fs embed.FS, dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create tracking table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// Get already applied versions
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("scan version: %w", err)
		}
		applied[v] = true
	}

	// Read migration files
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type migration struct {
		version int
		name    string
		path    string
	}
	var migrations []migration

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		// Parse version from filename: "002_price_board.sql" -> 2
		var ver int
		if _, err := fmt.Sscanf(e.Name(), "%d_", &ver); err != nil {
			slog.Warn("Skipping migration file (bad name)", "file", e.Name())
			continue
		}
		migrations = append(migrations, migration{
			version: ver,
			name:    e.Name(),
			path:    dir + "/" + e.Name(),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	// Apply pending migrations
	count := 0
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}

		sql, err := fs.ReadFile(m.path)
		if err != nil {
			return fmt.Errorf("read %s: %w", m.name, err)
		}

		slog.Info("Applying migration", "version", m.version, "name", m.name)

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", m.name, err)
		}

		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute %s: %w", m.name, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version, name) VALUES ($1, $2)", m.version, m.name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", m.name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", m.name, err)
		}

		slog.Info("Migration applied", "version", m.version, "name", m.name)
		count++
	}

	if count == 0 {
		slog.Info("Database is up to date", "applied_total", len(applied))
	} else {
		slog.Info("Migrations complete", "newly_applied", count, "total", len(applied)+count)
	}

	return nil
}
