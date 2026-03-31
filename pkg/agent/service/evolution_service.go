package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"evo-ai-core-service/internal/utils/contextutils"
	"evo-ai-core-service/pkg/agent/client/evolution"
	"evo-ai-core-service/pkg/agent/model"

	"github.com/google/uuid"
)

type EvolutionService interface {
	CreateAgentBot(ctx context.Context, agent *model.Agent, aiProcessorURL string) (*evolution.AgentBot, error)
	UpdateAgentBot(ctx context.Context, agent *model.Agent, aiProcessorURL string) (*evolution.AgentBot, error)
	DeleteAgentBot(ctx context.Context, agent *model.Agent) error
	DeleteAgentBotSafe(ctx context.Context, agent *model.Agent) error
	SyncAgentBot(ctx context.Context, agent *model.Agent, aiProcessorURL string) (*evolution.AgentBot, error)
	CleanupEvolutionBot(ctx context.Context, botID uuid.UUID)
}

type evolutionService struct {
	evolutionClient *evolution.Client
}

func NewEvolutionService(evolutionBaseURL string) EvolutionService {
	return &evolutionService{
		evolutionClient: evolution.NewClient(evolutionBaseURL),
	}
}

func (s *evolutionService) CreateAgentBot(ctx context.Context, agent *model.Agent, aiProcessorURL string) (*evolution.AgentBot, error) {
	if agent.EvolutionBotSync {
		log.Printf("Agent %s already has Evolution bot sync enabled with bot ID %v", agent.ID, agent.EvolutionBotID)
		return nil, nil
	}

	log.Printf("Creating Evolution bot for agent %s", agent.ID)

	apiKey, err := s.getAgentAPIKey(agent)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent API key: %w", err)
	}

	// Get the original Bearer token from context
	token, err := contextutils.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Bearer token: %w", err)
	}

	// Extract advanced bot configuration from agent config
	advancedConfig := s.getAdvancedBotConfig(agent)

	evolutionBot, err := s.evolutionClient.CreateAgentBot(
		ctx,
		agent.ID,
		agent.Name,
		agent.Description,
		aiProcessorURL,
		apiKey,
		advancedConfig,
		token,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Evolution agent bot: %w", err)
	}

	log.Printf("Successfully created Evolution bot %s for agent %s", evolutionBot.ID, agent.ID)

	return evolutionBot, nil
}

func (s *evolutionService) UpdateAgentBot(ctx context.Context, agent *model.Agent, aiProcessorURL string) (*evolution.AgentBot, error) {
	// If agent has EvolutionBotID but sync is disabled, enable sync and update the existing bot
	if agent.EvolutionBotID != nil && !agent.EvolutionBotSync {
		log.Printf("Agent %s has existing Evolution bot ID %d but sync disabled, enabling sync and updating bot", agent.ID, *agent.EvolutionBotID)
		agent.EvolutionBotSync = true
		// Continue with update logic below
	} else if agent.EvolutionBotID == nil {
		log.Printf("Agent %s does not have Evolution bot, attempting to create new bot", agent.ID)
		// Try to create the bot if it doesn't exist
		return s.CreateAgentBot(ctx, agent, aiProcessorURL)
	}

	apiKey, err := s.getAgentAPIKey(agent)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent API key: %w", err)
	}

	// Get Bearer token from context
	bearerToken, err := contextutils.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Bearer token: %w", err)
	}

	// Extract advanced bot configuration from agent config
	advancedConfig := s.getAdvancedBotConfig(agent)

	evolutionBot, err := s.evolutionClient.UpdateAgentBot(
		ctx,
		*agent.EvolutionBotID,
		agent.ID,
		agent.Name,
		agent.Description,
		aiProcessorURL,
		apiKey,
		advancedConfig,
		bearerToken,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update Evolution agent bot: %w", err)
	}

	return evolutionBot, nil
}

func (s *evolutionService) DeleteAgentBot(ctx context.Context, agent *model.Agent) error {
	if !agent.EvolutionBotSync || agent.EvolutionBotID == nil {
		log.Printf("Agent %s does not have Evolution bot sync enabled", agent.ID)
		return nil
	}

	// Get the original Bearer token from context
	token, err := contextutils.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Bearer token: %w", err)
	}

	err = s.evolutionClient.DeleteAgentBot(ctx, *agent.EvolutionBotID, token)
	if err != nil {
		log.Printf("Failed to delete Evolution bot %s for agent %s: %v", *agent.EvolutionBotID, agent.ID, err)
		return fmt.Errorf("failed to delete Evolution bot %s: %w", *agent.EvolutionBotID, err)
	}

	log.Printf("Successfully deleted Evolution bot %s for agent %s", *agent.EvolutionBotID, agent.ID)
	return nil
}

