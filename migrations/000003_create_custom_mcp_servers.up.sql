CREATE TABLE IF NOT EXISTS evo_core_custom_mcp_servers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(1024) NOT NULL,
    headers JSON NOT NULL,
    timeout INTEGER NOT NULL DEFAULT 0,
    retry_count INTEGER NOT NULL DEFAULT 0,
    tags VARCHAR(255)[] NOT NULL DEFAULT '{}',
    tools JSON NOT NULL DEFAULT '{}',
    is_active    BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_custom_mcp_servers'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_custom_mcp_servers'
        AND column_name = 'name'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_custom_mcp_servers'
            AND indexname = 'idx_evo_core_custom_mcp_servers_name'
        ) THEN
            CREATE INDEX idx_evo_core_custom_mcp_servers_name ON evo_core_custom_mcp_servers (name);
        END IF;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_custom_mcp_servers'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_custom_mcp_servers'
        AND column_name = 'is_active'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_custom_mcp_servers'
            AND indexname = 'idx_evo_core_custom_mcp_servers_is_active'
        ) THEN
            CREATE INDEX idx_evo_core_custom_mcp_servers_is_active ON evo_core_custom_mcp_servers (is_active);
        END IF;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'evo_core_custom_mcp_servers'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'evo_core_custom_mcp_servers'
        AND column_name = 'name'
    ) THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'evo_core_custom_mcp_servers'
            AND indexname = 'idx_evo_core_custom_mcp_servers_name_unique'
        ) THEN
            CREATE UNIQUE INDEX idx_evo_core_custom_mcp_servers_name_unique ON evo_core_custom_mcp_servers (name);
        END IF;
    END IF;
END
$$;
