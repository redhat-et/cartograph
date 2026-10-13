#!/usr/bin/env python3
"""One-time transform: convert intel repo data to Cartograph seed format.

Usage:
    python3 -I tools/transform-intel-data.py \
        --input-dir /path/to/intel/data/ \
        --output-dir schema/seed/
"""
import argparse
import json
import sys
from pathlib import Path

EDGE_TYPE_MAP = {
    "invests-in": "INVESTS_IN",
    "acquired": "ACQUIRED",
    "supplies-component": "SUPPLIES",
    "supplies-software": "SUPPLIES",
    "integrates-for": "PARTNERS_WITH",
    "co-develops": "PARTNERS_WITH",
    "licenses-from": "LICENSES_FROM",
    "deploys-fleet-of": "DEPLOYS",
    "trains-on-platform": "DEPLOYS",
}

KEEP_NODE_PROPS = [
    "name", "hq_country", "founded_year", "revenue_usd", "funding_usd",
    "headcount", "key_products", "sub_segments",
]


def transform_nodes(intel_nodes: list[dict]) -> list[dict]:
    result = []
    for node in intel_nodes:
        props = {}
        for key in KEEP_NODE_PROPS:
            if key in node and node[key] is not None:
                props[key] = node[key]
        if "name" not in props:
            print(f"WARNING: skipping node without name: {node}", file=sys.stderr)
            continue
        result.append({"label": "Company", "properties": props})
    return result


def transform_edges(intel_edges: list[dict]) -> list[dict]:
    result = []
    for edge in intel_edges:
        edge_type = EDGE_TYPE_MAP.get(edge.get("type"))
        if edge_type is None:
            print(f"WARNING: unknown edge type '{edge.get('type')}', skipping: {edge}", file=sys.stderr)
            continue

        props = {}
        for key in ("annotation", "date", "source_uri", "confidence"):
            if key in edge and edge[key] is not None:
                props[key] = edge[key]

        result.append({
            "type": edge_type,
            "source": edge["source"],
            "target": edge["target"],
            "properties": props,
        })
    return result


def main():
    parser = argparse.ArgumentParser(description="Transform intel repo data to Cartograph seed format")
    parser.add_argument("--input-dir", required=True, help="Path to intel data directory")
    parser.add_argument("--output-dir", required=True, help="Path to output directory")
    args = parser.parse_args()

    input_dir = Path(args.input_dir)
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    with open(input_dir / "nodes.json") as f:
        intel_nodes = json.load(f)
    with open(input_dir / "edges.json") as f:
        intel_edges = json.load(f)

    nodes = transform_nodes(intel_nodes)
    edges = transform_edges(intel_edges)

    with open(output_dir / "nodes.json", "w") as f:
        json.dump(nodes, f, indent=2)
    with open(output_dir / "edges.json", "w") as f:
        json.dump(edges, f, indent=2)

    print(f"Transformed {len(nodes)} nodes and {len(edges)} edges")
    print(f"Output: {output_dir / 'nodes.json'}, {output_dir / 'edges.json'}")


if __name__ == "__main__":
    main()
