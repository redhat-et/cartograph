package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type Node struct {
	ID         int64
	Label      string
	Properties map[string]any
}

func CreateNode(ctx context.Context, q Querier, label string, props map[string]any) (Node, error) {
	if err := ValidateNodeLabel(ctx, q, label); err != nil {
		return Node{}, err
	}

	cypherProps, err := toCypherMap(props)
	if err != nil {
		return Node{}, fmt.Errorf("invalid properties: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT * FROM cypher('ecosystem', $cypher$ CREATE (n:%s) SET n = %s RETURN id(n), properties(n) $cypher$) AS (id agtype, props agtype)`,
		label, cypherProps,
	)

	var idRaw, propsRaw string
	if err := q.QueryRow(ctx, query).Scan(&idRaw, &propsRaw); err != nil {
		return Node{}, fmt.Errorf("creating node: %w", err)
	}

	node := Node{Label: label}
	if err := json.Unmarshal([]byte(idRaw), &node.ID); err != nil {
		return Node{}, fmt.Errorf("parsing node id: %w", err)
	}
	if err := json.Unmarshal([]byte(propsRaw), &node.Properties); err != nil {
		return Node{}, fmt.Errorf("parsing properties: %w", err)
	}

	return node, nil
}

func GetNode(ctx context.Context, q Querier, id int64) (Node, error) {
	query := fmt.Sprintf(
		`SELECT * FROM cypher('ecosystem', $cypher$ MATCH (n) WHERE id(n) = %d RETURN label(n), properties(n) $cypher$) AS (label agtype, props agtype)`,
		id,
	)

	var labelRaw, propsRaw string
	err := q.QueryRow(ctx, query).Scan(&labelRaw, &propsRaw)
	if err != nil {
		return Node{}, fmt.Errorf("node %d not found: %w", id, err)
	}

	node := Node{ID: id}
	if err := json.Unmarshal([]byte(labelRaw), &node.Label); err != nil {
		return Node{}, fmt.Errorf("parsing label: %w", err)
	}
	if err := json.Unmarshal([]byte(propsRaw), &node.Properties); err != nil {
		return Node{}, fmt.Errorf("parsing properties: %w", err)
	}
	return node, nil
}

func UpdateNode(ctx context.Context, q Querier, id int64, props map[string]any) (Node, error) {
	cypherProps, err := toCypherMap(props)
	if err != nil {
		return Node{}, fmt.Errorf("invalid properties: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT * FROM cypher('ecosystem', $cypher$ MATCH (n) WHERE id(n) = %d SET n += %s RETURN label(n), properties(n) $cypher$) AS (label agtype, props agtype)`,
		id, cypherProps,
	)

	var labelRaw, propsRaw string
	if err := q.QueryRow(ctx, query).Scan(&labelRaw, &propsRaw); err != nil {
		return Node{}, fmt.Errorf("updating node %d: %w", id, err)
	}

	node := Node{ID: id}
	if err := json.Unmarshal([]byte(labelRaw), &node.Label); err != nil {
		return Node{}, fmt.Errorf("parsing label: %w", err)
	}
	if err := json.Unmarshal([]byte(propsRaw), &node.Properties); err != nil {
		return Node{}, fmt.Errorf("parsing properties: %w", err)
	}
	return node, nil
}

func DeleteNode(ctx context.Context, q Querier, id int64) error {
	query := fmt.Sprintf(
		`SELECT * FROM cypher('ecosystem', $cypher$ MATCH (n) WHERE id(n) = %d DETACH DELETE n $cypher$) AS (result agtype)`,
		id,
	)

	_, err := q.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("deleting node %d: %w", id, err)
	}
	return nil
}

func ListNodes(ctx context.Context, q Querier, label string) ([]Node, error) {
	var query string
	if label != "" {
		if err := ValidateNodeLabel(ctx, q, label); err != nil {
			return nil, err
		}
		query = fmt.Sprintf(
			`SELECT * FROM cypher('ecosystem', $cypher$ MATCH (n:%s) RETURN id(n), properties(n) $cypher$) AS (id agtype, props agtype)`,
			label,
		)
	} else {
		query = `SELECT * FROM cypher('ecosystem', $cypher$ MATCH (n) RETURN id(n), label(n), properties(n) $cypher$) AS (id agtype, label agtype, props agtype)`
	}

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var node Node
		var idRaw, propsRaw string
		if label != "" {
			if err := rows.Scan(&idRaw, &propsRaw); err != nil {
				return nil, err
			}
			node.Label = label
		} else {
			var labelRaw string
			if err := rows.Scan(&idRaw, &labelRaw, &propsRaw); err != nil {
				return nil, err
			}
			json.Unmarshal([]byte(labelRaw), &node.Label)
		}
		json.Unmarshal([]byte(idRaw), &node.ID)
		json.Unmarshal([]byte(propsRaw), &node.Properties)
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func SearchNodes(ctx context.Context, q Querier, label, namePattern string) ([]Node, error) {
	if err := ValidateNodeLabel(ctx, q, label); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		`SELECT * FROM cypher('ecosystem', $cypher$ MATCH (n:%s) WHERE n.name =~ '%s' RETURN id(n), properties(n) $cypher$) AS (id agtype, props agtype)`,
		label, escapeCypherString(cypherRegexFromLike(namePattern)),
	)

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var idRaw, propsRaw string
		if err := rows.Scan(&idRaw, &propsRaw); err != nil {
			return nil, err
		}
		var node Node
		node.Label = label
		json.Unmarshal([]byte(idRaw), &node.ID)
		json.Unmarshal([]byte(propsRaw), &node.Properties)
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func toCypherMap(props map[string]any) (string, error) {
	if len(props) == 0 {
		return "{}", nil
	}
	parts := make([]string, 0, len(props))
	for k, v := range props {
		if !validIdentifier.MatchString(k) {
			return "", fmt.Errorf("invalid property key: %q", k)
		}
		switch val := v.(type) {
		case string:
			parts = append(parts, fmt.Sprintf("%s: '%s'", k, escapeCypherString(val)))
		case float64, float32, int, int64, int32:
			parts = append(parts, fmt.Sprintf("%s: %v", k, val))
		case bool:
			parts = append(parts, fmt.Sprintf("%s: %v", k, val))
		default:
			j, _ := json.Marshal(val)
			parts = append(parts, fmt.Sprintf("%s: '%s'", k, escapeCypherString(string(j))))
		}
	}
	return "{" + strings.Join(parts, ", ") + "}", nil
}

func escapeCypherString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "$$", "")
	s = strings.ReplaceAll(s, "$cypher$", "")
	return s
}

func cypherRegexFromLike(pattern string) string {
	result := ""
	for _, c := range pattern {
		switch c {
		case '%':
			result += ".*"
		case '_':
			result += "."
		default:
			result += string(c)
		}
	}
	return result
}
