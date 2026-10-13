package graph_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redhat-et/cartograph/internal/db"
	"github.com/redhat-et/cartograph/internal/graph"
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

func TestListNodeTypes(t *testing.T) {
	pool := testPool(t)
	types, err := graph.ListNodeTypes(context.Background(), pool)
	if err != nil {
		t.Fatalf("ListNodeTypes: %v", err)
	}
	if len(types) != 4 {
		t.Fatalf("expected 4 node types, got %d", len(types))
	}

	labels := make(map[string]bool)
	for _, nt := range types {
		labels[nt.Label] = true
	}
	for _, expected := range []string{"Company", "Org", "Person", "Technology"} {
		if !labels[expected] {
			t.Errorf("missing node type: %s", expected)
		}
	}
}

func TestListEdgeTypes(t *testing.T) {
	pool := testPool(t)
	types, err := graph.ListEdgeTypes(context.Background(), pool)
	if err != nil {
		t.Fatalf("ListEdgeTypes: %v", err)
	}
	if len(types) != 13 {
		t.Fatalf("expected 13 edge types, got %d", len(types))
	}
}

func TestValidateNodeLabel(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	if err := graph.ValidateNodeLabel(ctx, pool, "Company"); err != nil {
		t.Errorf("Company should be valid: %v", err)
	}
	if err := graph.ValidateNodeLabel(ctx, pool, "Nonexistent"); err == nil {
		t.Error("Nonexistent should be invalid")
	}
}

func TestValidateEdge(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	if err := graph.ValidateEdge(ctx, pool, "INVESTS_IN", "Company", "Company"); err != nil {
		t.Errorf("Company->INVESTS_IN->Company should be valid: %v", err)
	}
	if err := graph.ValidateEdge(ctx, pool, "INVESTS_IN", "Person", "Company"); err == nil {
		t.Error("Person->INVESTS_IN->Company should be invalid")
	}
	if err := graph.ValidateEdge(ctx, pool, "FAKE_TYPE", "Company", "Company"); err == nil {
		t.Error("FAKE_TYPE should be invalid")
	}
}
