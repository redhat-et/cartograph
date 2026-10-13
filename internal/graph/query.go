package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

)

var agtypeAnnotation = regexp.MustCompile(`::(?:vertex|edge|path)`)

// stripAgtypeAnnotations removes AGE's ::vertex, ::edge, ::path suffixes from agtype output.
func stripAgtypeAnnotations(s string) string {
	return agtypeAnnotation.ReplaceAllString(s, "")
}

func GetNeighbors(ctx context.Context, q Querier, nodeID int64, edgeType, direction string) ([]Node, error) {
	if err := ValidateEdgeType(ctx, q, edgeType); err != nil {
		return nil, err
	}

	var pattern string
	switch direction {
	case "outgoing":
		pattern = fmt.Sprintf("(a)-[e:%s]->(b)", edgeType)
	case "incoming":
		pattern = fmt.Sprintf("(b)-[e:%s]->(a)", edgeType)
	case "both":
		pattern = fmt.Sprintf("(a)-[e:%s]-(b)", edgeType)
	default:
		return nil, fmt.Errorf("invalid direction: %q (use outgoing, incoming, or both)", direction)
	}

	query := fmt.Sprintf(
		`SELECT * FROM cypher('ecosystem', $cypher$
			MATCH %s
			WHERE id(a) = %d
			RETURN id(b), label(b), properties(b)
		$cypher$) AS (id agtype, label agtype, props agtype)`,
		pattern, nodeID,
	)

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var idRaw, labelRaw, propsRaw string
		if err := rows.Scan(&idRaw, &labelRaw, &propsRaw); err != nil {
			return nil, err
		}
		var n Node
		json.Unmarshal([]byte(idRaw), &n.ID)
		json.Unmarshal([]byte(labelRaw), &n.Label)
		json.Unmarshal([]byte(propsRaw), &n.Properties)
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func FindPath(ctx context.Context, q Querier, fromID, toID int64, maxHops int) ([]Node, []Edge, error) {
	if maxHops <= 0 || maxHops > 10 {
		maxHops = 5
	}

	// AGE may not support shortestPath — use variable-length path match
	query := fmt.Sprintf(
		`SELECT * FROM cypher('ecosystem', $cypher$
			MATCH p = (a)-[*1..%d]-(b)
			WHERE id(a) = %d AND id(b) = %d
			RETURN nodes(p), relationships(p)
			LIMIT 1
		$cypher$) AS (nodes agtype, rels agtype)`,
		maxHops, fromID, toID,
	)

	var nodesRaw, relsRaw string
	if err := q.QueryRow(ctx, query).Scan(&nodesRaw, &relsRaw); err != nil {
		return nil, nil, fmt.Errorf("no path found between %d and %d within %d hops: %w", fromID, toID, maxHops, err)
	}

	// AGE returns nodes/relationships with ::vertex/::edge annotations — strip them
	cleanNodes := stripAgtypeAnnotations(nodesRaw)
	cleanEdges := stripAgtypeAnnotations(relsRaw)

	var rawNodes []json.RawMessage
	var rawEdges []json.RawMessage
	if err := json.Unmarshal([]byte(cleanNodes), &rawNodes); err != nil {
		return nil, nil, fmt.Errorf("parsing path nodes: %w", err)
	}
	if err := json.Unmarshal([]byte(cleanEdges), &rawEdges); err != nil {
		return nil, nil, fmt.Errorf("parsing path edges: %w", err)
	}

	nodes := make([]Node, len(rawNodes))
	for i, raw := range rawNodes {
		var entry struct {
			ID         int64          `json:"id"`
			Label      string         `json:"label"`
			Properties map[string]any `json:"properties"`
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			return nil, nil, fmt.Errorf("parsing path node %d: %w", i, err)
		}
		nodes[i] = Node{ID: entry.ID, Label: entry.Label, Properties: entry.Properties}
	}

	edges := make([]Edge, len(rawEdges))
	for i, raw := range rawEdges {
		var entry struct {
			ID         int64          `json:"id"`
			Label      string         `json:"label"`
			StartID    int64          `json:"start_id"`
			EndID      int64          `json:"end_id"`
			Properties map[string]any `json:"properties"`
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			return nil, nil, fmt.Errorf("parsing path edge %d: %w", i, err)
		}
		edges[i] = Edge{ID: entry.ID, Type: entry.Label, SourceID: entry.StartID, TargetID: entry.EndID, Properties: entry.Properties}
	}

	return nodes, edges, nil
}
