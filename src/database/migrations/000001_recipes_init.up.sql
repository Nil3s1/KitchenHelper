-- The down migration preserves the context schema, so allow this migration to be reapplied.
CREATE SCHEMA IF NOT EXISTS recipes;

CREATE TABLE recipes.recipes (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL
);

CREATE INDEX recipes_payload_gin_idx
    ON recipes.recipes USING GIN (payload);
