package handler

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/biqiangwu/flowerSalesSystem/internal/address"
	"github.com/biqiangwu/flowerSalesSystem/internal/user"
	_ "github.com/mattn/go-sqlite3"
)

// setupAddressTestDB 创建地址测试数据库
func setupAddressTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'customer',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS addresses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		label TEXT,
		address TEXT NOT NULL,
		contact TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	if _, err := db.Exec(createTablesSQL); err != nil {
		db.Close()
		t.Fatalf("failed to create tables: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// setupAddressTestHandler 创建测试用的 Handler
func setupAddressTestHandler(t *testing.T) (*Handler, *user.User) {
	t.Helper()

	db := setupAddressTestDB(t)

	userRepo := user.NewMySQLUserRepository(db)
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	testUser := &user.User{
		Username:     "testuser",
		PasswordHash: hashedPassword,
		Role:         user.RoleCustomer,
	}
	if err := userRepo.Create(t.Context(), testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	addressRepo := address.NewAddressRepository(db)
	addressSvc := address.NewAddressService(addressRepo)

	h := &Handler{
		addressService: addressSvc,
		userRepo:       userRepo,
	}

	return h, testUser
}

// TestHandleListAddresses 测试获取地址列表
func TestHandleListAddresses(t *testing.T) {
	handler, testUser := setupAddressTestHandler(t)

	ctx := t.Context()
	for i := 1; i <= 2; i++ {
		req := &address.CreateAddressRequest{
			Address: fmt.Sprintf("测试地址%d号", i),
			Contact: "13800138000",
		}
		if err := handler.addressService.CreateAddress(ctx, testUser.ID, req); err != nil {
			t.Fatalf("failed to create test address: %v", err)
		}
	}

	addresses, err := handler.addressService.ListAddresses(ctx, testUser.ID)
	if err != nil {
		t.Fatalf("failed to list addresses: %v", err)
	}

	if len(addresses) != 2 {
		t.Errorf("got %d addresses, want 2", len(addresses))
	}
}
