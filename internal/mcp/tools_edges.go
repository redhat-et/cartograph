package mcp

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redhat-et/cartograph/internal/graph"
)

type createEdgeInput struct {
	Type       string         `json:"type" jsonschema:"edge type (e.g. INVESTS_IN, SUPPLIES, PARTNERS_WITH)"`
	SourceID   int64          `json:"source_id" jsonschema:"source node ID"`
	TargetID   int64          `json:"target_id" jsonschema:"target node ID"`
	Properties map[string]any `json:"properties,omitempty" jsonschema:"edge properties (optional)"`
}

type edgeIDInput struct {
	ID int64 `json:"id" jsonschema:"edge ID"`
}

type listEdgesInput struct {
	Type string `json:"type,omitempty" jsonschema:"filter by edge type (optional)"`
}

func registerEdgeTools(srv *gomcp.Server, pool *pgxpool.Pool) {
	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "create_edge",
		Description: "Create an edge between two nodes",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args createEdgeInput) (*gomcp.CallToolResult, any, error) {
		edge, err := graph.CreateEdge(ctx, pool, args.Type, args.SourceID, args.TargetID, args.Properties)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Created %s edge (id=%d): %d -> %d", edge.Type, edge.ID, edge.SourceID, edge.TargetID)), nil, nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "get_edge",
		Description: "Get an edge by its ID",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args edgeIDInput) (*gomcp.CallToolResult, any, error) {
		edge, err := graph.GetEdge(ctx, pool, args.ID)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(edge), nil, nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "delete_edge",
		Description: "Delete an edge by its ID",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args edgeIDInput) (*gomcp.CallToolResult, any, error) {
		if err := graph.DeleteEdge(ctx, pool, args.ID); err != nil {
			return errResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Deleted edge %d", args.ID)), nil, nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "list_edges",
		Description: "List edges, optionally filtered by type",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args listEdgesInput) (*gomcp.CallToolResult, any, error) {
		edges, err := graph.ListEdges(ctx, pool, args.Type)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(edges), nil, nil
	})
}
