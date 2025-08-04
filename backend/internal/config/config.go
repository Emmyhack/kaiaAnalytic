package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	ServerAddress string
	LogLevel      logrus.Level

	// Blockchain configuration
	KaiaRPCURL        string
	KaiaChainID       int64
	ContractAddresses ContractAddresses

	// External APIs
	KaiascanAPIKey string
	KaiascanURL    string
	CoinGeckoURL   string

	// Database configuration
	DatabaseURL string

	// Analytics configuration
	AnalyticsWorkerPoolSize int
	AnalyticsUpdateInterval time.Duration

	// Data collection configuration
	DataCollectionInterval time.Duration
	MaxRetries            int

	// Chat configuration
	ChatMaxConcurrentConnections int
	ChatMessageTimeout           time.Duration

	// Security
	JWTSecret string
	CORSOrigins []string
}

// ContractAddresses holds the deployed contract addresses
type ContractAddresses struct {
	AnalyticsRegistry  string
	DataContract       string
	SubscriptionContract string
	ActionContract     string
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		// Server configuration
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		LogLevel:      getLogLevel(getEnv("LOG_LEVEL", "info")),

		// Blockchain configuration
		KaiaRPCURL:  getEnv("KAIA_RPC_URL", "https://testnet-rpc.kaia.network"),
		KaiaChainID: getInt64Env("KAIA_CHAIN_ID", 1337), // Testnet chain ID

		// Contract addresses
		ContractAddresses: ContractAddresses{
			AnalyticsRegistry:   getEnv("CONTRACT_ANALYTICS_REGISTRY", ""),
			DataContract:        getEnv("CONTRACT_DATA", ""),
			SubscriptionContract: getEnv("CONTRACT_SUBSCRIPTION", ""),
			ActionContract:      getEnv("CONTRACT_ACTION", ""),
		},

		// External APIs
		KaiascanAPIKey: getEnv("KAIA_API_KEY", ""),
		KaiascanURL:    getEnv("KAIA_URL", "https://testnet.kaia.network"),
		CoinGeckoURL:   getEnv("COINGECKO_URL", "https://api.coingecko.com/api/v3"),

		// Database configuration
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/kaia_analytics"),

		// Analytics configuration
		AnalyticsWorkerPoolSize: getIntEnv("ANALYTICS_WORKER_POOL_SIZE", 10),
		AnalyticsUpdateInterval: getDurationEnv("ANALYTICS_UPDATE_INTERVAL", 30*time.Second),

		// Data collection configuration
		DataCollectionInterval: getDurationEnv("DATA_COLLECTION_INTERVAL", 1*time.Second),
		MaxRetries:            getIntEnv("MAX_RETRIES", 3),

		// Chat configuration
		ChatMaxConcurrentConnections: getIntEnv("CHAT_MAX_CONNECTIONS", 1000),
		ChatMessageTimeout:           getDurationEnv("CHAT_MESSAGE_TIMEOUT", 30*time.Second),

		// Security
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		CORSOrigins: getStringSliceEnv("CORS_ORIGINS", []string{"*"}),
	}

	return cfg
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getLogLevel(level string) logrus.Level {
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated values parsing
		// In production, you might want more sophisticated parsing
		return []string{value}
	}
	return defaultValue
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Add validation logic here
	// For example, check if required contract addresses are set
	if c.ContractAddresses.AnalyticsRegistry == "" {
		return fmt.Errorf("CONTRACT_ANALYTICS_REGISTRY is required")
	}
	if c.ContractAddresses.DataContract == "" {
		return fmt.Errorf("CONTRACT_DATA is required")
	}
	if c.ContractAddresses.SubscriptionContract == "" {
		return fmt.Errorf("CONTRACT_SUBSCRIPTION is required")
	}
	if c.ContractAddresses.ActionContract == "" {
		return fmt.Errorf("CONTRACT_ACTION is required")
	}
	return nil
}