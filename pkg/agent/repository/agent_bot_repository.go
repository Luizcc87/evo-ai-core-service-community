package repository

import (
	"context"
	"evo-ai-core-service/pkg/agent/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgentBotRepository interface {
	Create(ctx context.Context, bot model.AgentBot) (*model.AgentBot, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.AgentBot, error)
	Update(ctx context.Context, bot *model.AgentBot, id uuid.UUID) (*model.AgentBot, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type agentBotRepository struct {
	db *gorm.DB
}

func NewAgentBotRepository(db *gorm.DB) AgentBotRepository {
	return &agentBotRepository{db: db}
}

func (r *agentBotRepository) Create(ctx context.Context, bot model.AgentBot) (*model.AgentBot, error) {
	if err := r.db.WithContext(ctx).Create(&bot).Error; err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *agentBotRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.AgentBot, error) {
	var bot model.AgentBot
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&bot).Error; err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *agentBotRepository) Update(ctx context.Context, bot *model.AgentBot, id uuid.UUID) (*model.AgentBot, error) {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Updates(bot).First(bot).Error; err != nil {
		return nil, err
	}
	return bot, nil
}

func (r *agentBotRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AgentBot{}).Error; err != nil {
		return err
	}
	return nil
}
