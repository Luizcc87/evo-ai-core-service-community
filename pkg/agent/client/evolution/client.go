package evolution

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"evo-ai-core-service/internal/httpclient"
	"evo-ai-core-service/internal/utils/stringutils"

	"github.com/google/uuid"
)

type Client struct {
	baseURL string
}

type AgentBotRequest struct {
	Name                    string                 `json:"name"`
	Description             string                 `json:"description"`
	OutgoingURL             string                 `json:"outgoing_url"`
	APIKey                  string                 `json:"api_key"`
	MessageSignature        string                 `json:"message_signature"`
	TextSegmentationEnabled bool                   `json:"text_segmentation_enabled"`
	TextSegmentationLimit   int                    `json:"text_segmentation_limit"`
	TextSegmentationMinSize int                    `json:"text_segmentation_min_size"`
	DelayPerCharacter       float64                `json:"delay_per_character"`
	DebounceTime            int                    `json:"debounce_time"`
	BotType                 string                 `json:"bot_type"`
	BotProvider             string                 `json:"bot_provider"`
	BotConfig               map[string]interface{} `json:"bot_config"`
}

// FlexibleFloat handles fields that can be string or float64
type FlexibleFloat float64

func (f *FlexibleFloat) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		val, parseErr := strconv.ParseFloat(str, 64)
		if parseErr != nil {
			return parseErr
		}
		*f = FlexibleFloat(val)
		return nil
	}

	var num float64
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*f = FlexibleFloat(num)
	return nil
}

type AgentBot struct {
	ID                      uuid.UUID              `json:"id"`
	Name                    string                 `json:"name"`
	Description             string                 `json:"description"`
	OutgoingURL             string                 `json:"outgoing_url"`
	APIKey                  string                 `json:"api_key"`
	MessageSignature        string                 `json:"message_signature"`
	TextSegmentationEnabled bool                   `json:"text_segmentation_enabled"`
	TextSegmentationLimit   int                    `json:"text_segmentation_limit"`
	TextSegmentationMinSize int                    `json:"text_segmentation_min_size"`
	DelayPerCharacter       FlexibleFloat          `json:"delay_per_character"`
	DebounceTime            int                    `json:"debounce_time"`
	BotType                 string                 `json:"bot_type"`
	BotProvider             string                 `json:"bot_provider"`
	BotConfig               map[string]interface{} `json:"bot_config"`
	AvatarURL               string                 `json:"avatar_url,omitempty"`
	AccessToken             string                 `json:"access_token,omitempty"`
	SystemBot               bool                   `json:"system_bot,omitempty"`
	CreatedAt               time.Time              `json:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at"`
}