func (s *evolutionService) DeleteAgentBotSafe(ctx context.Context, agent *model.Agent) error {
	if !agent.EvolutionBotSync || agent.EvolutionBotID == nil {
		log.Printf("Agent %s does not have Evolution bot sync enabled", agent.ID)
		return nil
	}

	// Get Bearer token from context, but allow deletion to proceed even without token
	bearerToken := ""
	if token, err := contextutils.GetToken(ctx); err == nil {
		bearerToken = token
	}

	err := s.evolutionClient.DeleteAgentBot(ctx, *agent.EvolutionBotID, bearerToken)
	if err != nil {
		log.Printf("Safe delete - Failed to delete Evolution bot %s for agent %s: %v", *agent.EvolutionBotID, agent.ID, err)
		// In safe mode, we don't return the error to allow agent deletion to proceed
		// This is useful for force delete operations or when Evolution is temporarily unavailable
		return nil
	}

	log.Printf("Successfully deleted Evolution bot %s for agent %s", *agent.EvolutionBotID, agent.ID)
	return nil
}

func (s *evolutionService) SyncAgentBot(ctx context.Context, agent *model.Agent, aiProcessorURL string) (*evolution.AgentBot, error) {
	if agent.EvolutionBotSync && agent.EvolutionBotID != nil {
		// Update existing bot
		return s.UpdateAgentBot(ctx, agent, aiProcessorURL)
	}
	// Create new bot
	return s.CreateAgentBot(ctx, agent, aiProcessorURL)
}

func (s *evolutionService) getAgentAPIKey(agent *model.Agent) (string, error) {
	// Try to extract API key from agent config first
	if agent.Config != "" {
		// Parse config and look for api_key
		config := make(map[string]interface{})
		if err := json.Unmarshal([]byte(agent.Config), &config); err == nil {
			if apiKey, ok := config["api_key"].(string); ok && apiKey != "" {
				return apiKey, nil
			}
		}
	}

	// Generate a simple API key if none exists
	return "evo-ai-bot-" + agent.ID.String(), nil
}

func (s *evolutionService) CleanupEvolutionBot(ctx context.Context, botID uuid.UUID) {
	// Get Bearer token from context (best effort for cleanup)
	bearerToken, err := contextutils.GetToken(ctx)
	if err != nil {
		log.Printf("No Bearer token available for cleanup: %v", err)
		return
	}

	err = s.evolutionClient.DeleteAgentBot(ctx, botID, bearerToken)
	if err != nil {
		log.Printf("Failed to cleanup Evolution bot %s: %v", botID, err)
	}
}

func (s *evolutionService) getAdvancedBotConfig(agent *model.Agent) *evolution.AdvancedBotConfig {
	// Extract advanced bot configuration from agent config
	config := &evolution.AdvancedBotConfig{
		// Default values
		MessageWaitTime:         5,
		MessageSignature:        "",
		EnableTextSegmentation:  false,
		MaxCharactersPerSegment: 300,
		MinSegmentSize:          50,
		CharacterDelayMS:        0.05,
	}

	if agent.Config == "" {
		return config
	}

	// Parse agent config JSON
	var agentConfig map[string]interface{}
	if err := json.Unmarshal([]byte(agent.Config), &agentConfig); err != nil {
		log.Printf("Failed to parse agent config for advanced bot settings: %v", err)
		return config
	}

	// Extract values with type checking and defaults
	if messageWaitTime, ok := agentConfig["message_wait_time"].(float64); ok {
		config.MessageWaitTime = int(messageWaitTime)
	}

	if messageSignature, ok := agentConfig["message_signature"].(string); ok {
		config.MessageSignature = messageSignature
	}

	if enableTextSegmentation, ok := agentConfig["enable_text_segmentation"].(bool); ok {
		config.EnableTextSegmentation = enableTextSegmentation
	}

	if maxCharactersPerSegment, ok := agentConfig["max_characters_per_segment"].(float64); ok {
		config.MaxCharactersPerSegment = int(maxCharactersPerSegment)
	}

	if minSegmentSize, ok := agentConfig["min_segment_size"].(float64); ok {
		config.MinSegmentSize = int(minSegmentSize)
	}

	if characterDelayMs, ok := agentConfig["character_delay_ms"].(float64); ok {
		config.CharacterDelayMS = characterDelayMs
	}

	if sendAsReply, ok := agentConfig["send_as_reply"].(bool); ok {
		config.SendAsReply = sendAsReply
	}

	// Extract inactivity actions
	if inactivityActions, ok := agentConfig["inactivity_actions"].([]interface{}); ok {
		config.InactivityActions = make([]map[string]interface{}, 0, len(inactivityActions))
		for _, action := range inactivityActions {
			if actionMap, ok := action.(map[string]interface{}); ok {
				config.InactivityActions = append(config.InactivityActions, actionMap)
			}
		}
	}

	// Extract transfer rules
	if transferRules, ok := agentConfig["transfer_rules"].([]interface{}); ok {
		config.TransferRules = make([]map[string]interface{}, 0, len(transferRules))
		for _, rule := range transferRules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				config.TransferRules = append(config.TransferRules, ruleMap)
			}
		}
	}

	// Extract pipeline rules
	if pipelineRules, ok := agentConfig["pipeline_rules"].([]interface{}); ok {
		config.PipelineRules = make([]map[string]interface{}, 0, len(pipelineRules))
		for _, rule := range pipelineRules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				config.PipelineRules = append(config.PipelineRules, ruleMap)
			}
		}
	}

	// Extract contact edit config
	if contactEditConfig, ok := agentConfig["contact_edit_config"].(map[string]interface{}); ok {
		config.ContactEditConfig = contactEditConfig
	}

	return config
}
