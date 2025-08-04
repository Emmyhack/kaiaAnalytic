package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CORS middleware configuration
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// Logger middleware for request logging
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"timestamp": param.TimeStamp.Format(time.RFC3339),
			"status":    param.StatusCode,
			"latency":   param.Latency,
			"client_ip": param.ClientIP,
			"method":    param.Method,
			"path":      param.Path,
			"user_agent": param.Request.UserAgent(),
		}).Info("HTTP Request")
		
		return ""
	})
}

// Recovery middleware for panic recovery
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logrus.WithFields(logrus.Fields{
			"error":   recovered,
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
			"client_ip": c.ClientIP(),
		}).Error("Panic recovered")
		
		c.JSON(500, gin.H{
			"error": "Internal server error",
		})
	})
}

// RateLimit middleware for basic rate limiting
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, use Redis or similar for distributed rate limiting
	requests := make(map[string][]time.Time)
	
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()
		
		// Clean old requests
		if times, exists := requests[clientIP]; exists {
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) < window {
					valid = append(valid, t)
				}
			}
			requests[clientIP] = valid
		}
		
		// Check rate limit
		if len(requests[clientIP]) >= limit {
			c.JSON(429, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}
		
		// Add current request
		requests[clientIP] = append(requests[clientIP], now)
		c.Next()
	})
}

// Auth middleware for subscription-based access control
func Auth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Extract user address from header or query param
		userAddress := c.GetHeader("X-User-Address")
		if userAddress == "" {
			userAddress = c.Query("user_address")
		}
		
		if userAddress == "" {
			c.JSON(401, gin.H{
				"error": "User address required",
			})
			c.Abort()
			return
		}
		
		// Store user address in context for later use
		c.Set("user_address", userAddress)
		c.Next()
	})
}

// PremiumAuth middleware for premium feature access
func PremiumAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userAddress := c.GetString("user_address")
		if userAddress == "" {
			c.JSON(401, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}
		
		// In a real implementation, check subscription status
		// For now, allow all authenticated users
		c.Next()
	})
}

// RequestID middleware adds a unique request ID
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}