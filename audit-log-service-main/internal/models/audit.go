package models

import (
	"encoding/json"
	"time"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id" validate:"required"`
	UserID      string                 `json:"user_id" db:"user_id" validate:"required"`
	Resource    string                 `json:"resource" db:"resource" validate:"required"`
	Event       string                 `json:"event" db:"event" validate:"required"`
	Method      string                 `json:"method" db:"method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	IP          string                 `json:"ip" db:"ip" validate:"required,ip"`
	Status      string                 `json:"status" db:"status" validate:"required"`
	Data        json.RawMessage        `json:"data" db:"data"`
	Environment string                 `json:"environment" db:"environment" validate:"required,oneof=development staging production"`
	Meta        map[string]interface{} `json:"meta" db:"meta"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// CreateAuditLogRequest represents the request payload for creating an audit log
type CreateAuditLogRequest struct {
	TenantID    string                 `json:"tenant_id" validate:"required"`
	UserID      string                 `json:"user_id" validate:"required"`
	Resource    string                 `json:"resource" validate:"required"`
	Event       string                 `json:"event" validate:"required"`
	Method      string                 `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	IP          string                 `json:"ip" validate:"required,ip"`
	Status      string                 `json:"status,omitempty"`
	Data        json.RawMessage        `json:"data"`
	Environment string                 `json:"environment" validate:"required,oneof=development staging production"`
	Meta        map[string]interface{} `json:"meta"`
}

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	TenantID    string    `json:"tenant_id" form:"tenant_id"`
	UserID      string    `json:"user_id" form:"user_id"`
	Resource    string    `json:"resource" form:"resource"`
	Event       string    `json:"event" form:"event"`
	Method      string    `json:"method" form:"method"`
	Status      string    `json:"status" form:"status"`
	Environment string    `json:"environment" form:"environment"`
	StartDate   time.Time `json:"start_date" form:"start_date"`
	EndDate     time.Time `json:"end_date" form:"end_date"`
	Limit       int       `json:"limit" form:"limit"`
	Offset      int       `json:"offset" form:"offset"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data    []AuditLog `json:"data"`
	Total   int64      `json:"total"`
	Limit   int        `json:"limit"`
	Offset  int        `json:"offset"`
	HasMore bool       `json:"has_more"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// UpdateStatusRequest represents the request payload for updating audit log status
type UpdateStatusRequest struct {
	Status string                 `json:"status"`
	Data   json.RawMessage        `json:"data,omitempty"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}