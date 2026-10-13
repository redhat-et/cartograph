package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/redhat-et/cartograph/internal/db"
)

func TestConnect(t *testing.T) {
	connStr := os.Getenv("CARTOGRAPH_DB_URL")
	if connStr == "" {
		t.Skip("CARTOGRAPH_DB_URL not set — skipping integration test")
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer db.Close(pool)

	var result int
	err = pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("ping query failed: %v", err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %d", result)
	}
}

func TestConnectAGE(t *testing.T) {
	connStr := os.Getenv("CARTOGRAPH_DB_URL")
	if connStr == "" {
		t.Skip("CARTOGRAPH_DB_URL not set — skipping integration test")
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer db.Close(pool)

	var version string
	err = pool.QueryRow(ctx, "SELECT extversion FROM pg_extension WHERE extname = 'age'").Scan(&version)
	if err != nil {
		t.Fatalf("AGE extension query failed: %v", err)
	}
	if version == "" {
		t.Fatal("AGE extension version is empty")
	}
	t.Logf("AGE version: %s", version)
}
