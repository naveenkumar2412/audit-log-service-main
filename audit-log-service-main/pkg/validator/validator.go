package validator

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator provides validation functionality
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	validate := validator.New()

	// Register custom validation functions
	validate.RegisterValidation("tenant_id", validateTenantID)
	validate.RegisterValidation("user_id", validateUserID)
	validate.RegisterValidation("resource_name", validateResourceName)
	validate.RegisterValidation("event_name", validateEventName)
	validate.RegisterValidation("environment_name", validateEnvironment)
	validate.RegisterValidation("ip_address", validateIPAddress)

	return &Validator{
		validate: validate,
	}
}

// ValidateStruct validates a struct using the validator tags
func (v *Validator) ValidateStruct(s interface{}) error {
	err := v.validate.Struct(s)
	if err != nil {
		return formatValidationError(err)
	}
	return nil
}

// formatValidationError formats validation errors into a readable format
func formatValidationError(err error) error {
	var errorMessages []string

	for _, err := range err.(validator.ValidationErrors) {
		switch err.Tag() {
		case "required":
			errorMessages = append(errorMessages, fmt.Sprintf("%s is required", err.Field()))
		case "min":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param()))
		case "max":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be at most %s characters long", err.Field(), err.Param()))
		case "email":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid email address", err.Field()))
		case "ip":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid IP address", err.Field()))
		case "oneof":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param()))
		case "tenant_id":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid tenant ID (alphanumeric, hyphens, underscores)", err.Field()))
		case "user_id":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid user ID", err.Field()))
		case "resource_name":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid resource name", err.Field()))
		case "event_name":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid event name", err.Field()))
		case "environment_name":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid environment (development, staging, production)", err.Field()))
		case "ip_address":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid IP address", err.Field()))
		default:
			errorMessages = append(errorMessages, fmt.Sprintf("%s failed validation (%s)", err.Field(), err.Tag()))
		}
	}

	return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, ", "))
}

// Custom validation functions

// validateTenantID validates tenant ID format
func validateTenantID(fl validator.FieldLevel) bool {
	tenantID := fl.Field().String()
	if len(tenantID) == 0 || len(tenantID) > 255 {
		return false
	}

	// Allow alphanumeric, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, tenantID)
	return matched
}

// validateUserID validates user ID format
func validateUserID(fl validator.FieldLevel) bool {
	userID := fl.Field().String()
	if len(userID) == 0 || len(userID) > 255 {
		return false
	}

	// Allow alphanumeric, hyphens, underscores, and @ for emails
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_@.-]+$`, userID)
	return matched
}

// validateResourceName validates resource name format
func validateResourceName(fl validator.FieldLevel) bool {
	resource := fl.Field().String()
	if len(resource) == 0 || len(resource) > 255 {
		return false
	}

	// Allow alphanumeric, hyphens, underscores, and forward slashes for REST resources
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_/-]+$`, resource)
	return matched
}

// validateEventName validates event name format
func validateEventName(fl validator.FieldLevel) bool {
	event := fl.Field().String()
	if len(event) == 0 || len(event) > 255 {
		return false
	}

	// Allow alphanumeric, hyphens, underscores, and dots
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.-]+$`, event)
	return matched
}

// validateEnvironment validates environment values
func validateEnvironment(fl validator.FieldLevel) bool {
	environment := strings.ToLower(fl.Field().String())
	validEnvironments := []string{"development", "staging", "production", "test"}

	for _, env := range validEnvironments {
		if environment == env {
			return true
		}
	}
	return false
}

// validateIPAddress validates IP address format (IPv4 or IPv6)
func validateIPAddress(fl validator.FieldLevel) bool {
	ip := fl.Field().String()
	return net.ParseIP(ip) != nil
}

// ValidationRules contains common validation rules
type ValidationRules struct{}

// IsValidHTTPMethod checks if the HTTP method is valid
func (v *ValidationRules) IsValidHTTPMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	method = strings.ToUpper(method)

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

// IsValidTenantID checks if tenant ID is valid
func (v *ValidationRules) IsValidTenantID(tenantID string) bool {
	if len(tenantID) == 0 || len(tenantID) > 255 {
		return false
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, tenantID)
	return matched
}

// IsValidUserID checks if user ID is valid
func (v *ValidationRules) IsValidUserID(userID string) bool {
	if len(userID) == 0 || len(userID) > 255 {
		return false
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_@.-]+$`, userID)
	return matched
}

// IsValidEmail checks if email format is valid
func (v *ValidationRules) IsValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}

// IsValidURL checks if URL format is valid
func (v *ValidationRules) IsValidURL(url string) bool {
	urlRegex := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`
	matched, _ := regexp.MatchString(urlRegex, url)
	return matched
}

// SanitizeString removes potentially dangerous characters from a string
func (v *ValidationRules) SanitizeString(input string) string {
	// Remove null bytes and control characters
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove common SQL injection patterns (basic sanitization)
	dangerousPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"exec", "execute", "select", "insert", "update", "delete",
		"drop", "create", "alter", "union", "script",
	}

	for _, pattern := range dangerousPatterns {
		input = strings.ReplaceAll(strings.ToLower(input), pattern, "")
	}

	return strings.TrimSpace(input)
}

// ValidateStringLength checks if string length is within limits
func (v *ValidationRules) ValidateStringLength(str string, min, max int) bool {
	length := len(str)
	return length >= min && length <= max
}

// ValidateRequired checks if a string field is not empty
func (v *ValidationRules) ValidateRequired(str string) bool {
	return strings.TrimSpace(str) != ""
}
