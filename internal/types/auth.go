package types

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EvoAuthUser represents the user data from /api/v1/me response
type EvoAuthUser struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	DisplayName  *string   `json:"display_name"`
	Availability string    `json:"availability"`
	MFAEnabled   bool      `json:"mfa_enabled"`
	Confirmed    bool      `json:"confirmed"`
	Type         string    `json:"type"`
	Role         *Role     `json:"role"`
}

type EvoAuthFeature struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type EvoAuthPlan struct {
	ID       uuid.UUID        `json:"id"`
	PlanName string           `json:"name"`
	IsActive bool             `json:"is_active"`
	IsCustom bool             `json:"is_custom"`
	StartsAt string           `json:"starts_at"`
	EndsAt   string           `json:"ends_at"`
	Features []EvoAuthFeature `json:"features"`
}

// EvoAuthAccount represents an account from /api/v1/me response
type EvoAuthAccount struct {
	ID         uuid.UUID        `json:"id"`
	Name       string           `json:"name"`
	Status     string           `json:"status"`
	Locale     string           `json:"locale"`
	CreatedAt  string           `json:"created_at"`
	UpdatedAt  string           `json:"updated_at"`
	Features   json.RawMessage  `json:"features"`
	ActivePlan *EvoAuthPlan     `json:"active_plan,omitempty"`
}

// EvoAuthValidateToken represents the complete response from /api/v1/me
type EvoAuthValidateToken struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

type EvoAuthValidateTokenData struct {
	User     EvoAuthUser      `json:"user"`
	Accounts []EvoAuthAccount `json:"accounts"`
}

// PermissionResponse representa a resposta da API de validação de permissão
type PermissionResponse struct {
	HasPermission bool   `json:"has_permission"`
	PermissionKey string `json:"permission_key"`
	Message       string `json:"message,omitempty"`
}

// PermissionRequest representa a requisição para validação de permissão via POST
type PermissionRequest struct {
	PermissionKey string `json:"permission_key" binding:"required"`
}

// PermissionChecker interface para validação de permissões
type PermissionChecker interface {
	RequirePermission(resource, action string) gin.HandlerFunc
	CheckPermission(authToken, permissionKey string) (bool, error)
}

type Role struct {
	ID   uuid.UUID `json:"id"`
	Key  string    `json:"key"`
	Name string    `json:"name"`
}
