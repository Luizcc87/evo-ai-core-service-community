CREATE TABLE IF NOT EXISTS evo_core_agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description     TEXT,
    type            VARCHAR(10) NOT NULL,
    model           VARCHAR(255),
    api_key_id      UUID REFERENCES evo_core_api_keys(id) ON DELETE SET NULL,
    instruction     TEXT,
    card_url        VARCHAR(1024) NOT NULL,
    folder_id       UUID REFERENCES evo_core_folders(id) ON DELETE SET NULL,
    config          JSON DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_agent_type CHECK (
        type::TEXT = ANY (
            ARRAY[
                'llm'::VARCHAR,
                'sequential'::VARCHAR,
                'parallel'::VARCHAR,
                'loop'::VARCHAR,
                'a2a'::VARCHAR,
                'workflow'::VARCHAR,
                'crew_ai'::VARCHAR,
                'task'::VARCHAR
            ]::TEXT[]
        )
    )
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_agents'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_agents'
        AND column_name = 'name'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_agents'
            AND indexname = 'idx_evo_core_agents_name'
        ) THEN
            CREATE INDEX idx_evo_core_agents_name ON evo_core_agents (name);
        END IF;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_agents'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_agents'
        AND column_name = 'name'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_agents'
            AND indexname = 'idx_evo_core_agents_name_unique'
        ) THEN
            CREATE UNIQUE INDEX idx_evo_core_agents_name_unique ON evo_core_agents (name);
        END IF;
    END IF;
END
$$;
