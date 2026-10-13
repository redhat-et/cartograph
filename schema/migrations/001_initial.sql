-- Creates the AGE graph, schema migration tracking, and type registry tables.

-- AGE graph
DO $$ BEGIN
    PERFORM create_graph('ecosystem');
EXCEPTION WHEN OTHERS THEN
    -- graph already exists
END $$;

-- Type registry: node types
CREATE TABLE IF NOT EXISTS node_types (
    label       TEXT PRIMARY KEY,
    description TEXT NOT NULL DEFAULT ''
);

-- Type registry: edge types
CREATE TABLE IF NOT EXISTS edge_types (
    name          TEXT PRIMARY KEY,
    valid_sources TEXT[] NOT NULL,
    valid_targets TEXT[] NOT NULL,
    description   TEXT NOT NULL DEFAULT ''
);

-- V1 node types
INSERT INTO node_types (label, description) VALUES
    ('Company',    'A legal entity — corporation, startup, VC firm'),
    ('Org',        'A division, BU, or team within a Company'),
    ('Person',     'An individual'),
    ('Technology', 'A product, platform, or OSS project');

-- V1 edge types
INSERT INTO edge_types (name, valid_sources, valid_targets, description) VALUES
    ('SUBSIDIARY_OF', '{Company}',      '{Company}',           'Ownership / subsidiary relationship'),
    ('PART_OF',       '{Org}',          '{Company,Org}',       'Organisational containment'),
    ('MEMBER_OF',     '{Person}',       '{Org}',               'Employment / team membership'),
    ('REPORTS_TO',    '{Person}',       '{Person}',            'Management chain'),
    ('INVESTS_IN',    '{Company}',      '{Company}',           'Financial investment'),
    ('ACQUIRED',      '{Company}',      '{Company}',           'Acquisition'),
    ('PARTNERS_WITH', '{Company}',      '{Company}',           'Strategic partnership, co-development'),
    ('SUPPLIES',      '{Company}',      '{Company}',           'Component or software supply relationship'),
    ('DEPLOYS',       '{Company}',      '{Technology}',        'Uses / deploys a technology'),
    ('DEVELOPS',      '{Company,Org}',  '{Technology}',        'Builds / maintains a technology'),
    ('LICENSES_FROM', '{Company}',      '{Company}',           'Technology licensing'),
    ('CONTACTED',     '{Person}',       '{Person}',            'A conversation or meeting occurred'),
    ('ADVOCATES_FOR', '{Person}',       '{Technology,Company}','Champions or promotes');