type AgentBotResponse struct {
	Success bool                   `json:"success"`
	Data    AgentBot               `json:"data"`
	Meta    map[string]interface{} `json:"meta"`
	Message string                 `json:"message,omitempty"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
	}
}

type AdvancedBotConfig struct {
	MessageWaitTime         int                      `json:"message_wait_time"`
	MessageSignature        string                   `json:"message_signature"`
	EnableTextSegmentation  bool                     `json:"enable_text_segmentation"`
	MaxCharactersPerSegment int                      `json:"max_characters_per_segment"`
	MinSegmentSize          int                      `json:"min_segment_size"`
	CharacterDelayMS        float64                  `json:"character_delay_ms"`
	SendAsReply             bool                     `json:"send_as_reply"`
	InactivityActions       []map[string]interface{} `json:"inactivity_actions,omitempty"`
	TransferRules           []map[string]interface{} `json:"transfer_rules,omitempty"`
	PipelineRules           []map[string]interface{} `json:"pipeline_rules,omitempty"`
	ContactEditConfig       map[string]interface{}   `json:"contact_edit_config,omitempty"`
}

func (r *AgentBotRequest) toMap() map[string]interface{} {
	jsonStr := stringutils.ToJSON(*r)
	return stringutils.JSONToInterfaceMap(jsonStr)
}

func (c *Client) CreateAgentBot(ctx context.Context, agentID uuid.UUID, agentName, agentDescription, aiProcessorURL, apiKey string, advancedConfig *AdvancedBotConfig, bearerToken string) (*AgentBot, error) {
	url := fmt.Sprintf("%s/api/v1/agent_bots", c.baseURL)

	outgoingURL := fmt.Sprintf("%s/api/v1/a2a/%s", aiProcessorURL, agentID)

	// Use provided config or fallback to defaults
	var messageSignature string
	var textSegmentationEnabled bool
	var textSegmentationLimit int
	var textSegmentationMinSize int
	var delayPerCharacter float64
	var debounceTime int

	if advancedConfig != nil {
		messageSignature = advancedConfig.MessageSignature
		textSegmentationEnabled = advancedConfig.EnableTextSegmentation
		textSegmentationLimit = advancedConfig.MaxCharactersPerSegment
		textSegmentationMinSize = advancedConfig.MinSegmentSize
		delayPerCharacter = advancedConfig.CharacterDelayMS
		debounceTime = advancedConfig.MessageWaitTime
	} else {
		// Default values
		messageSignature = ""
		textSegmentationEnabled = false
		textSegmentationLimit = 300
		textSegmentationMinSize = 50
		delayPerCharacter = 0.05
		debounceTime = 5
	}

	// Build bot_config with all configurations
	botConfig := make(map[string]interface{})
	if advancedConfig != nil {
		if advancedConfig.SendAsReply {
			botConfig["send_as_reply"] = true
		}
		if len(advancedConfig.InactivityActions) > 0 {
			botConfig["inactivity_actions"] = advancedConfig.InactivityActions
		}
		if len(advancedConfig.TransferRules) > 0 {
			botConfig["transfer_rules"] = advancedConfig.TransferRules
		}
		if len(advancedConfig.PipelineRules) > 0 {
			botConfig["pipeline_rules"] = advancedConfig.PipelineRules
		}
		if len(advancedConfig.ContactEditConfig) > 0 {
			botConfig["contact_edit_config"] = advancedConfig.ContactEditConfig
		}
	}

	request := AgentBotRequest{
		Name:                    agentName,
		Description:             agentDescription,
		OutgoingURL:             outgoingURL,
		APIKey:                  apiKey,
		MessageSignature:        messageSignature,
		TextSegmentationEnabled: textSegmentationEnabled,
		TextSegmentationLimit:   textSegmentationLimit,
		TextSegmentationMinSize: textSegmentationMinSize,
		DelayPerCharacter:       delayPerCharacter,
		DebounceTime:            debounceTime,
		BotType:                 "webhook",
		BotProvider:             "evo_ai_provider",
		BotConfig:               botConfig,
	}

	fmt.Printf("[EVOLUTION] Creating bot - URL: %s\n", url)
	fmt.Printf("[EVOLUTION] Request: %+v\n", request)

	// Prepare headers with Bearer token
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	// Use Bearer token for OAuth authentication
	if bearerToken != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", bearerToken)
	}

	agentBotResponse, err := httpclient.DoPostJSON[AgentBotResponse](ctx, url, request.toMap(), headers, http.StatusCreated)
	if err != nil {
		fmt.Printf("[EVOLUTION] Request failed: %v\n", err)
		return nil, fmt.Errorf("failed to create agent bot: %w", err)
	}

	agentBot := &agentBotResponse.Data
	fmt.Printf("[EVOLUTION] Bot created successfully: ID=%s, Name=%s\n", agentBot.ID, agentBot.Name)
	return agentBot, nil
}

func (c *Client) DeleteAgentBot(ctx context.Context, agentBotID uuid.UUID, bearerToken string) error {
	url := fmt.Sprintf("%s/api/v1/agent_bots/%s", c.baseURL, agentBotID)

	// Prepare headers with Bearer token
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	// Use Bearer token for OAuth authentication
	if bearerToken != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", bearerToken)
	}

	_, err := httpclient.DoDeleteJSON[map[string]interface{}](ctx, url, nil, headers, http.StatusOK)
	if err != nil {
		return fmt.Errorf("failed to delete agent bot: %w", err)
	}

	return nil
}

func (c *Client) UpdateAgentBot(ctx context.Context, agentBotID uuid.UUID, agentID uuid.UUID, agentName, agentDescription, aiProcessorURL, apiKey string, advancedConfig *AdvancedBotConfig, bearerToken string) (*AgentBot, error) {
	url := fmt.Sprintf("%s/api/v1/agent_bots/%s", c.baseURL, agentBotID)

	outgoingURL := fmt.Sprintf("%s/api/v1/a2a/%s", aiProcessorURL, agentID)

	// Use provided config or fallback to defaults
	var messageSignature string
	var textSegmentationEnabled bool
	var textSegmentationLimit int
	var textSegmentationMinSize int
	var delayPerCharacter float64
	var debounceTime int

	if advancedConfig != nil {
		messageSignature = advancedConfig.MessageSignature
		textSegmentationEnabled = advancedConfig.EnableTextSegmentation
		textSegmentationLimit = advancedConfig.MaxCharactersPerSegment
		textSegmentationMinSize = advancedConfig.MinSegmentSize
		delayPerCharacter = advancedConfig.CharacterDelayMS
		debounceTime = advancedConfig.MessageWaitTime
	} else {
		// Default values
		messageSignature = ""
		textSegmentationEnabled = false
		textSegmentationLimit = 300
		textSegmentationMinSize = 50
		delayPerCharacter = 0.05
		debounceTime = 5
	}

	// Build bot_config with all configurations
	botConfig := make(map[string]interface{})
	if advancedConfig != nil {
		if advancedConfig.SendAsReply {
			botConfig["send_as_reply"] = true
		}
		if len(advancedConfig.InactivityActions) > 0 {
			botConfig["inactivity_actions"] = advancedConfig.InactivityActions
		}
		if len(advancedConfig.TransferRules) > 0 {
			botConfig["transfer_rules"] = advancedConfig.TransferRules
		}
		if len(advancedConfig.PipelineRules) > 0 {
			botConfig["pipeline_rules"] = advancedConfig.PipelineRules
		}
		if len(advancedConfig.ContactEditConfig) > 0 {
			botConfig["contact_edit_config"] = advancedConfig.ContactEditConfig
		}
	}

	request := AgentBotRequest{
		Name:                    agentName,
		Description:             agentDescription,
		OutgoingURL:             outgoingURL,
		APIKey:                  apiKey,
		MessageSignature:        messageSignature,
		TextSegmentationEnabled: textSegmentationEnabled,
		TextSegmentationLimit:   textSegmentationLimit,
		TextSegmentationMinSize: textSegmentationMinSize,
		DelayPerCharacter:       delayPerCharacter,
		DebounceTime:            debounceTime,
		BotType:                 "webhook",
		BotProvider:             "evo_ai_provider",
		BotConfig:               botConfig,
	}

	// Prepare headers with Bearer token
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}

	// Use Bearer token for OAuth authentication
	if bearerToken != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", bearerToken)
	}

	agentBotResponse, err := httpclient.DoPutJSON[AgentBotResponse](ctx, url, request.toMap(), headers, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("failed to update agent bot: %w", err)
	}

	agentBot := &agentBotResponse.Data
	fmt.Printf("[EVOLUTION] Bot updated successfully: ID=%s, Name=%s\n", agentBot.ID, agentBot.Name)
	return agentBot, nil
}
