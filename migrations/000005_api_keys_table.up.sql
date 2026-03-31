CREATE TABLE IF NOT EXISTS evo_core_api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    key text NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_api_keys'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_api_keys'
        AND column_name = 'name'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_api_keys'
            AND indexname = 'idx_evo_core_api_keys_name'
        ) THEN
            CREATE INDEX idx_evo_core_api_keys_name ON evo_core_api_keys (name);
        END IF;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_api_keys'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_api_keys'
        AND column_name = 'is_active'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_api_keys'
            AND indexname = 'idx_evo_core_api_keys_is_active'
        ) THEN
            CREATE INDEX idx_evo_core_api_keys_is_active ON evo_core_api_keys (is_active);
        END IF;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_api_keys'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_api_keys'
        AND column_name = 'name'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_api_keys'
            AND indexname = 'idx_evo_core_api_keys_name_unique'
        ) THEN
            CREATE UNIQUE INDEX idx_evo_core_api_keys_name_unique ON evo_core_api_keys (name);
        END IF;
    END IF;
END
$$;
