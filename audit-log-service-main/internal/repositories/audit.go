package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"audit-log-service/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditRepository handles audit log data operations
type AuditRepository struct {
	db *pgxpool.Pool
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, auditLog *models.AuditLog) error {
	auditLog.ID = uuid.New().String()
	auditLog.Timestamp = time.Now()
	auditLog.CreatedAt = time.Now()
	auditLog.UpdatedAt = time.Now()

	query := `
		INSERT INTO audit_logs (id, tenant_id, user_id, resource, event, method, ip, status, data, environment, meta, timestamp, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	metaJSON, err := json.Marshal(auditLog.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		auditLog.ID,
		auditLog.TenantID,
		auditLog.UserID,
		auditLog.Resource,
		auditLog.Event,
		auditLog.Method,
		auditLog.IP,
		auditLog.Status,
		auditLog.Data,
		auditLog.Environment,
		metaJSON,
		auditLog.Timestamp,
		auditLog.CreatedAt,
		auditLog.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetByID retrieves an audit log by ID
func (r *AuditRepository) GetByID(ctx context.Context, id string) (*models.AuditLog, error) {
	query := `
		SELECT id, tenant_id, user_id, resource, event, method, ip::text, status, data, environment, meta, timestamp, created_at, updated_at
		FROM audit_logs
		WHERE id = $1
	`

	var auditLog models.AuditLog
	var metaJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&auditLog.ID,
		&auditLog.TenantID,
		&auditLog.UserID,
		&auditLog.Resource,
		&auditLog.Event,
		&auditLog.Method,
		&auditLog.IP,
		&auditLog.Status,
		&auditLog.Data,
		&auditLog.Environment,
		&metaJSON,
		&auditLog.Timestamp,
		&auditLog.CreatedAt,
		&auditLog.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("audit log not found")
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	if len(metaJSON) > 0 {
		err = json.Unmarshal(metaJSON, &auditLog.Meta)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}
	}

	return &auditLog, nil
}

// List retrieves audit logs with filtering and pagination
func (r *AuditRepository) List(ctx context.Context, filter *models.AuditLogFilter) (*models.PaginatedResponse, error) {
	whereClause, args := r.buildWhereClause(filter)

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Build main query with pagination
	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, resource, event, method, ip::text, status, data, environment, meta, timestamp, created_at, updated_at
		FROM audit_logs
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var auditLogs []models.AuditLog
	for rows.Next() {
		var auditLog models.AuditLog
		var metaJSON []byte

		err := rows.Scan(
			&auditLog.ID,
			&auditLog.TenantID,
			&auditLog.UserID,
			&auditLog.Resource,
			&auditLog.Event,
			&auditLog.Method,
			&auditLog.IP,
			&auditLog.Status,
			&auditLog.Data,
			&auditLog.Environment,
			&metaJSON,
			&auditLog.Timestamp,
			&auditLog.CreatedAt,
			&auditLog.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if len(metaJSON) > 0 {
			err = json.Unmarshal(metaJSON, &auditLog.Meta)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
			}
		}

		auditLogs = append(auditLogs, auditLog)
	}

	hasMore := int64(filter.Offset+filter.Limit) < total

	return &models.PaginatedResponse{
		Data:    auditLogs,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: hasMore,
	}, nil
}

// Delete removes an audit log by ID
func (r *AuditRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM audit_logs WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete audit log: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("audit log not found")
	}

	return nil
}

// buildWhereClause builds WHERE clause for filtering
func (r *AuditRepository) buildWhereClause(filter *models.AuditLogFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.TenantID != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, filter.TenantID)
		argIndex++
	}

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, filter.UserID)
		argIndex++
	}

	if filter.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argIndex))
		args = append(args, filter.Resource)
		argIndex++
	}

	if filter.Event != "" {
		conditions = append(conditions, fmt.Sprintf("event = $%d", argIndex))
		args = append(args, filter.Event)
		argIndex++
	}

	if filter.Method != "" {
		conditions = append(conditions, fmt.Sprintf("method = $%d", argIndex))
		args = append(args, filter.Method)
		argIndex++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.Environment != "" {
		conditions = append(conditions, fmt.Sprintf("environment = $%d", argIndex))
		args = append(args, filter.Environment)
		argIndex++
	}

	if !filter.StartDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, filter.StartDate)
		argIndex++
	}

	if !filter.EndDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIndex))
		args = append(args, filter.EndDate)
		argIndex++
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

// UpdateStatus updates the status of an audit log entry
func (r *AuditRepository) UpdateStatus(ctx context.Context, id string, req *models.UpdateStatusRequest) error {
	// Build the query dynamically based on what fields are being updated
	setParts := []string{"status = $1", "updated_at = NOW()"}
	args := []interface{}{req.Status}
	argIndex := 2

	// Add data update if provided
	if req.Data != nil {
		setParts = append(setParts, fmt.Sprintf("data = $%d", argIndex))
		args = append(args, req.Data)
		argIndex++
	}

	// Add meta update if provided
	if req.Meta != nil {
		metaJSON, err := json.Marshal(req.Meta)
		if err != nil {
			return fmt.Errorf("failed to marshal meta: %w", err)
		}
		setParts = append(setParts, fmt.Sprintf("meta = $%d", argIndex))
		args = append(args, metaJSON)
		argIndex++
	}

	query := fmt.Sprintf(`
		UPDATE audit_logs 
		SET %s
		WHERE id = $%d
	`, strings.Join(setParts, ", "), argIndex)

	args = append(args, id)

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update audit log status: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("audit log not found")
	}

	return nil
}
