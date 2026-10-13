package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redhat-et/cartograph/internal/graph"
)

type createNodeInput struct {
	Label      string         `json:"label" jsonschema:"node label (Company, Org, Person, Technology)"`
	Properties map[string]any `json:"properties" jsonschema:"node properties (must include name)"`
}

type nodeIDInput struct {
	ID int64 `json:"id" jsonschema:"node ID"`
}

type updateNodeInput struct {
	ID         int64          `json:"id" jsonschema:"node ID"`
	Properties map[string]any `json:"properties" jsonschema:"properties to set or update"`
}

type listNodesInput struct {
	Label string `json:"label,omitempty" jsonschema:"filter by label (optional)"`
}

type searchNodesInput struct {
	Label   string `json:"label" jsonschema:"node label to search within"`
	Pattern string `json:"pattern" jsonschema:"name pattern (use %% as wildcard)"`
}

func registerNodeTools(srv *gomcp.Server, pool *pgxpool.Pool) {
	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "create_node",
		Description: "Create a new node in the ecosystem graph",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args createNodeInput) (*gomcp.CallToolResult, any, error) {
		node, err := graph.CreateNode(ctx, pool, args.Label, args.Properties)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Created %s node (id=%d): %s", node.Label, node.ID, node.Properties["name"])), nil, nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "get_node",
		Description: "Get a node by its ID",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args nodeIDInput) (*gomcp.CallToolResult, any, error) {
		node, err := graph.GetNode(ctx, pool, args.ID)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(node), nil, nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "update_node",
		Description: "Update a node's properties (merges with existing)",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args updateNodeInput) (*gomcp.CallToolResult, any, error) {
		node, err := graph.UpdateNode(ctx, pool, args.ID, args.Properties)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(node), nil, nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "delete_node",
		Description: "Delete a node and all its edges",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args nodeIDInput) (*gomcp.CallToolResult, any, error) {
		if err := graph.DeleteNode(ctx, pool, args.ID); err != nil {
			return errResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Deleted node %d", args.ID)), nil, nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "list_nodes",
		Description: "List nodes, optionally filtered by label",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args listNodesInput) (*gomcp.CallToolResult, any, error) {
		nodes, err := graph.ListNodes(ctx, pool, args.Label)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(nodes), nil, nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "search_nodes",
		Description: "Search nodes by name pattern within a label",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args searchNodesInput) (*gomcp.CallToolResult, any, error) {
		nodes, err := graph.SearchNodes(ctx, pool, args.Label, args.Pattern)
		if err != nil {
			return errResult(err), nil, nil
		}
		return jsonResult(nodes), nil, nil
	})
}

func textResult(text string) *gomcp.CallToolResult {
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: text}},
	}
}

func errResult(err error) *gomcp.CallToolResult {
	r := &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: err.Error()}},
	}
	r.IsError = true
	return r
}

func jsonResult(v any) *gomcp.CallToolResult {
	data, _ := json.MarshalIndent(v, "", "  ")
	return textResult(string(data))
}
