CREATE TABLE field_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id UUID NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,
    record_id UUID NOT NULL,
    field_key TEXT NOT NULL,
    value TEXT,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (record_id, field_key)
);

CREATE INDEX idx_field_values_schema ON field_values(schema_id);
CREATE INDEX idx_field_values_record ON field_values(record_id);
