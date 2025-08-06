package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server Configuration
	Port        string
	Environment string
	Debug       bool

	// Database Configuration
	DatabaseURL     string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration

	// JWT Configuration
	JWTSecret          string
	JWTExpirationHours int
	JWTRefreshHours    int

	// Blockchain Configuration
	Web3RPCURL      string
	PrivateKey      string
	ContractAddress string
	ChainID         int64
	GasLimit        uint64
	GasPrice        int64

	// IPFS Configuration
	IPFSAPIUrl        string
	IPFSProjectID     string
	IPFSProjectSecret string
	IPFSGatewayURL    string

	// Security Configuration
	BcryptCost     int
	RateLimitRPM   int
	MaxFileSize    int64
	AllowedOrigins []string

	// External Services
	ESGAPIUrl       string
	ESGAPIKey       string
	NotificationURL string

	// Feature Flags
	EnableZKProofs  bool
	EnableAuditLogs bool
	EnableMetrics   bool

	// File Upload
	UploadPath       string
	MaxZipSize       int64
	AllowedFileTypes []string
}

var AppConfig *Config

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		// Server defaults
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Debug:       getEnvBool("DEBUG", true),

		// Database defaults
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "password"),
		DBName:          getEnv("DB_NAME", "aluminium_passport"),
		DBSSLMode:       getEnv("DB_SSL_MODE", "disable"),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute,

		// JWT defaults
		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
		JWTRefreshHours:    getEnvInt("JWT_REFRESH_HOURS", 168), // 7 days

		// Blockchain defaults
		Web3RPCURL:      getEnv("WEB3_RPC_URL", "https://polygon-mainnet.infura.io/v3/YOUR_PROJECT_ID"),
		PrivateKey:      getEnv("PRIVATE_KEY", ""),
		ContractAddress: getEnv("CONTRACT_ADDRESS", ""),
		ChainID:         getEnvInt64("CHAIN_ID", 137), // Polygon mainnet
		GasLimit:        uint64(getEnvInt64("GAS_LIMIT", 300000)),
		GasPrice:        getEnvInt64("GAS_PRICE", 20000000000), // 20 gwei

		// IPFS defaults
		IPFSAPIUrl:        getEnv("IPFS_API_URL", "https://ipfs.infura.io:5001"),
		IPFSProjectID:     getEnv("IPFS_PROJECT_ID", ""),
		IPFSProjectSecret: getEnv("IPFS_PROJECT_SECRET", ""),
		IPFSGatewayURL:    getEnv("IPFS_GATEWAY_URL", "https://gateway.pinata.cloud/ipfs/"),

		// Security defaults
		BcryptCost:     getEnvInt("BCRYPT_COST", 12),
		RateLimitRPM:   getEnvInt("RATE_LIMIT_RPM", 100),
		MaxFileSize:    getEnvInt64("MAX_FILE_SIZE", 50*1024*1024), // 50MB
		AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),

		// External services
		ESGAPIUrl:       getEnv("ESG_API_URL", "https://api.esg-scoring.com/v1"),
		ESGAPIKey:       getEnv("ESG_API_KEY", ""),
		NotificationURL: getEnv("NOTIFICATION_URL", ""),

		// Feature flags
		EnableZKProofs:  getEnvBool("ENABLE_ZK_PROOFS", true),
		EnableAuditLogs: getEnvBool("ENABLE_AUDIT_LOGS", true),
		EnableMetrics:   getEnvBool("ENABLE_METRICS", true),

		// File upload
		UploadPath:       getEnv("UPLOAD_PATH", "./uploads"),
		MaxZipSize:       getEnvInt64("MAX_ZIP_SIZE", 100*1024*1024), // 100MB
		AllowedFileTypes: getEnvSlice("ALLOWED_FILE_TYPES", []string{".json", ".csv", ".xlsx"}),
	}

	// Build database URL if not provided
	if config.DatabaseURL == "" {
		config.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			config.DBUser,
			config.DBPassword,
			config.DBHost,
			config.DBPort,
			config.DBName,
			config.DBSSLMode,
		)
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, err
	}

	AppConfig = config
	return config, nil
}

func (c *Config) Validate() error {
	if c.JWTSecret == "your-super-secret-jwt-key-change-in-production" && c.Environment == "production" {
		return fmt.Errorf("JWT_SECRET must be changed in production")
	}

	if c.Environment == "production" {
		if c.Web3RPCURL == "" {
			return fmt.Errorf("WEB3_RPC_URL is required in production")
		}
		if c.PrivateKey == "" {
			return fmt.Errorf("PRIVATE_KEY is required in production")
		}
		if c.ContractAddress == "" {
			return fmt.Errorf("CONTRACT_ADDRESS is required in production")
		}
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) GetDatabaseDSN() string {
	return c.DatabaseURL
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		// In production, you might want more sophisticated parsing
		return []string{value}
	}
	return defaultValue
}
