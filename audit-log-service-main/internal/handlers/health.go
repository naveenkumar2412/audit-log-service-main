package handlers

import (
	"context"
	"net/http"
	"time"

	"audit-log-service/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.DB, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks"`
}

// HealthCheck performs a comprehensive health check
// @Summary Health check
// @Description Check the health of the service and its dependencies
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0", // This could be injected from build info
		Checks:    make(map[string]string),
	}

	// Check database health
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.db.Health(ctx); err != nil {
		h.logger.WithError(err).Error("Database health check failed")
		response.Status = "unhealthy"
		response.Checks["database"] = "unhealthy: " + err.Error()
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response.Checks["database"] = "healthy"
	c.JSON(http.StatusOK, response)
}

// ReadinessCheck checks if the service is ready to serve requests
// @Summary Readiness check
// @Description Check if the service is ready to serve requests
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Check database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.db.Health(ctx); err != nil {
		h.logger.WithError(err).Error("Readiness check failed")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// LivenessCheck checks if the service is alive
// @Summary Liveness check
// @Description Check if the service is alive
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /live [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
