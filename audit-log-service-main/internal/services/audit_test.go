package services

import (
	"context"
	"testing"
	"time"

	"audit-log-service/internal/config"
	"audit-log-service/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuditRepository is a mock implementation of AuditRepository
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, auditLog *models.AuditLog) error {
	args := m.Called(ctx, auditLog)
	return args.Error(0)
}

func (m *MockAuditRepository) GetByID(ctx context.Context, id string) (*models.AuditLog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) List(ctx context.Context, filter *models.AuditLogFilter) (*models.PaginatedResponse, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*models.PaginatedResponse), args.Error(1)
}

func (m *MockAuditRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAuditRepository) GetStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, startDate, endDate)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuditRepository) UpdateStatus(ctx context.Context, id string, req *models.UpdateStatusRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

// MockNotificationService is a mock implementation of NotificationService
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNotification(ctx context.Context, auditLog *models.AuditLog) error {
	args := m.Called(ctx, auditLog)
	return args.Error(0)
}

func TestAuditService_CreateAuditLog(t *testing.T) {
	// Setup
	mockRepo := new(MockAuditRepository)
	mockNotification := new(MockNotificationService)
	logger := logrus.New()
	
	// Create test config
	testConfig := &config.Config{
		Audit: config.AuditConfig{
			Enabled:       true,
			DefaultStatus: "pending",
			StatusValues:  []string{"pending", "processing", "completed", "failed", "archived"},
			ErrorStatus:   "failed",
		},
	}

	service := NewAuditService(mockRepo, mockNotification, testConfig, logger)

	// Test data
	req := &models.CreateAuditLogRequest{
		TenantID:    "tenant-123",
		UserID:      "user-456",
		Resource:    "users",
		Event:       "USER_CREATED",
		Method:      "POST",
		IP:          "192.168.1.100",
		Environment: "production",
		Data:        []byte(`{"user_id": "123", "email": "test@example.com"}`),
		Meta:        map[string]interface{}{"request_id": "req-789"},
	}

	// Expected audit log
	expectedAuditLog := &models.AuditLog{
		ID:          "generated-uuid",
		TenantID:    req.TenantID,
		UserID:      req.UserID,
		Resource:    req.Resource,
		Event:       req.Event,
		Method:      req.Method,
		IP:          req.IP,
		Environment: req.Environment,
		Data:        req.Data,
		Meta:        req.Meta,
		Timestamp:   time.Now(),
	}

	// Mock expectations
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.AuditLog")).Return(nil).Run(func(args mock.Arguments) {
		auditLog := args.Get(1).(*models.AuditLog)
		auditLog.ID = "generated-uuid" // Simulate ID generation
	})
	mockNotification.On("SendNotification", mock.Anything, mock.AnythingOfType("*models.AuditLog")).Return(nil)

	// Execute
	ctx := context.Background()
	result, err := service.CreateAuditLog(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedAuditLog.TenantID, result.TenantID)
	assert.Equal(t, expectedAuditLog.UserID, result.UserID)
	assert.Equal(t, expectedAuditLog.Resource, result.Resource)
	assert.Equal(t, expectedAuditLog.Event, result.Event)
	assert.Equal(t, expectedAuditLog.Method, result.Method)
	assert.NotEmpty(t, result.ID)
	assert.NotZero(t, result.Timestamp)

	// Verify mock calls
	mockRepo.AssertExpectations(t)
	mockNotification.AssertExpectations(t)
}

func TestAuditService_GetAuditLogByID(t *testing.T) {
	// Setup
	mockRepo := new(MockAuditRepository)
	mockNotification := new(MockNotificationService)
	logger := logrus.New()

	service := NewAuditService(mockRepo, mockNotification, logger)

	// Test data
	auditLogID := uuid.New().String()
	expectedAuditLog := &models.AuditLog{
		ID:          auditLogID,
		TenantID:    "tenant-123",
		UserID:      "user-456",
		Resource:    "users",
		Event:       "USER_CREATED",
		Method:      "POST",
		IP:          "192.168.1.100",
		Environment: "production",
		Data:        []byte(`{"user_id": "123"}`),
		Meta:        []byte(`{"request_id": "req-789"}`),
		Timestamp:   time.Now(),
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, auditLogID).Return(expectedAuditLog, nil)

	// Execute
	ctx := context.Background()
	result, err := service.GetAuditLogByID(ctx, auditLogID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedAuditLog, result)

	// Verify mock calls
	mockRepo.AssertExpectations(t)
}

