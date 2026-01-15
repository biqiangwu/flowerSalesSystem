package config

import (
	"os"
	"strconv"
)

// Config 应用程序配置
// 所有配置项通过环境变量注入，具有合理的默认值
type Config struct {
	// 数据库配置
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string

	// Session 配置
	SessionSecret string
	SessionExpiry int // hours

	// 服务器配置
	ServerPort int
	LogLevel   string

	// 业务配置
	StockWarningThreshold int
}

// Load 从环境变量加载配置
// 如果环境变量未设置，使用默认值
func Load() *Config {
	return &Config{
		DBHost:               getEnv("DB_HOST", "mysql-service"),
		DBPort:               getEnvInt("DB_PORT", 3306),
		DBName:               getEnv("DB_NAME", "flower_sales"),
		DBUser:               getEnv("DB_USER", "flower_user"),
		DBPassword:           getEnv("DB_PASSWORD", ""),
		SessionSecret:        getEnv("SESSION_SECRET", ""),
		SessionExpiry:        getEnvInt("SESSION_EXPIRY", 24),
		ServerPort:           getEnvInt("SERVER_PORT", 8080),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		StockWarningThreshold: getEnvInt("STOCK_WARNING_THRESHOLD", 10),
	}
}

// getEnv 从环境变量获取字符串值，如果未设置则返回默认值
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

// getEnvInt 从环境变量获取整数值，如果未设置或转换失败则返回默认值
func getEnvInt(key string, defaultVal int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}
	return value
}
