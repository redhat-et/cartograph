package graph_test

import (
	"context"
	"testing"

	"github.com/redhat-et/cartograph/internal/graph"
)

func TestGetNeighbors(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	a, _ := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Hub Co"})
	b, _ := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Spoke 1"})
	c, _ := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Spoke 2"})
	graph.CreateEdge(ctx, pool, "INVESTS_IN", a.ID, b.ID, nil)
	graph.CreateEdge(ctx, pool, "INVESTS_IN", a.ID, c.ID, nil)

	outgoing, err := graph.GetNeighbors(ctx, pool, a.ID, "INVESTS_IN", "outgoing")
	if err != nil {
		t.Fatalf("GetNeighbors: %v", err)
	}
	if len(outgoing) != 2 {
		t.Fatalf("expected 2 outgoing neighbors, got %d", len(outgoing))
	}

	incoming, err := graph.GetNeighbors(ctx, pool, b.ID, "INVESTS_IN", "incoming")
	if err != nil {
		t.Fatalf("GetNeighbors incoming: %v", err)
	}
	if len(incoming) != 1 {
		t.Fatalf("expected 1 incoming neighbor, got %d", len(incoming))
	}
}

func TestFindPath(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	a, _ := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Start"})
	b, _ := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Middle"})
	c, _ := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "End"})
	graph.CreateEdge(ctx, pool, "SUPPLIES", a.ID, b.ID, nil)
	graph.CreateEdge(ctx, pool, "SUPPLIES", b.ID, c.ID, nil)

	nodes, edges, err := graph.FindPath(ctx, pool, a.ID, c.ID, 5)
	if err != nil {
		t.Fatalf("FindPath: %v", err)
	}
	if len(nodes) < 3 {
		t.Fatalf("expected at least 3 nodes in path, got %d", len(nodes))
	}
	if len(edges) < 2 {
		t.Fatalf("expected at least 2 edges in path, got %d", len(edges))
	}
}