func TestAuditService_ListAuditLogs(t *testing.T) {
	// Setup
	mockRepo := new(MockAuditRepository)
	mockNotification := new(MockNotificationService)
	logger := logrus.New()

	service := NewAuditService(mockRepo, mockNotification, logger)

	// Test data
	filter := &models.AuditLogFilter{
		TenantID: "tenant-123",
		Limit:    10,
		Offset:   0,
	}

	expectedResponse := &models.PaginatedResponse{
		Data: []models.AuditLog{
			{
				ID:          uuid.New().String(),
				TenantID:    "tenant-123",
				UserID:      "user-456",
				Resource:    "users",
				Event:       "USER_CREATED",
				Method:      "POST",
				IP:          "192.168.1.100",
				Environment: "production",
				Timestamp:   time.Now(),
			},
		},
		Total:  1,
		Limit:  10,
		Offset: 0,
	}

	// Mock expectations
	mockRepo.On("List", mock.Anything, filter).Return(expectedResponse, nil)

	// Execute
	ctx := context.Background()
	result, err := service.ListAuditLogs(ctx, filter)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)

	// Verify mock calls
	mockRepo.AssertExpectations(t)
}

func TestAuditService_DeleteAuditLog(t *testing.T) {
	// Setup
	mockRepo := new(MockAuditRepository)
	mockNotification := new(MockNotificationService)
	logger := logrus.New()

	service := NewAuditService(mockRepo, mockNotification, logger)

	// Test data
	auditLogID := uuid.New().String()

	// Mock expectations
	mockRepo.On("Delete", mock.Anything, auditLogID).Return(nil)

	// Execute
	ctx := context.Background()
	err := service.DeleteAuditLog(ctx, auditLogID)

	// Assert
	assert.NoError(t, err)

	// Verify mock calls
	mockRepo.AssertExpectations(t)
}

func TestAuditService_GetAuditLogStats(t *testing.T) {
	// Setup
	mockRepo := new(MockAuditRepository)
	mockNotification := new(MockNotificationService)
	logger := logrus.New()

	service := NewAuditService(mockRepo, mockNotification, logger)

	// Test data
	tenantID := "tenant-123"
	startDate := time.Now().AddDate(0, -1, 0) // 1 month ago
	endDate := time.Now()

	expectedStats := map[string]interface{}{
		"total_events":       100,
		"events_by_type":     map[string]int{"USER_CREATED": 50, "USER_UPDATED": 30, "USER_DELETED": 20},
		"events_by_method":   map[string]int{"POST": 50, "PUT": 30, "DELETE": 20},
		"events_by_resource": map[string]int{"users": 80, "roles": 20},
	}

	// Mock expectations
	mockRepo.On("GetStats", mock.Anything, tenantID, startDate, endDate).Return(expectedStats, nil)

	// Execute
	ctx := context.Background()
	result, err := service.GetAuditLogStats(ctx, tenantID, startDate, endDate)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedStats, result)

	// Verify mock calls
	mockRepo.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkAuditService_CreateAuditLog(b *testing.B) {
	// Setup
	mockRepo := new(MockAuditRepository)
	mockNotification := new(MockNotificationService)
	logger := logrus.New()

	service := NewAuditService(mockRepo, mockNotification, logger)

	// Mock expectations
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.AuditLog")).Return(nil)
	mockNotification.On("SendNotification", mock.Anything, mock.AnythingOfType("*models.AuditLog")).Return(nil)

	req := &models.CreateAuditLogRequest{
		TenantID:    "tenant-123",
		UserID:      "user-456",
		Resource:    "users",
		Event:       "USER_CREATED",
		Method:      "POST",
		IP:          "192.168.1.100",
		Environment: "production",
		Data:        []byte(`{"user_id": "123"}`),
		Meta:        []byte(`{"request_id": "req-789"}`),
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateAuditLog(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
