# Cartograph

Map companies, people, organizations, partnerships, technologies, and market signals as a graph. Query and modify the graph through [MCP](https://modelcontextprotocol.io/) tools from any AI agent.

Go 1.25 / PostgreSQL 16 / [Apache AGE](https://age.apache.org/) 1.6.0 (OpenCypher) / [Crunchy PGO](https://www.crunchydata.com/products/crunchy-postgresql-for-kubernetes) v5

## Prerequisites

- OpenShift cluster with [Crunchy PGO](https://www.crunchydata.com/products/crunchy-postgresql-for-kubernetes) v5 operator installed
- `oc` CLI installed and logged in (`oc login`)
- [Go 1.25+](https://go.dev/dl/) (to install the CLI)

## Deploying

Deploy to OpenShift with Crunchy PGO. Images are on Quay.io.

### 1. Create project and secrets

```bash
oc new-project cartograph

oc create secret generic pgbackrest-s3 \
  --from-literal=aws-s3-key=YOUR_ACCESS_KEY \
  --from-literal=aws-s3-key-secret=YOUR_SECRET_KEY

oc create secret generic oauth2-proxy \
  --from-literal=client-id=YOUR_CLIENT_ID \
  --from-literal=client-secret=YOUR_CLIENT_SECRET \
  --from-literal=cookie-secret=$(openssl rand -base64 32)
```

### 2. Deploy with Kustomize

```bash
oc apply -k 'https://github.com/redhat-et/cartograph//deploy/prod?ref=main'
```

Includes the PostgresCluster CR, Cartograph server, OAuth2 proxy, and route.

### 3. Initialize the database

```bash
go install github.com/redhat-et/cartograph/cmd/cartograph@latest

export CARTOGRAPH_DB_URL=$(oc get secret cartograph-pguser-cartograph \
  -o jsonpath='{.data.uri}' | base64 -d)

cartograph migrate --db-url "$CARTOGRAPH_DB_URL"
cartograph seed --db-url "$CARTOGRAPH_DB_URL"
```

Seeds built-in example data. Pass `--from <dir>` to load your own `nodes.json` and `edges.json` instead.

## Using

Add Cartograph to your MCP client:

```bash
# Claude Code
claude mcp add cartograph --transport http https://cartograph.apps.your-cluster.example.com/

# OpenCode
opencode mcp add cartograph --url https://cartograph.apps.your-cluster.example.com/
```

### Tools

| Tool | Description |
| ------ | ------------- |
| `create_node` | Create a node with label and properties |
| `get_node` | Get a node by ID |
| `update_node` | Merge-update node properties |
| `delete_node` | Delete a node and all its edges |
| `list_nodes` | List nodes, optionally filtered by label |
| `search_nodes` | Search nodes by name pattern within a label |
| `create_edge` | Create an edge between two nodes |
| `get_edge` | Get an edge by ID |
| `delete_edge` | Delete an edge |
| `list_edges` | List edges, optionally filtered by type |
| `get_neighbors` | Traverse neighbors by edge type and direction |
| `find_path` | Find shortest path between two nodes |

### Resources

| URI | Description |
|-----|-------------|
| `cartograph://types/nodes` | Node labels and their descriptions |
| `cartograph://types/edges` | Edge types with valid source/target labels |

## Developing

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Podman](https://podman.io/) (or Docker)

### Local database

```bash
scripts/local-db.sh start    # starts PostgreSQL+AGE via podman
scripts/local-db.sh url      # prints CARTOGRAPH_DB_URL
scripts/local-db.sh stop
```

### Build and test

```bash
go build -o cartograph ./cmd/cartograph

# tests require a running local database
export CARTOGRAPH_DB_URL=$(scripts/local-db.sh url)
cartograph migrate --db-url "$CARTOGRAPH_DB_URL"
go test -p 1 ./...           # -p 1: packages share one database
```

### Run locally

```bash
# stdio mode (for MCP clients that launch the server)
cartograph serve --stdio --db-url "$CARTOGRAPH_DB_URL"

# HTTP mode
cartograph serve --http :8080 --db-url "$CARTOGRAPH_DB_URL"
# health check: curl http://localhost:8080/healthz
```

Add the local server to Claude Code:

```bash
claude mcp add cartograph -- cartograph serve --stdio
```

### Container images

```bash
# Database image (PostgreSQL 16 + AGE 1.6.0)
image/build.sh

# Server image
podman build -t quay.io/redhat-et/cartograph:latest .
```

### Connecting to a deployed database

```bash
scripts/connect.sh             # port-forwards and opens a psql session
scripts/connect.sh --namespace my-ns   # non-default namespace
```

### Other commands

```bash
cartograph seed                       # import built-in example data
cartograph seed --from ./my-data      # import custom data (nodes.json + edges.json)
cartograph export --to ./out          # export graph to JSON
```
