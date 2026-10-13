package seed_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redhat-et/cartograph/internal/db"
	"github.com/redhat-et/cartograph/internal/migrate"
	"github.com/redhat-et/cartograph/internal/seed"
	"github.com/redhat-et/cartograph/schema"
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

func TestImport(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	if err := migrate.Reset(ctx, pool, schema.Migrations, "migrations"); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	testdataDir, _ := filepath.Abs("../../testdata")
	result, err := seed.Import(ctx, pool,
		filepath.Join(testdataDir, "nodes.json"),
		filepath.Join(testdataDir, "edges.json"),
	)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if result.NodesCreated != 3 {
		t.Fatalf("expected 3 nodes, got %d", result.NodesCreated)
	}
	if result.EdgesCreated != 2 {
		t.Fatalf("expected 2 edges, got %d", result.EdgesCreated)
	}
}

func TestImportRejectsInvalidLabel(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	tmpDir := t.TempDir()
	nodesFile := filepath.Join(tmpDir, "nodes.json")
	os.WriteFile(nodesFile, []byte(`[{"label": "FakeLabel", "properties": {"name": "bad"}}]`), 0644)
	edgesFile := filepath.Join(tmpDir, "edges.json")
	os.WriteFile(edgesFile, []byte(`[]`), 0644)

	_, err := seed.Import(ctx, pool, nodesFile, edgesFile)
	if err == nil {
		t.Fatal("expected error for invalid label")
	}
}
