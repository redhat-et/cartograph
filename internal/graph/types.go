package graph

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NodeType struct {
	Label       string
	Description string
}

type EdgeType struct {
	Name         string
	ValidSources []string
	ValidTargets []string
	Description  string
}

func ListNodeTypes(ctx context.Context, pool *pgxpool.Pool) ([]NodeType, error) {
	rows, err := pool.Query(ctx, "SELECT label, description FROM node_types ORDER BY label")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []NodeType
	for rows.Next() {
		var nt NodeType
		if err := rows.Scan(&nt.Label, &nt.Description); err != nil {
			return nil, err
		}
		types = append(types, nt)
	}
	return types, rows.Err()
}

func ListEdgeTypes(ctx context.Context, pool *pgxpool.Pool) ([]EdgeType, error) {
	rows, err := pool.Query(ctx, "SELECT name, valid_sources, valid_targets, description FROM edge_types ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []EdgeType
	for rows.Next() {
		var et EdgeType
		if err := rows.Scan(&et.Name, &et.ValidSources, &et.ValidTargets, &et.Description); err != nil {
			return nil, err
		}
		types = append(types, et)
	}
	return types, rows.Err()
}

func ValidateNodeLabel(ctx context.Context, pool *pgxpool.Pool, label string) error {
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM node_types WHERE label = $1)", label).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("unknown node label: %q", label)
	}
	return nil
}

func ValidateEdge(ctx context.Context, pool *pgxpool.Pool, edgeType, sourceLabel, targetLabel string) error {
	var validSources, validTargets []string
	err := pool.QueryRow(ctx,
		"SELECT valid_sources, valid_targets FROM edge_types WHERE name = $1",
		edgeType).Scan(&validSources, &validTargets)
	if err != nil {
		return fmt.Errorf("unknown edge type: %q", edgeType)
	}

	if !contains(validSources, sourceLabel) {
		return fmt.Errorf("edge type %q does not allow source label %q (allowed: %v)", edgeType, sourceLabel, validSources)
	}
	if !contains(validTargets, targetLabel) {
		return fmt.Errorf("edge type %q does not allow target label %q (allowed: %v)", edgeType, targetLabel, validTargets)
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
