package user

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
)

// setupTestDB 创建测试数据库连接
// 使用环境变量配置测试数据库
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// 使用 sqlite 作为测试数据库（简单快速）
	// 在生产环境中应该使用 MySQL
	// 这里为了测试独立性，使用内存 sqlite
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// 创建测试表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'customer',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		t.Fatalf("failed to create users table: %v", err)
	}

	// 清理函数
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestMySQLUserRepository_Create 测试 Create 方法
func TestMySQLUserRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewMySQLUserRepository(db)

	tests := []struct {
		name    string
		user    *User
		wantErr bool
		errMsg  string
	}{
		{
			name: "create valid user",
			user: &User{
				Username:     "testuser1",
				PasswordHash: "hash123",
				Role:         RoleCustomer,
			},
			wantErr: false,
		},
		{
			name: "create admin user",
			user: &User{
				Username:     "admin1",
				PasswordHash: "adminhash",
				Role:         RoleAdmin,
			},
			wantErr: false,
		},
		{
			name: "create clerk user",
			user: &User{
				Username:     "clerk1",
				PasswordHash: "clerkhash",
				Role:         RoleClerk,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := repo.Create(ctx, tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.user.ID <= 0 {
				t.Errorf("Create() user ID = %d, want > 0", tt.user.ID)
			}

			if !tt.wantErr && tt.user.CreatedAt.IsZero() {
				t.Errorf("Create() CreatedAt is zero")
			}

			if !tt.wantErr && tt.user.UpdatedAt.IsZero() {
				t.Errorf("Create() UpdatedAt is zero")
			}
		})
	}
}

// TestMySQLUserRepository_Create_DuplicateUsername 测试重复用户名
func TestMySQLUserRepository_Create_DuplicateUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewMySQLUserRepository(db)
	ctx := context.Background()

	// 创建第一个用户
	user1 := &User{
		Username:     "duplicate",
		PasswordHash: "hash123",
		Role:         RoleCustomer,
	}
	if err := repo.Create(ctx, user1); err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	// 尝试创建相同用户名的用户
	user2 := &User{
		Username:     "duplicate",
		PasswordHash: "hash456",
		Role:         RoleCustomer,
	}
	err := repo.Create(ctx, user2)
	if err == nil {
		t.Error("Create() expected error for duplicate username, got nil")
	}
}

// TestMySQLUserRepository_GetByID 测试 GetByID 方法
func TestMySQLUserRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewMySQLUserRepository(db)
	ctx := context.Background()

	// 创建测试用户
	createdUser := &User{
		Username:     "getbyid",
		PasswordHash: "hash123",
		Role:         RoleAdmin,
	}
	if err := repo.Create(ctx, createdUser); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tests := []struct {
		name    string
		userID  int
		wantErr bool
	}{
		{
			name:    "get existing user",
			userID:  createdUser.ID,
			wantErr: false,
		},
		{
			name:    "get non-existing user",
			userID:  99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByID(ctx, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if user == nil {
					t.Error("GetByID() user is nil")
					return
				}
				if user.ID != tt.userID {
					t.Errorf("GetByID() ID = %d, want %d", user.ID, tt.userID)
				}
				if user.Username != "getbyid" {
					t.Errorf("GetByID() Username = %q, want %q", user.Username, "getbyid")
				}
			}
		})
	}
}

// TestMySQLUserRepository_GetByUsername 测试 GetByUsername 方法
func TestMySQLUserRepository_GetByUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewMySQLUserRepository(db)
	ctx := context.Background()

	// 创建测试用户
	createdUser := &User{
		Username:     "getbyuser",
		PasswordHash: "hash123",
		Role:         RoleClerk,
	}
	if err := repo.Create(ctx, createdUser); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "get existing user by username",
			username: "getbyuser",
			wantErr:  false,
		},
		{
			name:     "get non-existing user by username",
			username: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByUsername(ctx, tt.username)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if user == nil {
					t.Error("GetByUsername() user is nil")
					return
				}
				if user.Username != tt.username {
					t.Errorf("GetByUsername() Username = %q, want %q", user.Username, tt.username)
				}
			}
		})
	}
}

// TestMySQLUserRepository_List 测试 List 方法
func TestMySQLUserRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewMySQLUserRepository(db)
	ctx := context.Background()

	// 创建多个测试用户
	for i := 1; i <= 5; i++ {
		user := &User{
			Username:     fmt.Sprintf("listuser%d", i),
			PasswordHash: "hash123",
			Role:         RoleCustomer,
		}
		if err := repo.Create(ctx, user); err != nil {
			t.Fatalf("failed to create user %d: %v", i, err)
		}
	}

	tests := []struct {
		name       string
		page       int
		pageSize   int
		minCount   int
		maxCount   int
		wantErr    bool
	}{
		{
			name:     "list first page",
			page:     1,
			pageSize: 3,
			minCount: 1,
			maxCount: 3,
			wantErr:  false,
		},
		{
			name:     "list second page",
			page:     2,
			pageSize: 3,
			minCount: 1,
			maxCount: 2,
			wantErr:  false,
		},
		{
			name:     "list empty page",
			page:     100,
			pageSize: 10,
			minCount: 0,
			maxCount: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := repo.List(ctx, tt.page, tt.pageSize)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(users) < tt.minCount || len(users) > tt.maxCount {
					t.Errorf("List() count = %d, want between %d and %d", len(users), tt.minCount, tt.maxCount)
				}
			}
		})
	}
}

// TestMySQLUserRepository_Delete 测试 Delete 方法
func TestMySQLUserRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewMySQLUserRepository(db)
	ctx := context.Background()

	// 创建测试用户
	createdUser := &User{
		Username:     "deleteuser",
		PasswordHash: "hash123",
		Role:         RoleCustomer,
	}
	if err := repo.Create(ctx, createdUser); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tests := []struct {
		name    string
		userID  int
		wantErr bool
	}{
		{
			name:    "delete existing user",
			userID:  createdUser.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existing user",
			userID:  99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 验证删除后用户不存在
			if !tt.wantErr {
				_, err := repo.GetByID(ctx, tt.userID)
				if err == nil {
					t.Error("Delete() user still exists after deletion")
				}
			}
		})
	}
}

// TestMySQLUserRepository_UpdatePassword 测试 UpdatePassword 方法
func TestMySQLUserRepository_UpdatePassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewMySQLUserRepository(db)
	ctx := context.Background()

	// 创建测试用户
	createdUser := &User{
		Username:     "updatepwd",
		PasswordHash: "oldhash",
		Role:         RoleCustomer,
	}
	if err := repo.Create(ctx, createdUser); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	newHash := "newhash"
	err := repo.UpdatePassword(ctx, createdUser.ID, newHash)
	if err != nil {
		t.Errorf("UpdatePassword() error = %v", err)
		return
	}

	// 验证密码已更新
	user, err := repo.GetByID(ctx, createdUser.ID)
	if err != nil {
		t.Errorf("GetByID() after UpdatePassword error = %v", err)
		return
	}

	if user.PasswordHash != newHash {
		t.Errorf("UpdatePassword() PasswordHash = %q, want %q", user.PasswordHash, newHash)
	}
}
