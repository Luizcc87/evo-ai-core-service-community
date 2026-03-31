package handler

import (
	"evo-ai-core-service/internal/httpclient/errors"
	"evo-ai-core-service/internal/httpclient/response"
	"evo-ai-core-service/internal/middleware"
	"evo-ai-core-service/internal/utils/stringutils"
	"evo-ai-core-service/pkg/custom_mcp_server/model"
	"evo-ai-core-service/pkg/custom_mcp_server/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CustomMcpServerHandler interface defines the contract for custom mcp server handlers
type CustomMcpServerHandler interface {
	RegisterRoutesMiddleware(router gin.IRouter)
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Test(c *gin.Context)
}

// customMcpServerHandler implements the CustomMcpServerHandler interface.
type customMcpServerHandler struct {
	customMcpServerService service.CustomMcpServerService
}

// NewCustomMcpServerHandler creates a new custom mcp server handler
func NewCustomMcpServerHandler(customMcpServerService service.CustomMcpServerService) CustomMcpServerHandler {
	return &customMcpServerHandler{
		customMcpServerService: customMcpServerService,
	}
}

// RegisterRoutesMiddleware registers the routes for the custom mcp server handler with middleware
func (h *customMcpServerHandler) RegisterRoutesMiddleware(router gin.IRouter) {
	// Usar middleware global de permissão
	permissionMiddleware := middleware.GetGlobalPermissionMiddleware()

	customMcpServers := router.Group("/custom-mcp-servers")
	{
		// Read permissions
		customMcpServers.GET("",
			permissionMiddleware.RequirePermission("ai_custom_mcp_servers", "read"),
			h.List)
		customMcpServers.GET("/",
			permissionMiddleware.RequirePermission("ai_custom_mcp_servers", "read"),
			h.List)
		customMcpServers.GET("/:id",
			permissionMiddleware.RequirePermission("ai_custom_mcp_servers", "read"),
			h.GetByID)

		// Create permissions
		customMcpServers.POST("",
			permissionMiddleware.RequirePermission("ai_custom_mcp_servers", "create"),
			h.Create)
		customMcpServers.POST("/",
			permissionMiddleware.RequirePermission("ai_custom_mcp_servers", "create"),
			h.Create)

		// Update permissions
		customMcpServers.PUT("/:id",
			permissionMiddleware.RequirePermission("ai_custom_mcp_servers", "update"),
			h.Update)

		// Delete permissions
		customMcpServers.DELETE("/:id",
			permissionMiddleware.RequirePermission("ai_custom_mcp_servers", "delete"),
			h.Delete)

		// Test permissions
		customMcpServers.GET("/:id/test",
			permissionMiddleware.RequirePermission("ai_custom_mcp_servers", "read"),
			h.Test)
	}
}

// Create handles the create custom mcp server request.
func (h *customMcpServerHandler) Create(c *gin.Context) {
	var req *model.CustomMcpServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, err)
		return
	}

	customMcpServer := model.CustomMcpServer{
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		Headers:     stringutils.StringMapToJSON(req.Headers),
		Timeout:     req.Timeout,
		RetryCount:  req.RetryCount,
		Tags:        req.Tags,
	}

	createdCustomMcpServer, err := h.customMcpServerService.Create(c.Request.Context(), customMcpServer)
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.SuccessResponse(c, createdCustomMcpServer.ToResponse(), "Custom MCP server created successfully", http.StatusCreated)
}

// GetByID handles the get custom mcp server by id request.
func (h *customMcpServerHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	customMcpServer, err := h.customMcpServerService.GetByID(c.Request.Context(), id)

	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.SuccessResponse(c, customMcpServer.ToResponse(), "Custom MCP server retrieved successfully", http.StatusOK)
}

// List handles the list custom mcp servers request.
func (h *customMcpServerHandler) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	if skip := c.Query("skip"); skip != "" {
		skipVal, err := strconv.Atoi(skip)
		if err != nil {
			code, message, httpCode := errors.HandleError(err)
			response.ErrorResponse(c, code, message, nil, httpCode)
			return
		}

		limitVal, err := strconv.Atoi(c.DefaultQuery("limit", "100"))
		if err != nil {
			code, message, httpCode := errors.HandleError(err)
			response.ErrorResponse(c, code, message, nil, httpCode)
			return
		}

		page = (skipVal / limitVal) + 1
		pageSize = limitVal
	}

	search := c.DefaultQuery("search", "")
	tags := c.DefaultQuery("tags", "")

	var req model.CustomMcpServerListRequest
	req.Page = page
	req.PageSize = pageSize
	req.Search = search
	req.Tags = tags

	listCustomMcpServer, err := h.customMcpServerService.List(c.Request.Context(), req)

	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.PaginatedResponse(c, listCustomMcpServer.Items, listCustomMcpServer.Page, listCustomMcpServer.PageSize, int(listCustomMcpServer.TotalItems), "Custom MCP servers retrieved successfully", http.StatusOK)
}

// Update handles the update custom mcp server request.
func (h *customMcpServerHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	var req model.CustomMcpServerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, err)
		return
	}

	customMcpServer := &model.CustomMcpServer{
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		Headers:     stringutils.StringMapToJSON(req.Headers),
		Timeout:     req.Timeout,
		RetryCount:  req.RetryCount,
		Tags:        req.Tags,
	}

	updatedCustomMcpServer, err := h.customMcpServerService.Update(c.Request.Context(), customMcpServer, id)
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.SuccessResponse(c, updatedCustomMcpServer.ToResponse(), "Custom MCP server updated successfully", http.StatusOK)
}

// Delete handles the delete custom mcp server request.
func (h *customMcpServerHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	_, err = h.customMcpServerService.Delete(c.Request.Context(), id)
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.SuccessResponse(c, nil, "Custom MCP server deleted successfully", http.StatusNoContent)
}

// Test handles the test custom mcp server request.
func (h *customMcpServerHandler) Test(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	customMcpServer, err := h.customMcpServerService.Test(c.Request.Context(), id)
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.SuccessResponse(c, customMcpServer, "Custom MCP server test completed successfully", http.StatusOK)
}
