package mcp

import (
	"github.com/jackc/pgx/v5/pgxpool"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewServer(pool *pgxpool.Pool) *gomcp.Server {
	srv := gomcp.NewServer(&gomcp.Implementation{
		Name:    "cartograph",
		Version: "1.0.0",
	}, nil)

	registerNodeTools(srv, pool)
	registerEdgeTools(srv, pool)
	registerQueryTools(srv, pool)
	registerResources(srv, pool)

	return srv
}
