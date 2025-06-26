package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"audit-log-service/internal/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger provides structured logging capabilities
type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config *config.LoggingConfig) *Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	switch strings.ToLower(config.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	default:
		// Default to JSON format
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	// Set output destination
	switch strings.ToLower(config.Output) {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "file":
		// Use rotating log files
		logWriter := &lumberjack.Logger{
			Filename:   filepath.Join("logs", "app.log"),
			MaxSize:    config.MaxSize, // MB
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge, // days
			Compress:   true,
		}
		logger.SetOutput(logWriter)
	case "both":
		// Output to both stdout and file
		logWriter := &lumberjack.Logger{
			Filename:   filepath.Join("logs", "app.log"),
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   true,
		}
		logger.SetOutput(io.MultiWriter(os.Stdout, logWriter))
	default:
		logger.SetOutput(os.Stdout)
	}

	return &Logger{Logger: logger}
}

// NewDefaultLogger creates a logger with default configuration
func NewDefaultLogger() *Logger {
	config := &config.LoggingConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
	}
	return NewLogger(config)
}

// WithRequestID adds a request ID to the logger context
func (l *Logger) WithRequestID(requestID string) *logrus.Entry {
	return l.WithField("request_id", requestID)
}

// WithTenantID adds a tenant ID to the logger context
func (l *Logger) WithTenantID(tenantID string) *logrus.Entry {
	return l.WithField("tenant_id", tenantID)
}

// WithUserID adds a user ID to the logger context
func (l *Logger) WithUserID(userID string) *logrus.Entry {
	return l.WithField("user_id", userID)
}

// WithComponent adds a component name to the logger context
func (l *Logger) WithComponent(component string) *logrus.Entry {
	return l.WithField("component", component)
}

// WithService adds a service name to the logger context
func (l *Logger) WithService(service string) *logrus.Entry {
	return l.WithField("service", service)
}

// LogAuditEvent logs an audit event with structured data
func (l *Logger) LogAuditEvent(
	tenantID, userID, resource, event, method, ip, environment string,
	data map[string]interface{},
) {
	l.WithFields(logrus.Fields{
		"event_type":  "audit",
		"tenant_id":   tenantID,
		"user_id":     userID,
		"resource":    resource,
		"event":       event,
		"method":      method,
		"ip":          ip,
		"environment": environment,
		"data":        data,
	}).Info("Audit event logged")
}

// LogSecurityEvent logs a security-related event
func (l *Logger) LogSecurityEvent(eventType, message string, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["event_type"] = "security"
	fields["security_event"] = eventType

	l.WithFields(fields).Warn(message)
}

// LogPerformanceMetric logs performance metrics
func (l *Logger) LogPerformanceMetric(operation string, duration int64, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["event_type"] = "performance"
	fields["operation"] = operation
	fields["duration_ms"] = duration

	l.WithFields(fields).Info("Performance metric recorded")
}

// LogError logs an error with context
func (l *Logger) LogError(err error, message string, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["error"] = err.Error()

	l.WithFields(fields).Error(message)
}

// LogAPICall logs API call information
func (l *Logger) LogAPICall(method, path string, statusCode int, duration int64, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["event_type"] = "api_call"
	fields["method"] = method
	fields["path"] = path
	fields["status_code"] = statusCode
	fields["duration_ms"] = duration

	level := logrus.InfoLevel
	if statusCode >= 400 {
		level = logrus.WarnLevel
	}
	if statusCode >= 500 {
		level = logrus.ErrorLevel
	}

	l.WithFields(fields).Log(level, "API call completed")
}
