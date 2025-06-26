package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"audit-log-service/internal/config"
)

// WebhookSender interface for sending webhook notifications
type WebhookSender interface {
	Send(ctx context.Context, url string, payload map[string]interface{}) error
}

// HTTPWebhookSender implements WebhookSender using HTTP
type HTTPWebhookSender struct {
	config *config.WebhookConfig
	client *http.Client
}

// NewHTTPWebhookSender creates a new HTTP webhook sender
func NewHTTPWebhookSender(config *config.WebhookConfig) *HTTPWebhookSender {
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &HTTPWebhookSender{
		config: config,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Send sends a webhook notification
func (w *HTTPWebhookSender) Send(ctx context.Context, url string, payload map[string]interface{}) error {
	if !w.config.Enabled {
		return nil // Skip if webhook notifications are disabled
	}

	// Add metadata to payload
	payload["timestamp"] = time.Now().UTC()
	payload["service"] = "audit-log-service"

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "audit-log-service/1.0")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response status is acceptable (2xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// MockWebhookSender is a mock implementation for testing
type MockWebhookSender struct {
	SentWebhooks []MockWebhook
}

// MockWebhook represents a sent webhook for testing
type MockWebhook struct {
	URL     string
	Payload map[string]interface{}
}

// NewMockWebhookSender creates a new mock webhook sender
func NewMockWebhookSender() *MockWebhookSender {
	return &MockWebhookSender{
		SentWebhooks: make([]MockWebhook, 0),
	}
}

// Send records the webhook instead of actually sending it
func (m *MockWebhookSender) Send(ctx context.Context, url string, payload map[string]interface{}) error {
	m.SentWebhooks = append(m.SentWebhooks, MockWebhook{
		URL:     url,
		Payload: payload,
	})
	return nil
}

// GetSentWebhooks returns all sent webhooks
func (m *MockWebhookSender) GetSentWebhooks() []MockWebhook {
	return m.SentWebhooks
}

// Clear clears all sent webhooks
func (m *MockWebhookSender) Clear() {
	m.SentWebhooks = make([]MockWebhook, 0)
}
