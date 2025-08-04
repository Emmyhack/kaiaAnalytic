package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger creates a new structured logger
func NewLogger() *logrus.Logger {
	logger := logrus.New()

	// Set output to stdout
	logger.SetOutput(os.Stdout)

	// Set log level based on environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Set formatter
	if os.Getenv("ENVIRONMENT") == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	}

	return logger
}

// WithFields creates a logger with predefined fields
func WithFields(logger *logrus.Logger, fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}

// WithService creates a logger with service name
func WithService(logger *logrus.Logger, serviceName string) *logrus.Entry {
	return logger.WithField("service", serviceName)
}

// WithUser creates a logger with user context
func WithUser(logger *logrus.Logger, userID string) *logrus.Entry {
	return logger.WithField("user_id", userID)
}

// WithRequest creates a logger with request context
func WithRequest(logger *logrus.Logger, method, path, userAgent string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"method":     method,
		"path":       path,
		"user_agent": userAgent,
	})
}