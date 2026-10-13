package graph_test

import (
	"context"
	"testing"

	"github.com/redhat-et/cartograph/internal/graph"
)

func TestCreateAndGetNode(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	node, err := graph.CreateNode(ctx, pool, "Company", map[string]any{
		"name":       "TestCorp",
		"hq_country": "US",
	})
	if err != nil {
		t.Fatalf("CreateNode: %v", err)
	}
	if node.ID == 0 {
		t.Fatal("expected non-zero node ID")
	}
	if node.Label != "Company" {
		t.Fatalf("expected label Company, got %s", node.Label)
	}

	got, err := graph.GetNode(ctx, pool, node.ID)
	if err != nil {
		t.Fatalf("GetNode: %v", err)
	}
	if got.Properties["name"] != "TestCorp" {
		t.Fatalf("expected name TestCorp, got %v", got.Properties["name"])
	}
}

func TestCreateNodeRejectsInvalidLabel(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	_, err := graph.CreateNode(ctx, pool, "InvalidLabel", map[string]any{"name": "test"})
	if err == nil {
		t.Fatal("expected error for invalid label")
	}
}

func TestUpdateNode(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	node, err := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Before"})
	if err != nil {
		t.Fatalf("CreateNode: %v", err)
	}

	updated, err := graph.UpdateNode(ctx, pool, node.ID, map[string]any{"name": "After", "headcount": 50})
	if err != nil {
		t.Fatalf("UpdateNode: %v", err)
	}
	if updated.Properties["name"] != "After" {
		t.Fatalf("expected name After, got %v", updated.Properties["name"])
	}
}

func TestDeleteNode(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	node, err := graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "ToDelete"})
	if err != nil {
		t.Fatalf("CreateNode: %v", err)
	}

	if err := graph.DeleteNode(ctx, pool, node.ID); err != nil {
		t.Fatalf("DeleteNode: %v", err)
	}

	_, err = graph.GetNode(ctx, pool, node.ID)
	if err == nil {
		t.Fatal("expected error getting deleted node")
	}
}

func TestListNodes(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	_, err := graph.CreateNode(ctx, pool, "Technology", map[string]any{"name": "TestTech"})
	if err != nil {
		t.Fatalf("CreateNode: %v", err)
	}

	nodes, err := graph.ListNodes(ctx, pool, "Technology")
	if err != nil {
		t.Fatalf("ListNodes: %v", err)
	}
	if len(nodes) == 0 {
		t.Fatal("expected at least one Technology node")
	}
}

func TestSearchNodes(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	_, _ = graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Acme Robotics"})
	_, _ = graph.CreateNode(ctx, pool, "Company", map[string]any{"name": "Acme AI"})

	results, err := graph.SearchNodes(ctx, pool, "Company", "Acme%")
	if err != nil {
		t.Fatalf("SearchNodes: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
}
