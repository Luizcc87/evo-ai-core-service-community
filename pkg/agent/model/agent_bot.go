package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AgentBot struct {
	ID                      uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name                    string    `json:"name" gorm:"type:varchar(255)"`
	Description             string    `json:"description" gorm:"type:varchar(255)"`
	OutgoingURL             string    `json:"outgoing_url" gorm:"type:varchar(1024)"`
	BotType                 int       `json:"bot_type" gorm:"type:integer;default:0"`
	BotConfig               string    `json:"bot_config" gorm:"type:jsonb;default:'{}'"`
	APIKey                  string    `json:"api_key" gorm:"column:api_key;type:varchar(1000)"`
	BotProvider             string    `json:"bot_provider" gorm:"type:varchar(255);default:'webhook';not null"`
	MessageSignature        string    `json:"message_signature" gorm:"type:text"`
	TextSegmentationEnabled bool      `json:"text_segmentation_enabled" gorm:"type:boolean;default:false;not null"`
	TextSegmentationLimit   int       `json:"text_segmentation_limit" gorm:"type:integer;default:300"`
	TextSegmentationMinSize int       `json:"text_segmentation_min_size" gorm:"type:integer;default:50"`
	DelayPerCharacter       float64   `json:"delay_per_character" gorm:"type:decimal(8,2);default:50.0"`
	DebounceTime            int       `json:"debounce_time" gorm:"type:integer;default:5;not null"`
	CreatedAt               time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt               time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`
}

func (AgentBot) TableName() string {
	return "agent_bots"
}

func (a *AgentBot) GetBotConfigMap() map[string]interface{} {
	if a.BotConfig == "" || a.BotConfig == "{}" {
		return make(map[string]interface{})
	}
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(a.BotConfig), &config); err != nil {
		return make(map[string]interface{})
	}
	return config
}

func (a *AgentBot) SetBotConfigMap(config map[string]interface{}) {
	if config == nil || len(config) == 0 {
		a.BotConfig = "{}"
		return
	}
	data, err := json.Marshal(config)
	if err != nil {
		a.BotConfig = "{}"
		return
	}
	a.BotConfig = string(data)
}
