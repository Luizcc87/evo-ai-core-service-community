package agent

import (
	"evo-ai-core-service/internal/config"
	"evo-ai-core-service/pkg/agent/handler"
	"evo-ai-core-service/pkg/agent/repository"
	"evo-ai-core-service/pkg/agent/service"
	apiKeyService "evo-ai-core-service/pkg/api_key/service"
	customMCPServerService "evo-ai-core-service/pkg/custom_mcp_server/service"
	customToolService "evo-ai-core-service/pkg/custom_tool/service"
	folderService "evo-ai-core-service/pkg/folder/service"
	folderShareService "evo-ai-core-service/pkg/folder_share/service"
	mcpServerService "evo-ai-core-service/pkg/mcp_server/service"

	"gorm.io/gorm"
)

type Module struct {
	Handler handler.AgentHandler
	Service service.AgentService
	Repo    repository.AgentRepository
}

func New(
	db *gorm.DB,
	aiProcessorServiceConfig *config.AIProcessorServiceConfig,
	folderService folderService.FolderService,
	apiKeyService apiKeyService.ApiKeyService,
	mcpServerService mcpServerService.McpServerService,
	customToolService customToolService.CustomToolService,
	customMCPServerService customMCPServerService.CustomMcpServerService,
	folderShareService folderShareService.FolderShareService,
) *Module {
	r := repository.NewAgentRepository(db)
	agentBotRepo := repository.NewAgentBotRepository(db)
	evolutionService := service.NewEvolutionService(agentBotRepo)
	s := service.NewAgentService(r,
		folderService,
		apiKeyService,
		mcpServerService,
		customToolService,
		customMCPServerService,
		evolutionService,
		aiProcessorServiceConfig.URL,
	)
	h := handler.NewAgentHandler(s, aiProcessorServiceConfig, folderShareService)

	return &Module{
		Handler: h,
		Service: s,
		Repo:    r,
	}
}
