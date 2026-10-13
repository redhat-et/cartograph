# Cartograph — Agent Instructions

Ecosystem intelligence graph. Go 1.25 + PostgreSQL 16 + Apache AGE 1.6.0 (OpenCypher).
MCP server exposing graph CRUD to AI agents.

## Commands

```bash
# test (packages share one DB — no parallel package execution)
CARTOGRAPH_DB_URL=postgres://postgres:devpass@localhost:5432/cartograph \
  go test -p 1 ./...

# build
go build -o cartograph ./cmd/cartograph

# lint
goimports -l .    # must produce no output
go vet ./...

# local dev database
scripts/local-db.sh start   # podman; prints CARTOGRAPH_DB_URL
scripts/local-db.sh stop

# container images
podman build -t cartograph .                          # server
image/build.sh                                        # database (AGE)
```

## Sandbox environment

If running in a sandboxed environment without network or module cache access:

```bash
export GOMODCACHE=/tmp/gomodcache
export GONOSUMCHECK='*'
export GONOSUMDB='*'
export GOPATH=/tmp/gopath
export GOFLAGS=''
```

## Package layout

```text
cmd/cartograph/     CLI: serve, migrate, seed, export subcommands
internal/
  db/               connection pool (pgx/v5)
  graph/            node/edge CRUD, traversal — all Cypher via AGE
  mcp/              MCP server, tool + resource registration
  migrate/          schema migrations with advisory locking
  seed/             seed import (transactional)
  export/           graph export to JSON
image/              Containerfile + build script for Crunchy PG16 + AGE
deploy/             Kustomize manifests (base + prod overlay + minio)
scripts/            local-db.sh, connect.sh, backup-restore.sh
```

## Conventions

### Commits

Conventional Commits, imperative mood, no period, no scope:
`feat: add neighbor traversal depth limit`

### Go style

- `goimports` enforced (superset of `gofmt`)
- `any` not `interface{}`
- Wrap errors: `fmt.Errorf("context: %w", err)`
- No naked returns, no `init()` functions

### Design

- YAGNI: no speculative interfaces, feature flags, or backwards-compat shims
- Three similar lines beat a premature abstraction
- No comments unless the WHY is non-obvious

### Testing

- Table-driven tests
- Name: `TestGetNode_returns_error_for_missing_id`
- Real database, never mocks
- `-p 1` required: packages share a single database

### Cypher / injection safety

Never interpolate user input into Cypher strings. Use the helpers in
`internal/graph/nodes.go`: `escapeCypherString()` for values,
`toCypherMap()` for properties, `ValidateEdgeType()` for edge labels.
All queries use `$cypher$` dollar-quoting.

### Graph layer

- All graph functions accept `Querier`, not `*pgxpool.Pool`
- Wrap multi-step writes in a transaction: `pool.Begin()`, pass `tx` as `Querier`

### AGE wire format

AGE returns `agtype` values: IDs as bare integers, labels as JSON strings,
properties as JSON objects.

### MCP tools

Tool input structs use `jsonschema` tags with plain description text:

```go
Type string `json:"type" jsonschema:"required,node label (Company, Person)"`
```

Never use `description=...` prefix in the tag — it causes a panic.

### Dependencies

Keep minimal. Run `go mod tidy` before committing.
