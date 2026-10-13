package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: cartograph <command> [args]\n")
		fmt.Fprintf(os.Stderr, "Commands: serve, migrate, seed, export\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		fmt.Println("serve: not yet implemented")
	case "migrate":
		fmt.Println("migrate: not yet implemented")
	case "seed":
		fmt.Println("seed: not yet implemented")
	case "export":
		fmt.Println("export: not yet implemented")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
