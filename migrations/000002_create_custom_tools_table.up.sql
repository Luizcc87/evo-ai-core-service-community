CREATE TABLE IF NOT EXISTS evo_core_custom_tools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(1024) NOT NULL,
    headers json NOT NULL,
    path_params json NOT NULL,
    query_params json NOT NULL,
    body_params json NOT NULL,
    error_handling json NOT NULL,
    values json NOT NULL,
    tags VARCHAR(255)[] NOT NULL DEFAULT '{}',
    examples VARCHAR(255)[] NOT NULL DEFAULT '{}',
    input_modes VARCHAR(255)[] NOT NULL DEFAULT '{}',
    output_modes VARCHAR(255)[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_custom_tools'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_custom_tools'
        AND column_name = 'name'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_custom_tools'
            AND indexname = 'idx_evo_core_custom_tools_name'
        ) THEN
            CREATE INDEX idx_evo_core_custom_tools_name ON evo_core_custom_tools (name);
        END IF;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_custom_tools'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_custom_tools'
        AND column_name = 'is_active'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_custom_tools'
            AND indexname = 'idx_evo_core_custom_tools_is_active'
        ) THEN
            CREATE INDEX idx_evo_core_custom_tools_is_active ON evo_core_custom_tools (is_active);
        END IF;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_custom_tools'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_custom_tools'
        AND column_name = 'name'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_custom_tools'
            AND indexname = 'idx_evo_core_custom_tools_name_unique'
        ) THEN
            CREATE UNIQUE INDEX idx_evo_core_custom_tools_name_unique ON evo_core_custom_tools (name);
        END IF;
    END IF;
END
$$;
