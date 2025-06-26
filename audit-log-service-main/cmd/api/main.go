package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"audit-log-service/internal/config"
	"audit-log-service/internal/database"
	"audit-log-service/internal/handlers"
	"audit-log-service/internal/middleware"
	"audit-log-service/internal/notifications"
	"audit-log-service/internal/repositories"
	"audit-log-service/internal/services"
	"audit-log-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewLogger(&cfg.Logging)
	log.Info("Starting Audit Log Service")

	// Initialize database
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	log.Info("Database connection established")

	// Initialize repositories
	auditRepo := repositories.NewAuditRepository(db.Pool)

	// Initialize notification senders
	emailSender := notifications.NewSMTPEmailSender(&cfg.Notification.Email)
	slackSender := notifications.NewWebhookSlackSender(&cfg.Notification.Slack)
	webhookSender := notifications.NewHTTPWebhookSender(&cfg.Notification.Webhook)

	// Initialize services
	notificationService := services.NewNotificationService(
		&cfg.Notification,
		emailSender,
		slackSender,
		webhookSender,
		log.Logger,
	)

	auditService := services.NewAuditService(
		auditRepo,
		notificationService,
		cfg,
		log.Logger,
	)

	// Initialize handlers
	auditHandler := handlers.NewAuditHandler(auditService, log.Logger)
	healthHandler := handlers.NewHealthHandler(db, log.Logger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(&cfg.Auth, log.Logger)
	loggingMiddleware := middleware.NewLoggingMiddleware(log.Logger)

	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.New()

	// Add middleware
	router.Use(loggingMiddleware.RequestLogger())
	router.Use(loggingMiddleware.ErrorLogger())
	router.Use(loggingMiddleware.RecoveryLogger())

	// CORS middleware - manual implementation for Gin
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check routes (no auth required)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)
	router.GET("/live", healthHandler.LivenessCheck)

	// API routes with authentication
	api := router.Group("/api/v1")
	api.Use(authMiddleware.OptionalAuth()) // Allow both JWT and API key auth

	// Audit log routes
	audit := api.Group("/audit")
	{
		audit.POST("", auditHandler.CreateAuditLog)
		audit.GET("", auditHandler.ListAuditLogs)
		audit.GET("/:id", auditHandler.GetAuditLog)
		audit.PUT("/:id/status", auditHandler.UpdateAuditLogStatus)
		audit.DELETE("/:id", auditHandler.DeleteAuditLog)
		audit.GET("/stats", auditHandler.GetAuditLogStats)
	}

	// Create HTTP server
	serverAddr := cfg.Server.GetAddress()
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.WithField("address", serverAddr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	log.Info("Server stopped")
}
