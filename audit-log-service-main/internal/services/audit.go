package services

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"audit-log-service/internal/config"
	"audit-log-service/internal/models"
	"audit-log-service/internal/repositories"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// AuditService handles business logic for audit logs
type AuditService struct {
	repo                *repositories.AuditRepository
	notificationService *NotificationService
	config              *config.Config
	validator           *validator.Validate
	logger              *logrus.Logger
}

// NewAuditService creates a new audit service
func NewAuditService(
	repo *repositories.AuditRepository,
	notificationService *NotificationService,
	config *config.Config,
	logger *logrus.Logger,
) *AuditService {
	return &AuditService{
		repo:                repo,
		notificationService: notificationService,
		config:              config,
		validator:           validator.New(),
		logger:              logger,
	}
}

// CreateAuditLog creates a new audit log entry
func (s *AuditService) CreateAuditLog(ctx context.Context, req *models.CreateAuditLogRequest) (*models.AuditLog, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		s.logger.WithError(err).Error("Validation failed for audit log creation")
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate IP address
	if err := s.validateIP(req.IP); err != nil {
		s.logger.WithError(err).WithField("ip", req.IP).Error("Invalid IP address")
		return nil, fmt.Errorf("invalid IP address: %w", err)
	}

	// Handle status - use provided status or default
	status := req.Status
	if status == "" {
		status = s.config.Audit.DefaultStatus
	}

	// Validate status if audit is enabled
	if s.config.Audit.Enabled && !s.config.Audit.IsValidStatus(status) {
		s.logger.WithField("status", status).Error("Invalid status value")
		return nil, fmt.Errorf("invalid status '%s', must be one of: %s",
			status, strings.Join(s.config.Audit.StatusValues, ", "))
	}

	// Create audit log model
	auditLog := &models.AuditLog{
		TenantID:    req.TenantID,
		UserID:      req.UserID,
		Resource:    req.Resource,
		Event:       req.Event,
		Method:      strings.ToUpper(req.Method),
		IP:          req.IP,
		Data:        req.Data,
		Environment: req.Environment,
		Meta:        req.Meta,
		Status:      status,
	}

	// Save to database
	if err := s.repo.Create(ctx, auditLog); err != nil {
		s.logger.WithError(err).Error("Failed to create audit log")
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"audit_log_id": auditLog.ID,
		"tenant_id":    auditLog.TenantID,
		"user_id":      auditLog.UserID,
		"resource":     auditLog.Resource,
		"event":        auditLog.Event,
	}).Info("Audit log created successfully")

	// Send notifications if configured
	if s.notificationService != nil {
		go func() {
			if err := s.notificationService.SendNotification(context.Background(), auditLog); err != nil {
				s.logger.WithError(err).Error("Failed to send notification")
			}
		}()
	}

	return auditLog, nil
}

// GetAuditLogByID retrieves an audit log by ID
func (s *AuditService) GetAuditLogByID(ctx context.Context, id string) (*models.AuditLog, error) {
	if id == "" {
		return nil, fmt.Errorf("audit log ID is required")
	}

	auditLog, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get audit log")
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	return auditLog, nil
}

// ListAuditLogs retrieves audit logs with filtering and pagination
func (s *AuditService) ListAuditLogs(ctx context.Context, filter *models.AuditLogFilter) (*models.PaginatedResponse, error) {
	// Set default pagination if not provided
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000 // Maximum limit to prevent excessive queries
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Validate date range
	if !filter.StartDate.IsZero() && !filter.EndDate.IsZero() {
		if filter.StartDate.After(filter.EndDate) {
			return nil, fmt.Errorf("start date cannot be after end date")
		}
	}

	result, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list audit logs")
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"total":     result.Total,
		"limit":     filter.Limit,
		"offset":    filter.Offset,
		"tenant_id": filter.TenantID,
	}).Info("Audit logs retrieved successfully")

	return result, nil
}

// DeleteAuditLog deletes an audit log by ID
func (s *AuditService) DeleteAuditLog(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("audit log ID is required")
	}

	// First, get the audit log to ensure it exists and for logging
	auditLog, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("audit log not found: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to delete audit log")
		return fmt.Errorf("failed to delete audit log: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"audit_log_id": id,
		"tenant_id":    auditLog.TenantID,
		"resource":     auditLog.Resource,
	}).Info("Audit log deleted successfully")

	return nil
}

// GetAuditLogStats retrieves statistics for audit logs
func (s *AuditService) GetAuditLogStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	filter := &models.AuditLogFilter{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     1, // We only need the count
		Offset:    0,
	}

	errorFilter := &models.AuditLogFilter{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    s.config.Audit.ErrorStatus, // Assuming this is the status for errors
		Limit:     1, // We only need the count
		Offset:    0,
	}

	result, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log stats: %w", err)
	}

	errorResult, err := s.repo.List(ctx, errorFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_logs": result.Total,
		"error_count": errorResult.Total,
		"period": map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		},
		"tenant_id": tenantID,
	}

	return stats, nil
}

// UpdateAuditLogStatus updates the status of an audit log entry
func (s *AuditService) UpdateAuditLogStatus(ctx context.Context, id string, req *models.UpdateStatusRequest) error {
	if id == "" {
		return fmt.Errorf("audit log ID is required")
	}

	if req.Status == "" {
		return fmt.Errorf("status is required")
	}

	status := req.Status 

	// Validate status if audit is enabled
	if s.config.Audit.Enabled && !s.config.Audit.IsValidStatus(status) {
		s.logger.WithField("status", status).Error("Invalid status value")
		return fmt.Errorf("invalid status '%s', must be one of: %s",
			status, strings.Join(s.config.Audit.StatusValues, ", "))
	}

	// Update status in database
	if err := s.repo.UpdateStatus(ctx, id, req); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"audit_log_id": id,
			"status":       status,
		}).Error("Failed to update audit log status")
		return fmt.Errorf("failed to update audit log status: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"audit_log_id": id,
		"status":       status,
		"has_data":     req.Data != nil,
		"has_meta":     req.Meta != nil,
	}).Info("Audit log status updated successfully")

	return nil
}

// validateIP validates if the provided string is a valid IP address
func (s *AuditService) validateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format")
	}
	return nil
}
