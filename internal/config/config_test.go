package config

import (
	"os"
	"testing"
)

// TestConfigLoad_Defaults 测试所有环境变量未设置时使用默认值
func TestConfigLoad_Defaults(t *testing.T) {
	// 清除所有相关环境变量
	envVars := []string{
		"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD",
		"SESSION_SECRET", "SESSION_EXPIRY", "SERVER_PORT", "LOG_LEVEL", "STOCK_WARNING_THRESHOLD",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}

	cfg := Load()

	if cfg.DBHost != "mysql-service" {
		t.Errorf("DBHost = %s, want %s", cfg.DBHost, "mysql-service")
	}
	if cfg.DBPort != 3306 {
		t.Errorf("DBPort = %d, want %d", cfg.DBPort, 3306)
	}
	if cfg.DBName != "flower_sales" {
		t.Errorf("DBName = %s, want %s", cfg.DBName, "flower_sales")
	}
	if cfg.DBUser != "flower_user" {
		t.Errorf("DBUser = %s, want %s", cfg.DBUser, "flower_user")
	}
	if cfg.SessionExpiry != 24 {
		t.Errorf("SessionExpiry = %d, want %d", cfg.SessionExpiry, 24)
	}
	if cfg.ServerPort != 8080 {
		t.Errorf("ServerPort = %d, want %d", cfg.ServerPort, 8080)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %s, want %s", cfg.LogLevel, "info")
	}
	if cfg.StockWarningThreshold != 10 {
		t.Errorf("StockWarningThreshold = %d, want %d", cfg.StockWarningThreshold, 10)
	}
}

// TestConfigLoad_WithDBHost 测试设置 DB_HOST 环境变量
func TestConfigLoad_WithDBHost(t *testing.T) {
	t.Setenv("DB_HOST", "custom-db")

	cfg := Load()

	if cfg.DBHost != "custom-db" {
		t.Errorf("DBHost = %s, want %s", cfg.DBHost, "custom-db")
	}
}

// TestConfigLoad_WithDBPort 测试设置 DB_PORT 环境变量
func TestConfigLoad_WithDBPort(t *testing.T) {
	t.Setenv("DB_PORT", "3307")

	cfg := Load()

	if cfg.DBPort != 3307 {
		t.Errorf("DBPort = %d, want %d", cfg.DBPort, 3307)
	}
}

// TestConfigLoad_AllEnvVars 测试设置所有环境变量
func TestConfigLoad_AllEnvVars(t *testing.T) {
	t.Setenv("DB_HOST", "prod-db.example.com")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "production_db")
	t.Setenv("DB_USER", "admin_user")
	t.Setenv("DB_PASSWORD", "secret-password")
	t.Setenv("SESSION_SECRET", "my-secret-key")
	t.Setenv("SESSION_EXPIRY", "48")
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("STOCK_WARNING_THRESHOLD", "20")

	cfg := Load()

	tests := []struct {
		name  string
		got   any
		want  any
	}{
		{"DBHost", cfg.DBHost, "prod-db.example.com"},
		{"DBPort", cfg.DBPort, 5432},
		{"DBName", cfg.DBName, "production_db"},
		{"DBUser", cfg.DBUser, "admin_user"},
		{"DBPassword", cfg.DBPassword, "secret-password"},
		{"SessionSecret", cfg.SessionSecret, "my-secret-key"},
		{"SessionExpiry", cfg.SessionExpiry, 48},
		{"ServerPort", cfg.ServerPort, 9000},
		{"LogLevel", cfg.LogLevel, "debug"},
		{"StockWarningThreshold", cfg.StockWarningThreshold, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestGetEnv_Default 测试环境变量未设置时返回默认值
func TestGetEnv_Default(t *testing.T) {
	// 确保测试环境变量未设置
	os.Unsetenv("TEST_GETENV_VAR")

	result := getEnv("TEST_GETENV_VAR", "default-value")

	if result != "default-value" {
		t.Errorf("getEnv() = %s, want %s", result, "default-value")
	}
}

// TestGetEnv_WithEnvVar 测试环境变量已设置时返回环境变量值
func TestGetEnv_WithEnvVar(t *testing.T) {
	t.Setenv("TEST_GETENV_VAR", "actual-value")

	result := getEnv("TEST_GETENV_VAR", "default-value")

	if result != "actual-value" {
		t.Errorf("getEnv() = %s, want %s", result, "actual-value")
	}
}

// TestGetEnvInt_ValidInt 测试环境变量为有效整数
func TestGetEnvInt_ValidInt(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		defaultVal int
		want       int
	}{
		{"valid positive", "TEST_PORT", "9000", 8080, 9000},
		{"valid zero", "TEST_PORT_ZERO", "0", 8080, 0},
		{"valid negative", "TEST_PORT_NEG", "-100", 8080, -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			result := getEnvInt(tt.key, tt.defaultVal)
			if result != tt.want {
				t.Errorf("getEnvInt() = %d, want %d", result, tt.want)
			}
		})
	}
}

