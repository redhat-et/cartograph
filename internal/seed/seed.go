package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redhat-et/cartograph/internal/graph"
)

type Result struct {
	NodesCreated int
	EdgesCreated int
}

type seedNode struct {
	Label      string         `json:"label"`
	Properties map[string]any `json:"properties"`
}

type seedEdge struct {
	Type       string         `json:"type"`
	Source     string         `json:"source"`
	Target     string         `json:"target"`
	Properties map[string]any `json:"properties"`
}

func Import(ctx context.Context, pool *pgxpool.Pool, nodesFile, edgesFile string) (Result, error) {
	return ImportFS(ctx, pool, os.DirFS("."), nodesFile, edgesFile)
}

func ImportFS(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, nodesFile, edgesFile string) (Result, error) {
	nodes, err := loadNodes(fsys, nodesFile)
	if err != nil {
		return Result{}, fmt.Errorf("loading nodes: %w", err)
	}

	edges, err := loadEdges(fsys, edgesFile)
	if err != nil {
		return Result{}, fmt.Errorf("loading edges: %w", err)
	}

	for i, n := range nodes {
		if err := graph.ValidateNodeLabel(ctx, pool, n.Label); err != nil {
			return Result{}, fmt.Errorf("node %d (%s): %w", i, n.Properties["name"], err)
		}
	}

	nodeLabels := make(map[string]string)
	for _, n := range nodes {
		name, _ := n.Properties["name"].(string)
		nodeLabels[name] = n.Label
	}

	for i, e := range edges {
		srcLabel, ok := nodeLabels[e.Source]
		if !ok {
			return Result{}, fmt.Errorf("edge %d: unknown source node %q", i, e.Source)
		}
		tgtLabel, ok := nodeLabels[e.Target]
		if !ok {
			return Result{}, fmt.Errorf("edge %d: unknown target node %q", i, e.Target)
		}
		if err := graph.ValidateEdge(ctx, pool, e.Type, srcLabel, tgtLabel); err != nil {
			return Result{}, fmt.Errorf("edge %d (%s->%s->%s): %w", i, e.Source, e.Type, e.Target, err)
		}
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	nameToID := make(map[string]int64)
	for _, n := range nodes {
		created, err := graph.CreateNode(ctx, tx, n.Label, n.Properties)
		if err != nil {
			return Result{}, fmt.Errorf("creating node %v: %w", n.Properties["name"], err)
		}
		name, _ := n.Properties["name"].(string)
		nameToID[name] = created.ID
		slog.Debug("created node", "label", n.Label, "name", name, "id", created.ID)
	}

	edgesCreated := 0
	for _, e := range edges {
		srcID := nameToID[e.Source]
		tgtID := nameToID[e.Target]
		_, err := graph.CreateEdge(ctx, tx, e.Type, srcID, tgtID, e.Properties)
		if err != nil {
			return Result{}, fmt.Errorf("creating edge %s->%s->%s: %w", e.Source, e.Type, e.Target, err)
		}
		edgesCreated++
	}

	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("committing transaction: %w", err)
	}

	return Result{NodesCreated: len(nodes), EdgesCreated: edgesCreated}, nil
}

func loadNodes(fsys fs.FS, path string) ([]seedNode, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}
	var nodes []seedNode
	return nodes, json.Unmarshal(data, &nodes)
}

func loadEdges(fsys fs.FS, path string) ([]seedEdge, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}
	var edges []seedEdge
	return edges, json.Unmarshal(data, &edges)
}
