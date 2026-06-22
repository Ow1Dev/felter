CREATE TABLE IF NOT EXISTS schema_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id UUID NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('string', 'int', 'float', 'boolean', 'date', 'datetime')),
    created_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (schema_id, key)
);
