package handler

import (
	"evo-ai-core-service/internal/httpclient/errors"
	"evo-ai-core-service/internal/httpclient/response"
	"evo-ai-core-service/internal/middleware"
	"evo-ai-core-service/pkg/folder/model"
	"evo-ai-core-service/pkg/folder/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FolderHandler interface defines the contract for folder handlers
type FolderHandler interface {
	RegisterRoutesMiddleware(router gin.IRouter)
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

// folderHandler implements the FolderHandler interface
type folderHandler struct {
	folderService service.FolderService
}

// NewFolderHandler creates a new folder handler
func NewFolderHandler(folderService service.FolderService) FolderHandler {
	return &folderHandler{
		folderService: folderService,
	}
}

// RegisterRoutesMiddleware registers the routes for the folder handler with middleware
func (h *folderHandler) RegisterRoutesMiddleware(router gin.IRouter) {
	// Usar middleware global de permissão
	permissionMiddleware := middleware.GetGlobalPermissionMiddleware()
	folders := router.Group("/agents/folders")
	{
		// Read permissions
		folders.GET("",
			permissionMiddleware.RequirePermission("ai_folders", "read"),
			h.List)
		folders.GET("/",
			permissionMiddleware.RequirePermission("ai_folders", "read"),
			h.List)
		folders.GET("/:id",
			permissionMiddleware.RequirePermission("ai_folders", "read"),
			h.GetByID)

		// Create permissions
		folders.POST("",
			permissionMiddleware.RequirePermission("ai_folders", "create"),
			h.Create)
		folders.POST("/",
			permissionMiddleware.RequirePermission("ai_folders", "create"),
			h.Create)

		// Update permissions
		folders.PUT("/:id",
			permissionMiddleware.RequirePermission("ai_folders", "update"),
			h.Update)

		// Delete permissions
		folders.DELETE("/:id",
			permissionMiddleware.RequirePermission("ai_folders", "delete"),
			h.Delete)
	}

	// Endpoint específico que o frontend espera
	router.GET("/agents/accessible-folders",
		permissionMiddleware.RequirePermission("ai_folders", "access_shared"),
		h.List)
	router.GET("/agents/accessible-folders/", h.List) // Com barra também
}

// Create handles the create folder request
func (h *folderHandler) Create(c *gin.Context) {
	var req *model.FolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, err)
		return
	}

	folder := model.Folder{
		Name:        req.Name,
		Description: req.Description,
	}

	createdFolder, err := h.folderService.Create(c.Request.Context(), folder)
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.SuccessResponse(c, createdFolder.ToResponse(), "Folder created successfully", http.StatusCreated)
}

// GetByID handles the get folder by id request
func (h *folderHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	folder, err := h.folderService.GetByID(c.Request.Context(), id)

	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.SuccessResponse(c, folder.ToResponse(), "Folder retrieved successfully", http.StatusOK)
}

// List handles the list folders request
func (h *folderHandler) List(c *gin.Context) {
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

	listFolders, err := h.folderService.List(c.Request.Context(), page, pageSize)

	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.PaginatedResponse(c, listFolders.Items, listFolders.Page, listFolders.PageSize, int(listFolders.TotalItems), "Folders retrieved successfully", http.StatusOK)
}

// Update handles the update folder request
func (h *folderHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	var req model.FolderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, err)
		return
	}

	folder := &model.Folder{
		Name:        req.Name,
		Description: req.Description,
	}

	updatedFolder, err := h.folderService.Update(c.Request.Context(), folder, id)
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.SuccessResponse(c, updatedFolder.ToResponse(), "Folder updated successfully", http.StatusOK)
}

// Delete handles the delete folder request
func (h *folderHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	_, err = h.folderService.Delete(c.Request.Context(), id)
	if err != nil {
		code, message, httpCode := errors.HandleError(err)
		response.ErrorResponse(c, code, message, nil, httpCode)
		return
	}

	response.SuccessResponse(c, nil, "Folder deleted successfully", http.StatusNoContent)
}
