package mcp

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redhat-et/cartograph/internal/graph"
)

type getNeighborsInput struct {
	NodeID    int64  `json:"node_id" jsonschema:"node ID to find neighbors of"`
	EdgeType  string `json:"edge_type" jsonschema:"edge type to traverse"`
	Direction string `json:"direction" jsonschema:"traversal direction: outgoing, incoming, or both"`
}

type findPathInput struct {
	FromID  int64 `json:"from_id" jsonschema:"start node ID"`
	ToID    int64 `json:"to_id" jsonschema:"end node ID"`
	MaxHops int   `json:"max_hops,omitempty" jsonschema:"maximum path length (default 5, max 10)"`
}

func registerQueryTools(srv *gomcp.Server, pool *pgxpool.Pool) {
	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "get_neighbors",
		Description: "Get neighboring nodes connected by a specific edge type",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args getNeighborsInput) (*gomcp.CallToolResult, any, error) {
		nodes, err := graph.GetNeighbors(ctx, pool, args.NodeID, args.EdgeType, args.Direction)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Found %d neighbors:\n%s", len(nodes), mustJSON(nodes))), nil, nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "find_path",
		Description: "Find the shortest path between two nodes",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args findPathInput) (*gomcp.CallToolResult, any, error) {
		maxHops := args.MaxHops
		if maxHops == 0 {
			maxHops = 5
		}
		nodes, edges, err := graph.FindPath(ctx, pool, args.FromID, args.ToID, maxHops)
		if err != nil {
			return errResult(err), nil, nil
		}
		result := map[string]any{"nodes": nodes, "edges": edges}
		return jsonResult(result), nil, nil
	})
}
