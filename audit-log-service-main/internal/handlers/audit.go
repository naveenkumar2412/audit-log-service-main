package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"audit-log-service/internal/models"
	"audit-log-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuditHandler handles HTTP requests for audit logs
type AuditHandler struct {
	auditService *services.AuditService
	logger       *logrus.Logger
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService, logger *logrus.Logger) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		logger:       logger,
	}
}

// CreateAuditLog creates a new audit log entry
// @Summary Create audit log
// @Description Create a new audit log entry
// @Tags audit
// @Accept json
// @Produce json
// @Param audit body models.CreateAuditLogRequest true "Audit log data"
// @Success 201 {object} models.AuditLog
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/audit [post]
func (h *AuditHandler) CreateAuditLog(c *gin.Context) {
	var req models.CreateAuditLogRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get client IP if not provided
	if req.IP == "" {
		req.IP = c.ClientIP()
	}

	auditLog, err := h.auditService.CreateAuditLog(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create audit log")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create audit log",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, auditLog)
}

// GetAuditLog retrieves an audit log by ID
// @Summary Get audit log by ID
// @Description Get a specific audit log by its ID
// @Tags audit
// @Produce json
// @Param id path string true "Audit log ID"
// @Success 200 {object} models.AuditLog
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/audit/{id} [get]
func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing audit log ID",
			Message: "Audit log ID is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	auditLog, err := h.auditService.GetAuditLogByID(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to get audit log")
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Audit log not found",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, auditLog)
}

// ListAuditLogs retrieves audit logs with filtering and pagination
// @Summary List audit logs
// @Description Get a list of audit logs with optional filtering and pagination
// @Tags audit
// @Produce json
// @Param tenant_id query string false "Filter by tenant ID"
// @Param user_id query string false "Filter by user ID"
// @Param resource query string false "Filter by resource"
// @Param event query string false "Filter by event"
// @Param method query string false "Filter by HTTP method"
// @Param status query string false "Filter by status"
// @Param environment query string false "Filter by environment"
// @Param start_date query string false "Filter by start date (RFC3339 format)"
// @Param end_date query string false "Filter by end date (RFC3339 format)"
// @Param limit query int false "Number of results to return (default: 50, max: 1000)"
// @Param offset query int false "Number of results to skip (default: 0)"
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/audit [get]
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	filter := &models.AuditLogFilter{
		TenantID:    c.Query("tenant_id"),
		UserID:      c.Query("user_id"),
		Resource:    c.Query("resource"),
		Event:       c.Query("event"),
		Method:      c.Query("method"),
		Status:      c.Query("status"),
		Environment: c.Query("environment"),
	}

	// Parse dates
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = startDate
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid start_date format",
				Message: "start_date must be in RFC3339 format",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = endDate
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid end_date format",
				Message: "end_date must be in RFC3339 format",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid limit format",
				Message: "limit must be a valid integer",
				Code:    http.StatusBadRequest,
			})
			return
		}
	} else {
		filter.Limit = 50 // Default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid offset format",
				Message: "offset must be a valid integer",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}

	result, err := h.auditService.ListAuditLogs(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list audit logs")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve audit logs",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteAuditLog deletes an audit log by ID
// @Summary Delete audit log
// @Description Delete a specific audit log by its ID
// @Tags audit
// @Param id path string true "Audit log ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/audit/{id} [delete]
func (h *AuditHandler) DeleteAuditLog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing audit log ID",
			Message: "Audit log ID is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err := h.auditService.DeleteAuditLog(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("id", id).Error("Failed to delete audit log")
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Failed to delete audit log",
			Message: err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAuditLogStats retrieves audit log statistics
// @Summary Get audit log statistics
// @Description Get statistics for audit logs within a date range
// @Tags audit
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Param start_date query string true "Start date (RFC3339 format)"
// @Param end_date query string true "End date (RFC3339 format)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/audit/stats [get]
func (h *AuditHandler) GetAuditLogStats(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing tenant_id",
			Message: "tenant_id is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing date parameters",
			Message: "Both start_date and end_date are required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid start_date format",
			Message: "start_date must be in RFC3339 format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid end_date format",
			Message: "end_date must be in RFC3339 format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	stats, err := h.auditService.GetAuditLogStats(c.Request.Context(), tenantID, startDate, endDate)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get audit log stats")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to retrieve statistics",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateAuditLogStatusRequest represents the request payload for updating audit log status
type UpdateAuditLogStatusRequest struct {
	Status string `json:"status" validate:"required"`
	Data   json.RawMessage       `json:"data,omitempty"` // Optional data for the status update
	Meta   map[string]interface{} `json:"meta,omitempty"` // Optional metadata for the status update
}

// UpdateAuditLogStatus updates the status of an audit log entry
// @Summary Update audit log status
// @Description Update the status of a specific audit log entry
// @Tags audit
// @Accept json
// @Produce json
// @Param id path string true "Audit log ID"
// @Param status body UpdateAuditLogStatusRequest true "New status"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/audit/{id}/status [put]
func (h *AuditHandler) UpdateAuditLogStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: "Audit log ID is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req UpdateAuditLogStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	updateReq := &models.UpdateStatusRequest{
		Status: req.Status,
		Data:   req.Data,
		Meta:   req.Meta,
	}
	err := h.auditService.UpdateAuditLogStatus(c.Request.Context(), id, updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Audit log not found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		if strings.Contains(err.Error(), "invalid status") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid status",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		h.logger.WithError(err).Error("Failed to update audit log status")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update audit log status",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"message": "Audit log status updated successfully",
		"id":      id,
		"status":  req.Status,
	})
}
