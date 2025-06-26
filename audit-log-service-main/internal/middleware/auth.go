package middleware

import (
	"net/http"
	"strings"

	"audit-log-service/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	apiKeyHeader        = "X-API-Key"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	config *config.AuthConfig
	logger *logrus.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(config *config.AuthConfig, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		config: config,
		logger: logger,
	}
}

// JWTAuth validates JWT tokens
func (am *AuthMiddleware) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeader)
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			am.logger.Warn("Missing or malformed Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header missing or malformed",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(am.config.JWTSecret), nil
		})

		if err != nil {
			am.logger.WithError(err).Warn("Failed to parse JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			am.logger.Warn("Invalid JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["sub"])
			c.Set("tenant_id", claims["tenant_id"])
			c.Set("roles", claims["roles"])
		}

		c.Next()
	}
}

// APIKeyAuth validates API keys
func (am *AuthMiddleware) APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(apiKeyHeader)
		if apiKey == "" {
			// Try to get from query parameter as fallback
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			am.logger.Warn("Missing API key")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key is required",
			})
			c.Abort()
			return
		}

		// Validate API key
		valid := false
		for _, validKey := range am.config.APIKeys {
			if apiKey == validKey {
				valid = true
				break
			}
		}

		if !valid {
			am.logger.WithField("api_key", apiKey).Warn("Invalid API key")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		// For API key auth, we don't have user context
		c.Set("auth_type", "api_key")
		c.Next()
	}
}

// OptionalAuth allows both JWT and API key authentication
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT first
		authHeader := c.GetHeader(authorizationHeader)
		if strings.HasPrefix(authHeader, bearerPrefix) {
			tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(am.config.JWTSecret), nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					c.Set("user_id", claims["sub"])
					c.Set("tenant_id", claims["tenant_id"])
					c.Set("roles", claims["roles"])
					c.Set("auth_type", "jwt")
				}
				c.Next()
				return
			}
		}

		// Try API key
		apiKey := c.GetHeader(apiKeyHeader)
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey != "" {
			for _, validKey := range am.config.APIKeys {
				if apiKey == validKey {
					c.Set("auth_type", "api_key")
					c.Next()
					return
				}
			}
		}

		// No valid authentication found
		am.logger.Warn("No valid authentication provided")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Valid authentication is required (JWT token or API key)",
		})
		c.Abort()
	}
}

// RequireTenant ensures the user belongs to a specific tenant
func (am *AuthMiddleware) RequireTenant(requiredTenant string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Tenant information not available",
			})
			c.Abort()
			return
		}

		if tenantID.(string) != requiredTenant {
			am.logger.WithFields(logrus.Fields{
				"required_tenant": requiredTenant,
				"user_tenant":     tenantID,
			}).Warn("Tenant access denied")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied for this tenant",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole ensures the user has a specific role
func (am *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Role information not available",
			})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]interface{})
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid role information",
			})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range userRoles {
			if roleStr, ok := role.(string); ok && roleStr == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			am.logger.WithFields(logrus.Fields{
				"required_role": requiredRole,
				"user_roles":    userRoles,
			}).Warn("Role access denied")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient privileges",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
