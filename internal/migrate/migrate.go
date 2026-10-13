package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AppliedMigration struct {
	Version int
	Name    string
}

func Run(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, dir string) ([]AppliedMigration, error) {
	if err := ensureMigrationsTable(ctx, pool); err != nil {
		return nil, err
	}

	applied, err := getAppliedVersions(ctx, pool)
	if err != nil {
		return nil, err
	}

	migrations, err := loadMigrationFiles(fsys, dir)
	if err != nil {
		return nil, err
	}

	var result []AppliedMigration
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}

		slog.Info("applying migration", "version", m.version, "name", m.name)
		if err := applyMigration(ctx, pool, m); err != nil {
			return result, fmt.Errorf("migration %03d_%s: %w", m.version, m.name, err)
		}
		result = append(result, AppliedMigration{Version: m.version, Name: m.name})
	}

	return result, nil
}

func Reset(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, dir string) error {
	slog.Info("resetting database — dropping all objects")

	_, _ = pool.Exec(ctx, "SELECT drop_graph('ecosystem', true)")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS schema_migrations CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS node_types CASCADE")
	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS edge_types CASCADE")

	_, err := Run(ctx, pool, fsys, dir)
	return err
}

type migrationFile struct {
	version int
	name    string
	sql     string
}

func loadMigrationFiles(fsys fs.FS, dir string) ([]migrationFile, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("reading migrations dir: %w", err)
	}

	var migrations []migrationFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}

		parts := strings.SplitN(strings.TrimSuffix(e.Name(), ".sql"), "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid migration filename: %s (expected NNN_name.sql)", e.Name())
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid version in %s: %w", e.Name(), err)
		}

		content, err := fs.ReadFile(fsys, dir+"/"+e.Name())
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", e.Name(), err)
		}

		migrations = append(migrations, migrationFile{
			version: version,
			name:    parts[1],
			sql:     string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			name       TEXT NOT NULL,
			applied_at TIMESTAMPTZ DEFAULT now()
		)`)
	return err
}

func getAppliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[int]bool, error) {
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, m migrationFile) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, m.sql); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
		m.version, m.name); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
