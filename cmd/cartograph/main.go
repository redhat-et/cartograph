package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/redhat-et/cartograph/internal/db"
	"github.com/redhat-et/cartograph/internal/migrate"
	"github.com/redhat-et/cartograph/schema"
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
	case "serve":
		fmt.Println("serve: not yet implemented")
	case "seed":
		fmt.Println("seed: not yet implemented")
	case "export":
		fmt.Println("export: not yet implemented")
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
