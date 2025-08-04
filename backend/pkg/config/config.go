package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Environment string
	Port        int
	Host        string

	// Database configuration
	DatabaseURL string
	RedisURL    string

	// Blockchain configuration
	KaiaRPCURL        string
	KaiaTestnetRPCURL string
	KaiaMainnetRPCURL string
	NetworkID         int64

	// Contract addresses
	ContractAddresses ContractAddresses

	// External API configuration
	KaiascanAPIKey    string
	KaiascanBaseURL   string
	CoinGeckoAPIKey   string
	CoinGeckoBaseURL  string

	// AI/NLP configuration
	OpenAIAPIKey      string
	LangChainEndpoint string

	// Service configuration
	WorkerPoolSize    int
	MaxConcurrentJobs int
	DataRetentionDays int

	// Security configuration
	JWTSecret           string
	EncryptionKey       string
	RateLimitPerMinute  int
	MaxRequestSize      int64

	// Monitoring configuration
	LogLevel        string
	MetricsEnabled  bool
	TracingEnabled  bool

	// Feature flags
	EnableAnalytics bool
	EnableChat      bool
	EnableActions   bool
}

// ContractAddresses holds all smart contract addresses
type ContractAddresses struct {
	AnalyticsRegistry  string
	DataContract       string
	SubscriptionContract string
	ActionContract     string
	KAIAToken          string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		// Server defaults
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnvAsInt("PORT", 8080),
		Host:        getEnv("HOST", "localhost"),

		// Database defaults
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/kaia_analytics?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),

		// Blockchain defaults
		KaiaRPCURL:        getEnv("KAIA_RPC_URL", "https://rpc.kaia.io"),
		KaiaTestnetRPCURL: getEnv("KAIA_TESTNET_RPC_URL", "https://rpc-testnet.kaia.io"),
		KaiaMainnetRPCURL: getEnv("KAIA_MAINNET_RPC_URL", "https://rpc-mainnet.kaia.io"),
		NetworkID:         getEnvAsInt64("NETWORK_ID", 1001), // Kaia testnet

		// Contract addresses
		ContractAddresses: ContractAddresses{
			AnalyticsRegistry:    getEnv("ANALYTICS_REGISTRY_ADDRESS", ""),
			DataContract:         getEnv("DATA_CONTRACT_ADDRESS", ""),
			SubscriptionContract: getEnv("SUBSCRIPTION_CONTRACT_ADDRESS", ""),
			ActionContract:       getEnv("ACTION_CONTRACT_ADDRESS", ""),
			KAIAToken:           getEnv("KAIA_TOKEN_ADDRESS", ""),
		},

		// External APIs
		KaiascanAPIKey:   getEnv("KAIASCAN_API_KEY", ""),
		KaiascanBaseURL:  getEnv("KAIASCAN_BASE_URL", "https://api.kaiascan.io"),
		CoinGeckoAPIKey:  getEnv("COINGECKO_API_KEY", ""),
		CoinGeckoBaseURL: getEnv("COINGECKO_BASE_URL", "https://api.coingecko.com/api/v3"),

		// AI/NLP
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		LangChainEndpoint: getEnv("LANGCHAIN_ENDPOINT", ""),

		// Service configuration
		WorkerPoolSize:    getEnvAsInt("WORKER_POOL_SIZE", 10),
		MaxConcurrentJobs: getEnvAsInt("MAX_CONCURRENT_JOBS", 100),
		DataRetentionDays: getEnvAsInt("DATA_RETENTION_DAYS", 90),

		// Security
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key"),
		EncryptionKey:      getEnv("ENCRYPTION_KEY", "your-encryption-key"),
		RateLimitPerMinute: getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
		MaxRequestSize:     getEnvAsInt64("MAX_REQUEST_SIZE", 10485760), // 10MB

		// Monitoring
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		MetricsEnabled: getEnvAsBool("METRICS_ENABLED", true),
		TracingEnabled: getEnvAsBool("TRACING_ENABLED", false),

		// Feature flags
		EnableAnalytics: getEnvAsBool("ENABLE_ANALYTICS", true),
		EnableChat:      getEnvAsBool("ENABLE_CHAT", true),
		EnableActions:   getEnvAsBool("ENABLE_ACTIONS", true),
	}

	return config, nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsTestnet returns true if connected to testnet
func (c *Config) IsTestnet() bool {
	return c.NetworkID == 1001 // Kaia testnet
}

// IsMainnet returns true if connected to mainnet
func (c *Config) IsMainnet() bool {
	return c.NetworkID == 8217 // Kaia mainnet
}

// GetRPCURL returns the appropriate RPC URL based on network
func (c *Config) GetRPCURL() string {
	if c.IsMainnet() {
		return c.KaiaMainnetRPCURL
	}
	return c.KaiaTestnetRPCURL
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Add validation logic here
	return nil
}

// Helper functions for environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string, separator string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, separator)
	}
	return defaultValue
}