package service

import (
	"context"
	"encoding/json"
	"evo-ai-core-service/internal/infra/postgres"
	"evo-ai-core-service/pkg/agent_integration/model"
	"evo-ai-core-service/pkg/agent_integration/repository"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AgentIntegrationService interface {
	Upsert(ctx context.Context, agentID uuid.UUID, request model.AgentIntegrationRequest) (*model.AgentIntegrationResponse, error)
	GetByProvider(ctx context.Context, agentID uuid.UUID, provider string) (*model.AgentIntegrationResponse, error)
	ListByAgent(ctx context.Context, agentID uuid.UUID) ([]*model.AgentIntegrationResponse, error)
	Delete(ctx context.Context, agentID uuid.UUID, provider string) error
}

type agentIntegrationService struct {
	repository repository.AgentIntegrationRepository
}

func NewAgentIntegrationService(repository repository.AgentIntegrationRepository) AgentIntegrationService {
	return &agentIntegrationService{
		repository: repository,
	}
}

func (s *agentIntegrationService) Upsert(ctx context.Context, agentID uuid.UUID, request model.AgentIntegrationRequest) (*model.AgentIntegrationResponse, error) {
	// Convert map to datatypes.JSON
	configBytes, err := json.Marshal(request.Config)
	if err != nil {
		return nil, err
	}

	integration := model.AgentIntegration{
		AgentID:  agentID,
		Provider: request.Provider,
		Config:   datatypes.JSON(configBytes),
	}

	result, err := s.repository.Upsert(ctx, integration)
	if err != nil {
		return nil, postgres.MapDBError(err, model.AgentIntegrationErrors)
	}

	return result.ToResponse(), nil
}

func (s *agentIntegrationService) GetByProvider(ctx context.Context, agentID uuid.UUID, provider string) (*model.AgentIntegrationResponse, error) {
	integration, err := s.repository.GetByAgentAndProvider(ctx, agentID, provider)
	if err != nil {
		return nil, postgres.MapDBError(err, model.AgentIntegrationErrors)
	}

	return integration.ToResponse(), nil
}

func (s *agentIntegrationService) ListByAgent(ctx context.Context, agentID uuid.UUID) ([]*model.AgentIntegrationResponse, error) {
	integrations, err := s.repository.ListByAgent(ctx, agentID)
	if err != nil {
		return nil, postgres.MapDBError(err, model.AgentIntegrationErrors)
	}

	responses := make([]*model.AgentIntegrationResponse, len(integrations))
	for i, integration := range integrations {
		responses[i] = integration.ToResponse()
	}

	return responses, nil
}

func (s *agentIntegrationService) Delete(ctx context.Context, agentID uuid.UUID, provider string) error {
	err := s.repository.Delete(ctx, agentID, provider)
	if err != nil {
		return postgres.MapDBError(err, model.AgentIntegrationErrors)
	}

	return nil
}
