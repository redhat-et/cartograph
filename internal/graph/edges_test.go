package graph_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redhat-et/cartograph/internal/graph"
)

func createTestCompanies(t *testing.T, pool *pgxpool.Pool, ctx context.Context) (graph.Node, graph.Node) {
	t.Helper()
	a, err := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Company A"})
	if err != nil {
		t.Fatalf("CreateNode A: %v", err)
	}
	b, err := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Company B"})
	if err != nil {
		t.Fatalf("CreateNode B: %v", err)
	}
	return a, b
}

func TestCreateAndGetEdge(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	a, b := createTestCompanies(t, pool, ctx)

	edge, err := graph.CreateEdge(ctx, pool, "INVESTS_IN", a.ID, b.ID, map[string]any{
		"annotation": "Series B lead",
		"confidence": "confirmed",
	})
	if err != nil {
		t.Fatalf("CreateEdge: %v", err)
	}
	if edge.ID == 0 {
		t.Fatal("expected non-zero edge ID")
	}

	got, err := graph.GetEdge(ctx, pool, edge.ID)
	if err != nil {
		t.Fatalf("GetEdge: %v", err)
	}
	if got.Type != "INVESTS_IN" {
		t.Fatalf("expected INVESTS_IN, got %s", got.Type)
	}
	if got.SourceID != a.ID || got.TargetID != b.ID {
		t.Fatalf("wrong source/target: %d->%d, expected %d->%d", got.SourceID, got.TargetID, a.ID, b.ID)
	}
}

func TestCreateEdgeRejectsInvalidType(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	a, b := createTestCompanies(t, pool, ctx)

	_, err := graph.CreateEdge(ctx, pool, "FAKE_TYPE", a.ID, b.ID, nil)
	if err == nil {
		t.Fatal("expected error for invalid edge type")
	}
}

func TestCreateEdgeRejectsInvalidSourceTarget(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	company, _ := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Corp"})
	person, _ := graph.CreateNode(ctx, pool, "Person", map[string]any{"name": "Alice"})

	_, err := graph.CreateEdge(ctx, pool, "INVESTS_IN", person.ID, company.ID, nil)
	if err == nil {
		t.Fatal("expected error: Person cannot be source of INVESTS_IN")
	}
}

func TestDeleteEdge(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	a, b := createTestCompanies(t, pool, ctx)

	edge, _ := graph.CreateEdge(ctx, pool, "PARTNERS_WITH", a.ID, b.ID, nil)
	if err := graph.DeleteEdge(ctx, pool, edge.ID); err != nil {
		t.Fatalf("DeleteEdge: %v", err)
	}

	_, err := graph.GetEdge(ctx, pool, edge.ID)
	if err == nil {
		t.Fatal("expected error getting deleted edge")
	}
}
