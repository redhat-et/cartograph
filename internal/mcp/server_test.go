package mcp_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redhat-et/cartograph/internal/db"
	cartomcp "github.com/redhat-et/cartograph/internal/mcp"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	connStr := os.Getenv("CARTOGRAPH_DB_URL")
	if connStr == "" {
		t.Skip("CARTOGRAPH_DB_URL not set")
	}
	pool, err := db.Connect(context.Background(), connStr)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { db.Close(pool) })
	return pool
}

func TestNewServer(t *testing.T) {
	pool := testPool(t)
	srv := cartomcp.NewServer(pool)
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
}
