package service

import (
	"context"
	"evo-ai-core-service/internal/httpclient"
	errorsPostgres "evo-ai-core-service/internal/infra/postgres"
	"evo-ai-core-service/internal/utils/stringutils"
	model "evo-ai-core-service/pkg/custom_tool/model"
	repository "evo-ai-core-service/pkg/custom_tool/repository"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CustomToolService interface {
	Create(ctx context.Context, request model.CustomTool) (*model.CustomTool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.CustomTool, error)
	List(ctx context.Context, request model.CustomToolListRequest) (*model.CustomToolListResponse, error)
	Update(ctx context.Context, request *model.CustomTool, id uuid.UUID) (*model.CustomTool, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
	ConvertToHTTPTool(tool model.CustomToolResponse) map[string]interface{}
	Test(ctx context.Context, id uuid.UUID) (*model.CustomToolTestResponse, error)
}

type customToolService struct {
	customToolRepository repository.CustomToolRepository
}

func NewCustomToolService(customToolRepository repository.CustomToolRepository) CustomToolService {
	return &customToolService{
		customToolRepository: customToolRepository,
	}
}

func (s *customToolService) Create(ctx context.Context, request model.CustomTool) (*model.CustomTool, error) {
	customTool, err := s.customToolRepository.Create(ctx, request)

	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.CustomToolErrors)
	}

	return customTool, nil
}

func (s *customToolService) GetByID(ctx context.Context, id uuid.UUID) (*model.CustomTool, error) {
	customTool, err := s.customToolRepository.GetByID(ctx, id)

	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.CustomToolErrors)
	}

	return customTool, nil
}

func (s *customToolService) List(ctx context.Context, request model.CustomToolListRequest) (*model.CustomToolListResponse, error) {
	// Get paginated items
	customTools, err := s.customToolRepository.List(ctx, request)
	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.CustomToolErrors)
	}

	// Get total count
	totalItems, err := s.customToolRepository.Count(ctx, request)
	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.CustomToolErrors)
	}

	// Convert to response items
	items := make([]model.CustomToolResponse, len(customTools))
	for i, customTool := range customTools {
		items[i] = *customTool.ToResponse()
	}

	// Calculate pagination metadata
	totalPages := int((totalItems + int64(request.PageSize) - 1) / int64(request.PageSize))
	skip := (request.Page - 1) * request.PageSize
	limit := request.PageSize

	return &model.CustomToolListResponse{
		Items:      items,
		Page:       request.Page,
		PageSize:   request.PageSize,
		Skip:       skip,
		Limit:      limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

func (s *customToolService) Update(ctx context.Context, request *model.CustomTool, id uuid.UUID) (*model.CustomTool, error) {
	_, err := s.GetByID(ctx, id)

	if err != nil {
		return nil, err
	}

	customTool, err := s.customToolRepository.Update(ctx, request, id)

	if err != nil {
		return nil, errorsPostgres.MapDBError(err, model.CustomToolErrors)
	}

	return customTool, nil
}

func (s *customToolService) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := s.GetByID(ctx, id)

	if err != nil {
		return false, err
	}

	deleted, err := s.customToolRepository.Delete(ctx, id)

	if err != nil {
		return false, errorsPostgres.MapDBError(err, model.CustomToolErrors)
	}

	return deleted, nil
}

func (s *customToolService) ConvertToHTTPTool(tool model.CustomToolResponse) map[string]interface{} {
	var errorHandling map[string]interface{}
	if tool.ErrorHandling != nil {
		errorHandling = tool.ErrorHandling
	}

	if _, ok := errorHandling["timeout"]; !ok {
		errorHandling["timeout"] = 30
	}
	if _, ok := errorHandling["retry_count"]; !ok {
		errorHandling["retry_count"] = 0
	}
	if _, ok := errorHandling["fallback_response"]; !ok {
		errorHandling["fallback_response"] = map[string]string{
			"error":   "",
			"message": "",
		}
	}

	return map[string]interface{}{
		"name":     tool.Name,
		"method":   tool.Method,
		"endpoint": tool.Endpoint,
		"headers":  tool.Headers,
		"parameters": map[string]interface{}{
			"path_params":  tool.PathParams,
			"query_params": tool.QueryParams,
			"body_params":  tool.BodyParams,
		},
		"description":    tool.Description,
		"error_handling": errorHandling,
		"values":         tool.Values,
	}
}

func (s *customToolService) Test(ctx context.Context, id uuid.UUID) (*model.CustomToolTestResponse, error) {
	customTool, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := customTool.ToResponse()
	testResult := &model.TestResult{
		Success: false,
	}

	headers := stringutils.JSONToStringMap(customTool.Headers)
	start := time.Now()

	type TestResponse struct{}
	var httpErr error

	switch strings.ToUpper(customTool.Method) {
	case http.MethodGet:
		_, httpErr = httpclient.DoGetJSON[TestResponse](ctx,
			customTool.Endpoint,
			headers,
			http.StatusOK,
		)
	case http.MethodPost:
		_, httpErr = httpclient.DoPostJSON[TestResponse](ctx,
			customTool.Endpoint,
			nil,
			headers,
			http.StatusOK,
		)
	case http.MethodPut:
		_, httpErr = httpclient.DoPutJSON[TestResponse](ctx,
			customTool.Endpoint,
			nil,
			headers,
			http.StatusOK,
		)
	case http.MethodDelete:
		_, httpErr = httpclient.DoDeleteJSON[TestResponse](ctx,
			customTool.Endpoint,
			nil,
			headers,
			http.StatusOK,
		)
	default:
		testResult.Error = fmt.Sprintf("Unsupported method: %s", customTool.Method)
		return nil, fmt.Errorf("unsupported method: %s", customTool.Method)
	}

	elapsed := time.Since(start)

	if httpErr != nil {
		testResult.Error = httpErr.Error()
		if strings.Contains(httpErr.Error(), "connection") {
			testResult.Error = "Connection error"
		}
		if strings.Contains(httpErr.Error(), "context deadline exceeded") || strings.Contains(httpErr.Error(), "timeout") {
			testResult.Error = "Request timeout"
		}
		return &model.CustomToolTestResponse{
			Tool:       response,
			TestResult: testResult,
		}, nil
	}

	testResult.Success = true
	testResult.StatusCode = http.StatusOK
	testResult.ResponseTime = elapsed.Seconds()

	return &model.CustomToolTestResponse{
		Tool:       response,
		TestResult: testResult,
	}, nil
}
