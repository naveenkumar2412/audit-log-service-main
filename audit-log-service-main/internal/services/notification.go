package services

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"audit-log-service/internal/config"
	"audit-log-service/internal/models"
	"audit-log-service/internal/notifications"

	"github.com/sirupsen/logrus"
)

// NotificationService handles sending notifications for audit events
type NotificationService struct {
	config        *config.NotificationConfig
	emailSender   notifications.EmailSender
	slackSender   notifications.SlackSender
	webhookSender notifications.WebhookSender
	logger        *logrus.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	config *config.NotificationConfig,
	emailSender notifications.EmailSender,
	slackSender notifications.SlackSender,
	webhookSender notifications.WebhookSender,
	logger *logrus.Logger,
) *NotificationService {
	return &NotificationService{
		config:        config,
		emailSender:   emailSender,
		slackSender:   slackSender,
		webhookSender: webhookSender,
		logger:        logger,
	}
}

// SendNotification sends notifications based on audit log events
func (ns *NotificationService) SendNotification(ctx context.Context, auditLog *models.AuditLog) error {
	// Check if any notification is enabled
	if !ns.shouldSendNotification(auditLog) {
		return nil
	}

	var wg sync.WaitGroup
	errors := make(chan error, 3) // Buffer for up to 3 notification types

	// Send email notification
	if ns.config.Email.Enabled && ns.shouldNotifyByEmail(auditLog) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ns.sendEmailNotification(ctx, auditLog); err != nil {
				errors <- fmt.Errorf("email notification failed: %w", err)
			}
		}()
	}

	// Send Slack notification
	if ns.config.Slack.Enabled && ns.shouldNotifyBySlack(auditLog) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ns.sendSlackNotification(ctx, auditLog); err != nil {
				errors <- fmt.Errorf("slack notification failed: %w", err)
			}
		}()
	}

	// Send webhook notification
	if ns.config.Webhook.Enabled && ns.shouldNotifyByWebhook(auditLog) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ns.sendWebhookNotification(ctx, auditLog); err != nil {
				errors <- fmt.Errorf("webhook notification failed: %w", err)
			}
		}()
	}

	// Wait for all notifications to complete
	go func() {
		wg.Wait()
		close(errors)
	}()

	// Collect any errors
	var notificationErrors []string
	for err := range errors {
		notificationErrors = append(notificationErrors, err.Error())
		ns.logger.WithError(err).Error("Notification failed")
	}

	if len(notificationErrors) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(notificationErrors, "; "))
	}

	return nil
}

// sendEmailNotification sends email notification
func (ns *NotificationService) sendEmailNotification(ctx context.Context, auditLog *models.AuditLog) error {
	if ns.emailSender == nil {
		return fmt.Errorf("email sender not configured")
	}

	message := ns.formatEmailMessage(auditLog)
	subject := fmt.Sprintf("Audit Alert: %s - %s", auditLog.Event, auditLog.Resource)

	for _, recipient := range ns.config.Email.To {
		if err := ns.emailSender.Send(ctx, recipient, subject, message); err != nil {
			return fmt.Errorf("failed to send email to %s: %w", recipient, err)
		}
	}

	ns.logger.WithFields(logrus.Fields{
		"audit_log_id": auditLog.ID,
		"recipients":   len(ns.config.Email.To),
	}).Info("Email notification sent successfully")

	return nil
}

// sendSlackNotification sends Slack notification
func (ns *NotificationService) sendSlackNotification(ctx context.Context, auditLog *models.AuditLog) error {
	if ns.slackSender == nil {
		return fmt.Errorf("slack sender not configured")
	}

	message := ns.formatSlackMessage(auditLog)
	channel := ns.config.Slack.Channel
	if channel == "" {
		channel = "#audit-alerts"
	}

	if err := ns.slackSender.Send(ctx, channel, message); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}

	ns.logger.WithFields(logrus.Fields{
		"audit_log_id": auditLog.ID,
		"channel":      channel,
	}).Info("Slack notification sent successfully")

	return nil
}

