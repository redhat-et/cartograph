package schema

import "embed"

//go:embed migrations
var Migrations embed.FS

//go:embed seed
var Seed embed.FS