// TestGetEnvInt_InvalidInt 测试环境变量为无效整数时返回默认值
func TestGetEnvInt_InvalidInt(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		defaultVal int
		want       int
	}{
		{"invalid string", "TEST_PORT", "abc", 8080, 8080},
		{"invalid float", "TEST_PORT", "80.5", 8080, 8080},
		{"invalid empty", "TEST_PORT", "", 8080, 8080},
		{"invalid with space", "TEST_PORT", " 8080 ", 8080, 8080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			result := getEnvInt(tt.key, tt.defaultVal)
			if result != tt.want {
				t.Errorf("getEnvInt() = %d, want %d", result, tt.want)
			}
		})
	}
}

// TestGetEnvInt_NotSet 测试环境变量未设置时返回默认值
func TestGetEnvInt_NotSet(t *testing.T) {
	os.Unsetenv("TEST_PORT_NOT_SET")

	result := getEnvInt("TEST_PORT_NOT_SET", 8080)

	if result != 8080 {
		t.Errorf("getEnvInt() = %d, want %d", result, 8080)
	}
}

// TestConfigLoad_InvalidDBPort 测试 DB_PORT 为无效值时使用默认值
func TestConfigLoad_InvalidDBPort(t *testing.T) {
	t.Setenv("DB_PORT", "invalid")

	cfg := Load()

	if cfg.DBPort != 3306 {
		t.Errorf("DBPort = %d, want %d (default)", cfg.DBPort, 3306)
	}
}

// TestConfigLoad_InvalidSessionExpiry 测试 SESSION_EXPIRY 为无效值时使用默认值
func TestConfigLoad_InvalidSessionExpiry(t *testing.T) {
	t.Setenv("SESSION_EXPIRY", "not-a-number")

	cfg := Load()

	if cfg.SessionExpiry != 24 {
		t.Errorf("SessionExpiry = %d, want %d (default)", cfg.SessionExpiry, 24)
	}
}

// TestConfigLoad_InvalidServerPort 测试 SERVER_PORT 为无效值时使用默认值
func TestConfigLoad_InvalidServerPort(t *testing.T) {
	t.Setenv("SERVER_PORT", "abc")

	cfg := Load()

	if cfg.ServerPort != 8080 {
		t.Errorf("ServerPort = %d, want %d (default)", cfg.ServerPort, 8080)
	}
}

// TestConfigLoad_InvalidStockWarningThreshold 测试 STOCK_WARNING_THRESHOLD 为无效值时使用默认值
func TestConfigLoad_InvalidStockWarningThreshold(t *testing.T) {
	t.Setenv("STOCK_WARNING_THRESHOLD", "xyz")

	cfg := Load()

	if cfg.StockWarningThreshold != 10 {
		t.Errorf("StockWarningThreshold = %d, want %d (default)", cfg.StockWarningThreshold, 10)
	}
}
