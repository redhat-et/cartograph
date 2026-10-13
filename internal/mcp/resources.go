package mcp

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redhat-et/cartograph/internal/graph"
)

func registerResources(srv *gomcp.Server, pool *pgxpool.Pool) {
	srv.AddResource(&gomcp.Resource{
		URI:         "cartograph://types/nodes",
		Name:        "Node Types",
		Description: "Available node labels and their descriptions",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *gomcp.ReadResourceRequest) (*gomcp.ReadResourceResult, error) {
		types, err := graph.ListNodeTypes(ctx, pool)
		if err != nil {
			return nil, err
		}
		data, _ := json.MarshalIndent(types, "", "  ")
		return &gomcp.ReadResourceResult{
			Contents: []*gomcp.ResourceContents{{
				URI:      "cartograph://types/nodes",
				MIMEType: "application/json",
				Text:     string(data),
			}},
		}, nil
	})

	srv.AddResource(&gomcp.Resource{
		URI:         "cartograph://types/edges",
		Name:        "Edge Types",
		Description: "Available edge types with valid source/target labels",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *gomcp.ReadResourceRequest) (*gomcp.ReadResourceResult, error) {
		types, err := graph.ListEdgeTypes(ctx, pool)
		if err != nil {
			return nil, err
		}
		data, _ := json.MarshalIndent(types, "", "  ")
		return &gomcp.ReadResourceResult{
			Contents: []*gomcp.ResourceContents{{
				URI:      "cartograph://types/edges",
				MIMEType: "application/json",
				Text:     string(data),
			}},
		}, nil
	})
}

func mustJSON(v any) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}
