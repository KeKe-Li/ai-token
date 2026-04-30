package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	EncryptionKey   string
	AllowedOrigins  string

	// 供应商密钥（快速启动用，生产环境从数据库加载）
	OpenAIKey       string
	OpenAIBaseURL   string
	AnthropicKey    string
	AnthropicBaseURL string
	GoogleKey       string
	GoogleBaseURL   string
	DeepSeekKey     string
	DeepSeekBaseURL string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://aitoken:aitoken@localhost:5432/aitoken?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		EncryptionKey:  getEnv("ENCRYPTION_KEY", "dev-encryption-key-32bytes!!!!!"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),

		OpenAIKey:        getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL:    getEnv("OPENAI_BASE_URL", "https://api.openai.com"),
		AnthropicKey:     getEnv("ANTHROPIC_API_KEY", ""),
		AnthropicBaseURL: getEnv("ANTHROPIC_BASE_URL", "https://api.anthropic.com"),
		GoogleKey:        getEnv("GOOGLE_API_KEY", ""),
		GoogleBaseURL:    getEnv("GOOGLE_BASE_URL", "https://generativelanguage.googleapis.com"),
		DeepSeekKey:      getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL:  getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
