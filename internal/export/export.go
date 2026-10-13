package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redhat-et/cartograph/internal/graph"
)

type Result struct {
	NodesExported int
	EdgesExported int
}

type exportNode struct {
	Label      string         `json:"label"`
	Properties map[string]any `json:"properties"`
}

type exportEdge struct {
	Type       string         `json:"type"`
	Source     string         `json:"source"`
	Target     string         `json:"target"`
	Properties map[string]any `json:"properties"`
}

func Export(ctx context.Context, pool *pgxpool.Pool, nodesFile, edgesFile string) (Result, error) {
	nodes, err := graph.ListNodes(ctx, pool, "")
	if err != nil {
		return Result{}, fmt.Errorf("listing nodes: %w", err)
	}

	idToName := make(map[int64]string)
	var exportNodes []exportNode
	for _, n := range nodes {
		name, _ := n.Properties["name"].(string)
		idToName[n.ID] = name
		exportNodes = append(exportNodes, exportNode{
			Label:      n.Label,
			Properties: n.Properties,
		})
	}

	edges, err := graph.ListEdges(ctx, pool, "")
	if err != nil {
		return Result{}, fmt.Errorf("listing edges: %w", err)
	}

	var exportEdges []exportEdge
	for _, e := range edges {
		exportEdges = append(exportEdges, exportEdge{
			Type:       e.Type,
			Source:     idToName[e.SourceID],
			Target:     idToName[e.TargetID],
			Properties: e.Properties,
		})
	}

	if err := writeJSON(nodesFile, exportNodes); err != nil {
		return Result{}, fmt.Errorf("writing nodes: %w", err)
	}
	if err := writeJSON(edgesFile, exportEdges); err != nil {
		return Result{}, fmt.Errorf("writing edges: %w", err)
	}

	return Result{NodesExported: len(exportNodes), EdgesExported: len(exportEdges)}, nil
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
