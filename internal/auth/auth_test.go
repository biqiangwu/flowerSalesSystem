package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/biqiangwu/flowerSalesSystem/internal/user"
	_ "github.com/mattn/go-sqlite3" // SQLite 驱动用于测试
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

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

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestHashPassword 测试密码哈希
func TestHashPassword(t *testing.T) {
	authSvc := NewAuthService(nil, nil)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid password",
			input:   "password123",
			wantErr: false,
		},
		{
			name:    "empty password",
			input:   "",
			wantErr: true,
		},
		{
			name:    "short password",
			input:   "123",
			wantErr: true,
		},
		{
			name:    "long password",
			input:   "this_is_a_very_long_password_that_should_work_fine",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := authSvc.HashPassword(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if hash == "" {
					t.Error("HashPassword() hash is empty")
				}
				if hash == tt.input {
					t.Error("HashPassword() hash equals input password")
				}
			}
		})
	}
}

// TestVerifyPassword 测试密码验证
func TestVerifyPassword(t *testing.T) {
	authSvc := NewAuthService(nil, nil)
	validPassword := "testPassword123"
	hash, _ := authSvc.HashPassword(validPassword)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: validPassword,
			hash:     hash,
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongPassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty hash",
			password: validPassword,
			hash:     "",
			want:     false,
		},
		{
			name:     "invalid hash format",
			password: validPassword,
			hash:     "invalid_hash",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := authSvc.VerifyPassword(tt.password, tt.hash); got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRegister 测试用户注册
func TestRegister(t *testing.T) {
	db := setupTestDB(t)
	userRepo := user.NewMySQLUserRepository(db)
	sessionMgr := NewMemorySessionManager()
	authSvc := NewAuthService(userRepo, sessionMgr)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "register new customer",
			username: "newcustomer",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "register new admin",
			username: "newadmin",
			password: "admin123",
			wantErr:  false,
		},
		{
			name:     "register with empty username",
			username: "",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "register with empty password",
			username: "testuser",
			password: "",
			wantErr:  true,
		},
		{
			name:     "register with short password",
			username: "shortpwd",
			password: "123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			u, err := authSvc.Register(ctx, tt.username, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if u == nil {
					t.Error("Register() user is nil")
					return
				}
				if u.Username != tt.username {
					t.Errorf("Register() username = %q, want %q", u.Username, tt.username)
				}
				if u.ID <= 0 {
					t.Errorf("Register() ID = %d, want > 0", u.ID)
				}
				if u.Role != user.RoleCustomer {
					t.Errorf("Register() role = %q, want %q", u.Role, user.RoleCustomer)
				}
			}
		})
	}
}

// TestRegister_DuplicateUsername 测试注册重复用户名
func TestRegister_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	userRepo := user.NewMySQLUserRepository(db)
	sessionMgr := NewMemorySessionManager()
	authSvc := NewAuthService(userRepo, sessionMgr)
	ctx := context.Background()

	// 注册第一个用户
	_, err := authSvc.Register(ctx, "duplicate", "password123")
	if err != nil {
		t.Fatalf("failed to register first user: %v", err)
	}

	// 尝试注册相同用户名
	_, err = authSvc.Register(ctx, "duplicate", "differentpassword")
	if err == nil {
		t.Error("Register() expected error for duplicate username, got nil")
	}
}

// TestLogin 测试用户登录
func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	userRepo := user.NewMySQLUserRepository(db)
	sessionMgr := NewMemorySessionManager()
	authSvc := NewAuthService(userRepo, sessionMgr)
	ctx := context.Background()

	// 先注册一个测试用户
	testUsername := "loginuser"
	testPassword := "testpass123"
	_, err := authSvc.Register(ctx, testUsername, testPassword)
	if err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "login with correct credentials",
			username: testUsername,
			password: testPassword,
			wantErr:  false,
		},
		{
			name:     "login with wrong password",
			username: testUsername,
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "login with non-existing user",
			username: "nonexistent",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "login with empty username",
			username: "",
			password: testPassword,
			wantErr:  true,
		},
		{
			name:     "login with empty password",
			username: testUsername,
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := authSvc.Login(ctx, tt.username, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if session == nil {
					t.Error("Login() session is nil")
					return
				}
				if session.Token == "" {
					t.Error("Login() session token is empty")
				}
				if session.UserID <= 0 {
					t.Errorf("Login() session UserID = %d, want > 0", session.UserID)
				}
				if session.Username != tt.username {
					t.Errorf("Login() session username = %q, want %q", session.Username, tt.username)
				}
				if session.ExpiresAt.Before(time.Now()) {
					t.Error("Login() session already expired")
				}
			}
		})
	}
}

// TestValidateSession 测试 Session 验证
func TestValidateSession(t *testing.T) {
	db := setupTestDB(t)
	userRepo := user.NewMySQLUserRepository(db)
	sessionMgr := NewMemorySessionManager()
	authSvc := NewAuthService(userRepo, sessionMgr)
	ctx := context.Background()

	// 注册并登录用户
	testUsername := "sessionuser"
	testPassword := "sessionpass123"
	_, err := authSvc.Register(ctx, testUsername, testPassword)
	if err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}

	session, err := authSvc.Login(ctx, testUsername, testPassword)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	tests := []struct {
		name     string
		token    string
		wantErr  bool
		checkUser func(*testing.T, *user.User)
	}{
		{
			name:    "validate valid session",
			token:   session.Token,
			wantErr: false,
			checkUser: func(t *testing.T, u *user.User) {
				if u.Username != testUsername {
					t.Errorf("ValidateSession() username = %q, want %q", u.Username, testUsername)
				}
			},
		},
		{
			name:    "validate invalid session",
			token:   "invalid_token_12345",
			wantErr: true,
		},
		{
			name:    "validate empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := authSvc.ValidateSession(ctx, tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if u == nil {
					t.Error("ValidateSession() user is nil")
					return
				}
				if tt.checkUser != nil {
					tt.checkUser(t, u)
				}
			}
		})
	}
}

// TestLogout 测试用户登出
func TestLogout(t *testing.T) {
	db := setupTestDB(t)
	userRepo := user.NewMySQLUserRepository(db)
	sessionMgr := NewMemorySessionManager()
	authSvc := NewAuthService(userRepo, sessionMgr)
	ctx := context.Background()

	// 注册并登录用户
	testUsername := "logoutuser"
	testPassword := "logoutpass123"
	_, err := authSvc.Register(ctx, testUsername, testPassword)
	if err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}

	session, err := authSvc.Login(ctx, testUsername, testPassword)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	// 登出
	err = authSvc.Logout(ctx, session.Token)
	if err != nil {
		t.Errorf("Logout() error = %v", err)
		return
	}

	// 验证 session 已失效
	_, err = authSvc.ValidateSession(ctx, session.Token)
	if err == nil {
		t.Error("Logout() session still valid after logout")
	}
}
