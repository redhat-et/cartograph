package migrate_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redhat-et/cartograph/internal/db"
	"github.com/redhat-et/cartograph/internal/migrate"
	"github.com/redhat-et/cartograph/schema"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	connStr := os.Getenv("CARTOGRAPH_DB_URL")
	if connStr == "" {
		t.Skip("CARTOGRAPH_DB_URL not set")
	}
	ctx := context.Background()
	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { db.Close(pool) })
	return pool
}

func TestRunAppliesMigrations(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	err := migrate.Reset(ctx, pool, schema.Migrations, "migrations")
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}

	var count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count < 1 {
		t.Fatal("expected at least 1 migration applied")
	}
}

func TestRunIsIdempotent(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	err := migrate.Reset(ctx, pool, schema.Migrations, "migrations")
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}

	applied, err := migrate.Run(ctx, pool, schema.Migrations, "migrations")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(applied) != 0 {
		t.Fatalf("expected 0 newly applied migrations, got %d", len(applied))
	}
}

func TestNodeTypesPopulated(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	err := migrate.Reset(ctx, pool, schema.Migrations, "migrations")
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}

	var count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM node_types").Scan(&count)
	if err != nil {
		t.Fatalf("query node_types: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected 4 node types, got %d", count)
	}
}

func TestEdgeTypesPopulated(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	err := migrate.Reset(ctx, pool, schema.Migrations, "migrations")
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}

	var count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM edge_types").Scan(&count)
	if err != nil {
		t.Fatalf("query edge_types: %v", err)
	}
	if count != 13 {
		t.Fatalf("expected 13 edge types, got %d", count)
	}
}
