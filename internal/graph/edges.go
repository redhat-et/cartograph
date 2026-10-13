package graph

import (
	"context"
	"encoding/json"
	"fmt"

)

type Edge struct {
	ID         int64
	Type       string
	SourceID   int64
	TargetID   int64
	Properties map[string]any
}

func CreateEdge(ctx context.Context, q Querier, edgeType string, sourceID, targetID int64, props map[string]any) (Edge, error) {
	sourceNode, err := GetNode(ctx, q, sourceID)
	if err != nil {
		return Edge{}, fmt.Errorf("source node: %w", err)
	}
	targetNode, err := GetNode(ctx, q, targetID)
	if err != nil {
		return Edge{}, fmt.Errorf("target node: %w", err)
	}

	if err := ValidateEdge(ctx, q, edgeType, sourceNode.Label, targetNode.Label); err != nil {
		return Edge{}, err
	}

	if props == nil {
		props = map[string]any{}
	}
	cypherProps, err := toCypherMap(props)
	if err != nil {
		return Edge{}, fmt.Errorf("invalid properties: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT * FROM cypher('ecosystem', $cypher$
			MATCH (a), (b)
			WHERE id(a) = %d AND id(b) = %d
			CREATE (a)-[e:%s %s]->(b)
			RETURN id(e)
		$cypher$) AS (id agtype)`,
		sourceID, targetID, edgeType, cypherProps,
	)

	var idRaw string
	if err := q.QueryRow(ctx, query).Scan(&idRaw); err != nil {
		return Edge{}, fmt.Errorf("creating edge: %w", err)
	}

	edge := Edge{Type: edgeType, SourceID: sourceID, TargetID: targetID, Properties: props}
	json.Unmarshal([]byte(idRaw), &edge.ID)
	return edge, nil
}

func GetEdge(ctx context.Context, q Querier, id int64) (Edge, error) {
	query := fmt.Sprintf(
		`SELECT * FROM cypher('ecosystem', $cypher$
			MATCH (a)-[e]->(b)
			WHERE id(e) = %d
			RETURN label(e), id(a), id(b), properties(e)
		$cypher$) AS (type agtype, src agtype, tgt agtype, props agtype)`,
		id,
	)

	var typeRaw, srcRaw, tgtRaw, propsRaw string
	if err := q.QueryRow(ctx, query).Scan(&typeRaw, &srcRaw, &tgtRaw, &propsRaw); err != nil {
		return Edge{}, fmt.Errorf("edge %d not found: %w", id, err)
	}

	edge := Edge{ID: id}
	json.Unmarshal([]byte(typeRaw), &edge.Type)
	json.Unmarshal([]byte(srcRaw), &edge.SourceID)
	json.Unmarshal([]byte(tgtRaw), &edge.TargetID)
	json.Unmarshal([]byte(propsRaw), &edge.Properties)
	return edge, nil
}

func DeleteEdge(ctx context.Context, q Querier, id int64) error {
	query := fmt.Sprintf(
		`SELECT * FROM cypher('ecosystem', $cypher$
			MATCH ()-[e]->()
			WHERE id(e) = %d
			DELETE e
		$cypher$) AS (result agtype)`,
		id,
	)
	_, err := q.Exec(ctx, query)
	return err
}

func ListEdges(ctx context.Context, q Querier, edgeType string) ([]Edge, error) {
	var query string
	if edgeType != "" {
		if err := ValidateEdgeType(ctx, q, edgeType); err != nil {
			return nil, err
		}
		query = fmt.Sprintf(
			`SELECT * FROM cypher('ecosystem', $cypher$
				MATCH (a)-[e:%s]->(b)
				RETURN id(e), id(a), id(b), properties(e)
			$cypher$) AS (id agtype, src agtype, tgt agtype, props agtype)`,
			edgeType,
		)
	} else {
		query = `SELECT * FROM cypher('ecosystem', $cypher$
			MATCH (a)-[e]->(b)
			RETURN id(e), label(e), id(a), id(b), properties(e)
		$cypher$) AS (id agtype, type agtype, src agtype, tgt agtype, props agtype)`
	}

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var edge Edge
		var idRaw, srcRaw, tgtRaw, propsRaw string
		if edgeType != "" {
			if err := rows.Scan(&idRaw, &srcRaw, &tgtRaw, &propsRaw); err != nil {
				return nil, err
			}
			edge.Type = edgeType
		} else {
			var typeRaw string
			if err := rows.Scan(&idRaw, &typeRaw, &srcRaw, &tgtRaw, &propsRaw); err != nil {
				return nil, err
			}
			json.Unmarshal([]byte(typeRaw), &edge.Type)
		}
		json.Unmarshal([]byte(idRaw), &edge.ID)
		json.Unmarshal([]byte(srcRaw), &edge.SourceID)
		json.Unmarshal([]byte(tgtRaw), &edge.TargetID)
		json.Unmarshal([]byte(propsRaw), &edge.Properties)
		edges = append(edges, edge)
	}
	return edges, rows.Err()
}
