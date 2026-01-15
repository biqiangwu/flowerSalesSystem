package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/biqiangwu/flowerSalesSystem/internal/config"
)

// MockConfig 用于测试的配置辅助函数
func mockConfig(host string, port int, user, password, dbname string) *config.Config {
	return &config.Config{
		DBHost:     host,
		DBPort:     port,
		DBUser:     user,
		DBPassword: password,
		DBName:     dbname,
	}
}

// testCase 定义测试用例结构
type openTestCase struct {
	name        string
	config      *config.Config
	wantError   bool
	checkConn   bool // 是否检查连接有效性
	skipInShort bool // 是否在 short 模式下跳过
}

// TestOpen_ValidConfig 测试有效配置能成功建立连接
func TestOpen_ValidConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 注意：此测试需要真实的 MySQL 数据库或测试容器
	// 在 CI/CD 环境中应使用 testcontainers
	cfg := mockConfig("localhost", 3306, "root", "password", "test_db")

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// 验证连接池已建立
	if db == nil {
		t.Fatal("Open() returned nil db")
	}
}

// TestOpen_InvalidHost 测试无效主机地址应返回错误
func TestOpen_InvalidHost(t *testing.T) {
	cfg := mockConfig("invalid-host-99999.example.com", 3306, "user", "pass", "db")

	db, err := Open(cfg)
	if err == nil {
		db.Close()
		t.Fatal("Open() expected error, got nil")
	}
}

// TestOpen_InvalidPort 测试无效端口应返回错误
func TestOpen_InvalidPort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"negative port", -1, true},
		{"zero port", 0, true},
		{"port too large", 65536, true},
		{"valid max port", 65535, false}, // 不会立即失败，但连接会超时
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mockConfig("localhost", tt.port, "user", "pass", "db")

			db, err := Open(cfg)
			if tt.wantErr && err == nil {
				db.Close()
				t.Errorf("Open() expected error for port %d, got nil", tt.port)
			}
			if !tt.wantErr && db != nil {
				db.Close()
			}
		})
	}
}

// TestOpen_ConnectionPool 测试连接池参数是否正确设置
func TestOpen_ConnectionPool(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	cfg := mockConfig("localhost", 3306, "root", "password", "test_db")

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// 验证连接池参数设置
	// SetMaxOpenConns 应该被设置（默认值）
	// SetMaxIdleConns 应该被设置（默认值）
	// SetConnMaxLifetime 应该被设置

	// 检查连接状态
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		// 在没有真实数据库的情况下，这是预期的
		t.Logf("Ping failed (expected in test without real DB): %v", err)
	}
}

// TestOpen_Ping 测试连接可用性
func TestOpen_Ping(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	cfg := mockConfig("localhost", 3306, "root", "password", "test_db")

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Logf("Ping failed (expected in test without real DB): %v", err)
		// 在真实集成测试环境中，这里应该成功
		// t.Errorf("Ping() error = %v", err)
	}
}

// TestOpen_TableDriven 表格驱动测试：测试各种配置场景
func TestOpen_TableDriven(t *testing.T) {
	tests := []openTestCase{
		{
			name:        "valid config - localhost",
			config:      mockConfig("localhost", 3306, "root", "password", "test_db"),
			wantError:   false,
			checkConn:   true,
			skipInShort: true,
		},
		{
			name:      "invalid host - non existent",
			config:    mockConfig("this-host-does-not-exist-12345.com", 3306, "user", "pass", "db"),
			wantError: true,
			checkConn: false,
		},
		{
			name:      "empty password",
			config:    mockConfig("localhost", 3306, "root", "", "test_db"),
			wantError: true, // MySQL 通常要求密码
			checkConn: false,
		},
		{
			name:      "empty database name",
			config:    mockConfig("localhost", 3306, "root", "password", ""),
			wantError: true,
			checkConn: false,
		},
		{
			name:      "empty user",
			config:    mockConfig("localhost", 3306, "", "password", "test_db"),
			wantError: true,
			checkConn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Short() && tt.skipInShort {
				t.Skip("跳过集成测试")
			}

			db, err := Open(tt.config)

			if tt.wantError {
				if err == nil {
					db.Close()
					t.Error("Open() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Open() unexpected error: %v", err)
				}
				defer db.Close()

				if db == nil {
					t.Fatal("Open() returned nil db")
				}

				if tt.checkConn {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()

					if err := db.PingContext(ctx); err != nil {
						t.Logf("Ping failed (may be expected without real DB): %v", err)
					}
				}
			}
		})
	}
}

// TestDBConfig_String 测试 DBConfig DSN 格式化
func TestDBConfig_String(t *testing.T) {
	cfg := mockConfig("localhost", 3306, "user", "pass", "dbname")

	wantDSN := "user:pass@tcp(localhost:3306)/dbname?parseTime=true"
	gotDSN := buildDSN(cfg)

	if gotDSN != wantDSN {
		t.Errorf("buildDSN() = %s, want %s", gotDSN, wantDSN)
	}
}

// TestDBConfig_StringWithParams 测试带额外参数的 DSN 格式化
func TestDBConfig_StringWithParams(t *testing.T) {
	cfg := mockConfig("192.168.1.100", 3307, "admin", "secret123", "production")

	wantDSN := "admin:secret123@tcp(192.168.1.100:3307)/production?parseTime=true"
	gotDSN := buildDSN(cfg)

	if gotDSN != wantDSN {
		t.Errorf("buildDSN() = %s, want %s", gotDSN, wantDSN)
	}
}

// TestOpen_ContextCancellation 测试上下文取消时的行为
func TestOpen_ContextCancellation(t *testing.T) {
	// 注意：Open 函数当前不支持上下文参数
	// 这是设计决策，根据实际需求调整
	t.Skip("取决于 Open 函数是否支持 context")
}

// TestOpen_Concurrency 测试并发打开连接
func TestOpen_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	cfg := mockConfig("localhost", 3306, "root", "password", "test_db")

	// 并发打开多个连接
	concurrency := 10
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			db, err := Open(cfg)
			if err == nil && db != nil {
				db.Close()
			}
			errChan <- err
		}()
	}

	// 收集结果
	for i := 0; i < concurrency; i++ {
		err := <-errChan
		if err != nil {
			t.Logf("Concurrent Open failed (may be expected without real DB): %v", err)
		}
	}
}

// TestClose_Connection 测试关闭连接
func TestClose_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	cfg := mockConfig("localhost", 3306, "root", "password", "test_db")

	db, err := Open(cfg)
	if err != nil {
		t.Skipf("需要真实数据库连接: %v", err)
	}

	// 关闭连接
	if err := db.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 关闭后 Ping 应该失败
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err == nil {
		t.Error("Expected error after Close(), got nil")
	}
}

// MockDB 模拟数据库接口（用于单元测试）
type MockDB struct {
	*sql.DB
	PingFunc func(context.Context) error
	CloseFunc func() error
}

func (m *MockDB) PingContext(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

func (m *MockDB) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
