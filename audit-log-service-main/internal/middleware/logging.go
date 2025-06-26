package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LoggingMiddleware provides comprehensive request logging
type LoggingMiddleware struct {
	logger *logrus.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *logrus.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// RequestLogger logs HTTP requests with detailed information
func (lm *LoggingMiddleware) RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Record start time
		start := time.Now()

		// Log request
		lm.logger.WithFields(logrus.Fields{
			"request_id":     requestID,
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"query":          c.Request.URL.RawQuery,
			"ip":             c.ClientIP(),
			"user_agent":     c.Request.UserAgent(),
			"content_length": c.Request.ContentLength,
		}).Info("Request started")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log response
		logLevel := logrus.InfoLevel
		if c.Writer.Status() >= 400 {
			logLevel = logrus.WarnLevel
		}
		if c.Writer.Status() >= 500 {
			logLevel = logrus.ErrorLevel
		}

		lm.logger.WithFields(logrus.Fields{
			"request_id":    requestID,
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status":        c.Writer.Status(),
			"duration":      duration.String(),
			"duration_ms":   float64(duration.Nanoseconds()) / 1000000,
			"response_size": c.Writer.Size(),
			"ip":            c.ClientIP(),
		}).Log(logLevel, "Request completed")
	}
}

// ResponseWriter wraps gin.ResponseWriter to capture response body
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w ResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// DetailedRequestLogger logs requests with request/response body (use with caution in production)
func (lm *LoggingMiddleware) DetailedRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore the io.ReadCloser to its original state
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrap response writer
		responseWriter := &ResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseWriter

		// Record start time
		start := time.Now()

		// Log request with body
		lm.logger.WithFields(logrus.Fields{
			"request_id":     requestID,
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"query":          c.Request.URL.RawQuery,
			"ip":             c.ClientIP(),
			"user_agent":     c.Request.UserAgent(),
			"request_body":   string(requestBody),
			"content_length": c.Request.ContentLength,
		}).Info("Detailed request started")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log response with body
		logLevel := logrus.InfoLevel
		if c.Writer.Status() >= 400 {
			logLevel = logrus.WarnLevel
		}
		if c.Writer.Status() >= 500 {
			logLevel = logrus.ErrorLevel
		}

		lm.logger.WithFields(logrus.Fields{
			"request_id":    requestID,
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status":        c.Writer.Status(),
			"duration":      duration.String(),
			"duration_ms":   float64(duration.Nanoseconds()) / 1000000,
			"response_size": c.Writer.Size(),
			"response_body": responseWriter.body.String(),
			"ip":            c.ClientIP(),
		}).Log(logLevel, "Detailed request completed")
	}
}

// ErrorLogger logs errors with context
func (lm *LoggingMiddleware) ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log any errors that occurred during request processing
		for _, err := range c.Errors {
			requestID, _ := c.Get("request_id")
			lm.logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"ip":         c.ClientIP(),
				"error":      err.Error(),
				"type":       err.Type,
			}).Error("Request error occurred")
		}
	}
}

// RecoveryLogger logs panic recoveries
func (lm *LoggingMiddleware) RecoveryLogger() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID, _ := c.Get("request_id")
		lm.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP(),
			"panic":      recovered,
		}).Error("Panic recovered")

		c.AbortWithStatus(500)
	})
}