// sendWebhookNotification sends webhook notification
func (ns *NotificationService) sendWebhookNotification(ctx context.Context, auditLog *models.AuditLog) error {
	if ns.webhookSender == nil {
		return fmt.Errorf("webhook sender not configured")
	}

	payload := ns.formatWebhookPayload(auditLog)

	for _, url := range ns.config.Webhook.URLs {
		if err := ns.webhookSender.Send(ctx, url, payload); err != nil {
			return fmt.Errorf("failed to send webhook to %s: %w", url, err)
		}
	}

	ns.logger.WithFields(logrus.Fields{
		"audit_log_id": auditLog.ID,
		"webhook_urls": len(ns.config.Webhook.URLs),
	}).Info("Webhook notification sent successfully")

	return nil
}

// shouldSendNotification determines if any notification should be sent
func (ns *NotificationService) shouldSendNotification(auditLog *models.AuditLog) bool {
	// Check if any notification method is enabled
	return (ns.config.Email.Enabled && ns.shouldNotifyByEmail(auditLog)) ||
		(ns.config.Slack.Enabled && ns.shouldNotifyBySlack(auditLog)) ||
		(ns.config.Webhook.Enabled && ns.shouldNotifyByWebhook(auditLog))
}

// shouldNotifyByEmail determines if email notification should be sent
func (ns *NotificationService) shouldNotifyByEmail(auditLog *models.AuditLog) bool {
	// Add logic to determine when to send email notifications
	// For example, only for critical events or specific resources
	criticalEvents := []string{"DELETE", "UNAUTHORIZED_ACCESS", "SECURITY_BREACH"}

	for _, event := range criticalEvents {
		if strings.Contains(strings.ToUpper(auditLog.Event), event) {
			return true
		}
	}

	return false
}

// shouldNotifyBySlack determines if Slack notification should be sent
func (ns *NotificationService) shouldNotifyBySlack(auditLog *models.AuditLog) bool {
	// Add logic to determine when to send Slack notifications
	// For example, for all events in production environment
	return auditLog.Environment == "production"
}

// shouldNotifyByWebhook determines if webhook notification should be sent
func (ns *NotificationService) shouldNotifyByWebhook(auditLog *models.AuditLog) bool {
	// Add logic to determine when to send webhook notifications
	// For example, for all events
	return true
}

// formatEmailMessage formats the audit log for email
func (ns *NotificationService) formatEmailMessage(auditLog *models.AuditLog) string {
	return fmt.Sprintf(`
Audit Log Alert

ID: %s
Tenant: %s
User: %s
Resource: %s
Event: %s
Method: %s
IP Address: %s
Environment: %s
Timestamp: %s

Data: %s

This is an automated notification from the Audit Log Service.
`,
		auditLog.ID,
		auditLog.TenantID,
		auditLog.UserID,
		auditLog.Resource,
		auditLog.Event,
		auditLog.Method,
		auditLog.IP,
		auditLog.Environment,
		auditLog.Timestamp.Format("2006-01-02 15:04:05 UTC"),
		string(auditLog.Data),
	)
}

// formatSlackMessage formats the audit log for Slack
func (ns *NotificationService) formatSlackMessage(auditLog *models.AuditLog) string {
	return fmt.Sprintf(`🚨 *Audit Alert*

*Event:* %s
*Resource:* %s
*User:* %s
*Environment:* %s
*IP:* %s
*Time:* %s

*Tenant:* %s
*Method:* %s
*ID:* %s`,
		auditLog.Event,
		auditLog.Resource,
		auditLog.UserID,
		auditLog.Environment,
		auditLog.IP,
		auditLog.Timestamp.Format("2006-01-02 15:04:05 UTC"),
		auditLog.TenantID,
		auditLog.Method,
		auditLog.ID,
	)
}

// formatWebhookPayload formats the audit log for webhook
func (ns *NotificationService) formatWebhookPayload(auditLog *models.AuditLog) map[string]interface{} {
	return map[string]interface{}{
		"id":                auditLog.ID,
		"tenant_id":         auditLog.TenantID,
		"user_id":           auditLog.UserID,
		"resource":          auditLog.Resource,
		"event":             auditLog.Event,
		"method":            auditLog.Method,
		"ip":                auditLog.IP,
		"environment":       auditLog.Environment,
		"timestamp":         auditLog.Timestamp,
		"data":              auditLog.Data,
		"meta":              auditLog.Meta,
		"notification_type": "audit_alert",
	}
}
