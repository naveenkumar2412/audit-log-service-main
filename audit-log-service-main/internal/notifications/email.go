package notifications

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"audit-log-service/internal/config"
)

// EmailSender interface for sending emails
type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// SMTPEmailSender implements EmailSender using SMTP
type SMTPEmailSender struct {
	config *config.EmailConfig
}

// NewSMTPEmailSender creates a new SMTP email sender
func NewSMTPEmailSender(config *config.EmailConfig) *SMTPEmailSender {
	return &SMTPEmailSender{
		config: config,
	}
}

// Send sends an email using SMTP
func (s *SMTPEmailSender) Send(ctx context.Context, to, subject, body string) error {
	if !s.config.Enabled {
		return nil // Skip if email notifications are disabled
	}

	// Create the email message
	msg := s.buildMessage(s.config.From, to, subject, body)

	// Set up authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)

	// Send the email
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	err := smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildMessage builds the email message
func (s *SMTPEmailSender) buildMessage(from, to, subject, body string) string {
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)
	return msg.String()
}

// MockEmailSender is a mock implementation for testing
type MockEmailSender struct {
	SentEmails []MockEmail
}

// MockEmail represents a sent email for testing
type MockEmail struct {
	To      string
	Subject string
	Body    string
}

// NewMockEmailSender creates a new mock email sender
func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{
		SentEmails: make([]MockEmail, 0),
	}
}

// Send records the email instead of actually sending it
func (m *MockEmailSender) Send(ctx context.Context, to, subject, body string) error {
	m.SentEmails = append(m.SentEmails, MockEmail{
		To:      to,
		Subject: subject,
		Body:    body,
	})
	return nil
}

// GetSentEmails returns all sent emails
func (m *MockEmailSender) GetSentEmails() []MockEmail {
	return m.SentEmails
}

// Clear clears all sent emails
func (m *MockEmailSender) Clear() {
	m.SentEmails = make([]MockEmail, 0)
}
