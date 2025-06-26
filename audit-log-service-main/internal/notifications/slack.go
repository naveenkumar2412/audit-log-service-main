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

// SlackSender interface for sending Slack messages
type SlackSender interface {
	Send(ctx context.Context, channel, message string) error
}

// WebhookSlackSender implements SlackSender using Slack webhooks
type WebhookSlackSender struct {
	config *config.SlackConfig
	client *http.Client
}

// NewWebhookSlackSender creates a new webhook Slack sender
func NewWebhookSlackSender(config *config.SlackConfig) *WebhookSlackSender {
	return &WebhookSlackSender{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SlackMessage represents a Slack message payload
type SlackMessage struct {
	Text     string `json:"text"`
	Channel  string `json:"channel,omitempty"`
	Username string `json:"username,omitempty"`
}

// Send sends a message to Slack
func (s *WebhookSlackSender) Send(ctx context.Context, channel, message string) error {
	if !s.config.Enabled || s.config.WebhookURL == "" {
		return nil // Skip if Slack notifications are disabled or webhook URL is not set
	}

	slackMessage := SlackMessage{
		Text:     message,
		Channel:  channel,
		Username: s.config.Username,
	}

	payload, err := json.Marshal(slackMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// MockSlackSender is a mock implementation for testing
type MockSlackSender struct {
	SentMessages []MockSlackMessage
}

// MockSlackMessage represents a sent Slack message for testing
type MockSlackMessage struct {
	Channel string
	Message string
}

// NewMockSlackSender creates a new mock Slack sender
func NewMockSlackSender() *MockSlackSender {
	return &MockSlackSender{
		SentMessages: make([]MockSlackMessage, 0),
	}
}

// Send records the message instead of actually sending it
func (m *MockSlackSender) Send(ctx context.Context, channel, message string) error {
	m.SentMessages = append(m.SentMessages, MockSlackMessage{
		Channel: channel,
		Message: message,
	})
	return nil
}

// GetSentMessages returns all sent messages
func (m *MockSlackSender) GetSentMessages() []MockSlackMessage {
	return m.SentMessages
}

// Clear clears all sent messages
func (m *MockSlackSender) Clear() {
	m.SentMessages = make([]MockSlackMessage, 0)
}