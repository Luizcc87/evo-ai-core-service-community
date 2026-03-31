package middleware

import (
	"net/http"

	agentService "evo-ai-core-service/pkg/agent/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AgentAccessMiddleware interface {
	GetAgentAccessMiddleware() gin.HandlerFunc
}

type agentAccessMiddleware struct {
	agentService agentService.AgentService
}

func NewAgentAccessMiddleware(agentService agentService.AgentService) AgentAccessMiddleware {
	return &agentAccessMiddleware{agentService: agentService}
}

func bodyToMap(c *gin.Context) (map[string]interface{}, error) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		return nil, err
	}

	return body, nil
}

func (a *agentAccessMiddleware) GetAgentAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var agentID uuid.UUID
		var errAgentID error

		switch c.Request.Method {
		case http.MethodPost:
			agentID, errAgentID = uuid.Parse(c.Param("id"))
		case http.MethodDelete, http.MethodPut:
			agentID, errAgentID = uuid.Parse(c.Param("id"))
		default:
			agentID, errAgentID = uuid.Parse(c.Param("id"))
		}

		if errAgentID != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
			return
		}

		_, err := a.agentService.GetByID(ctx, agentID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
			return
		}

		c.Next()
	}
}
