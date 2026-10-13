package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/redhat-et/cartograph/internal/db"
	"github.com/redhat-et/cartograph/internal/export"
	cartomcp "github.com/redhat-et/cartograph/internal/mcp"
	"github.com/redhat-et/cartograph/internal/migrate"
	"github.com/redhat-et/cartograph/internal/seed"
	"github.com/redhat-et/cartograph/schema"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: cartograph <command> [args]\n")
		fmt.Fprintf(os.Stderr, "Commands: serve, migrate, seed, export\n")
		os.Exit(1)
	}

	ctx := context.Background()

	switch os.Args[1] {
	case "migrate":
		cmdMigrate(ctx)
	case "seed":
		cmdSeed(ctx)
	case "export":
		cmdExport(ctx)
	case "serve":
		cmdServe(ctx)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func cmdMigrate(ctx context.Context) {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	dbURL := fs.String("db-url", os.Getenv("CARTOGRAPH_DB_URL"), "database connection URL")
	reset := fs.Bool("reset", false, "drop everything and re-apply all migrations")
	fs.Parse(os.Args[2:])

	if *dbURL == "" {
		slog.Error("--db-url or CARTOGRAPH_DB_URL required")
		os.Exit(1)
	}

	pool, err := db.Connect(ctx, *dbURL)
	if err != nil {
		slog.Error("connecting to database", "error", err)
		os.Exit(1)
	}
	defer db.Close(pool)

	if *reset {
		if err := migrate.Reset(ctx, pool, schema.Migrations, "migrations"); err != nil {
			slog.Error("reset failed", "error", err)
			os.Exit(1)
		}
		slog.Info("database reset complete")
		return
	}

	applied, err := migrate.Run(ctx, pool, schema.Migrations, "migrations")
	if err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	if len(applied) == 0 {
		slog.Info("database is up to date")
	} else {
		for _, m := range applied {
			slog.Info("applied migration", "version", m.Version, "name", m.Name)
		}
	}
}

func cmdSeed(ctx context.Context) {
	fs := flag.NewFlagSet("seed", flag.ExitOnError)
	dbURL := fs.String("db-url", os.Getenv("CARTOGRAPH_DB_URL"), "database connection URL")
	from := fs.String("from", "", "directory containing nodes.json and edges.json (default: built-in example)")
	fs.Parse(os.Args[2:])

	if *dbURL == "" {
		slog.Error("--db-url or CARTOGRAPH_DB_URL required")
		os.Exit(1)
	}

	pool, err := db.Connect(ctx, *dbURL)
	if err != nil {
		slog.Error("connecting to database", "error", err)
		os.Exit(1)
	}
	defer db.Close(pool)

	var result seed.Result
	if *from != "" {
		result, err = seed.Import(ctx, pool,
			filepath.Join(*from, "nodes.json"),
			filepath.Join(*from, "edges.json"),
		)
	} else {
		result, err = seed.ImportFS(ctx, pool, schema.Seed,
			"seed/nodes.json",
			"seed/edges.json",
		)
	}
	if err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
	slog.Info("seed complete", "nodes", result.NodesCreated, "edges", result.EdgesCreated)
}

func cmdExport(ctx context.Context) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	dbURL := fs.String("db-url", os.Getenv("CARTOGRAPH_DB_URL"), "database connection URL")
	to := fs.String("to", ".", "output directory for nodes.json and edges.json")
	fs.Parse(os.Args[2:])

	if *dbURL == "" {
		slog.Error("--db-url or CARTOGRAPH_DB_URL required")
		os.Exit(1)
	}

	pool, err := db.Connect(ctx, *dbURL)
	if err != nil {
		slog.Error("connecting to database", "error", err)
		os.Exit(1)
	}
	defer db.Close(pool)

	result, err := export.Export(ctx, pool,
		filepath.Join(*to, "nodes.json"),
		filepath.Join(*to, "edges.json"),
	)
	if err != nil {
		slog.Error("export failed", "error", err)
		os.Exit(1)
	}
	slog.Info("export complete", "nodes", result.NodesExported, "edges", result.EdgesExported)
}

func cmdServe(ctx context.Context) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	dbURL := fs.String("db-url", os.Getenv("CARTOGRAPH_DB_URL"), "database connection URL")
	stdio := fs.Bool("stdio", false, "serve MCP over stdio")
	httpAddr := fs.String("http", "", "serve MCP over HTTP (e.g. :8080)")
	fs.Parse(os.Args[2:])

	if *dbURL == "" {
		slog.Error("--db-url or CARTOGRAPH_DB_URL required")
		os.Exit(1)
	}

	if !*stdio && *httpAddr == "" {
		slog.Error("specify --stdio and/or --http <addr>")
		os.Exit(1)
	}

	pool, err := db.Connect(ctx, *dbURL)
	if err != nil {
		slog.Error("connecting to database", "error", err)
		os.Exit(1)
	}
	defer db.Close(pool)

	srv := cartomcp.NewServer(pool)

	if *stdio {
		slog.Info("serving MCP over stdio")
		if err := srv.Run(ctx, &gomcp.StdioTransport{}); err != nil {
			slog.Error("stdio server error", "error", err)
			os.Exit(1)
		}
		return
	}

	if *httpAddr != "" {
		mcpHandler := gomcp.NewStreamableHTTPHandler(
			func(_ *http.Request) *gomcp.Server { return srv },
			nil,
		)
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			if err := pool.Ping(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, "db: %v", err)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
		mux.Handle("/", mcpHandler)
		slog.Info("serving MCP over HTTP", "addr", *httpAddr)
		if err := http.ListenAndServe(*httpAddr, mux); err != nil {
			slog.Error("http server error", "error", err)
			os.Exit(1)
		}
	}
}
