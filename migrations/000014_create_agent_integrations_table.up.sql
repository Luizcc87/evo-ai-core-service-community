CREATE TABLE IF NOT EXISTS evo_core_agent_integrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES evo_core_agents(id) ON DELETE CASCADE,
    provider VARCHAR(100) NOT NULL,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_agent_integration UNIQUE (agent_id, provider)
);

-- Create indexes for faster lookups
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE tablename = 'evo_core_agent_integrations'
        AND indexname = 'idx_evo_core_agent_integrations_agent'
    ) THEN
        CREATE INDEX idx_evo_core_agent_integrations_agent
        ON evo_core_agent_integrations (agent_id);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE tablename = 'evo_core_agent_integrations'
        AND indexname = 'idx_evo_core_agent_integrations_provider'
    ) THEN
        CREATE INDEX idx_evo_core_agent_integrations_provider
        ON evo_core_agent_integrations (provider);
    END IF;
END
$$;
