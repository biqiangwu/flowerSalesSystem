package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/biqiangwu/flowerSalesSystem/internal/auth"
	"github.com/biqiangwu/flowerSalesSystem/internal/user"
	_ "github.com/mattn/go-sqlite3" // SQLite 驱动用于测试
)

// testContext 包含测试所需的上下文
type testContext struct {
	db          *sql.DB
	handler     *Handler
	userRepo    user.UserRepository
	sessionMgr  auth.SessionManager
	authSvc     auth.AuthService
}

// setupUserTestHandler 创建测试用的用户管理 Handler
func setupUserTestHandler(t *testing.T) (*testContext, *sql.DB) {
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

	userRepo := user.NewMySQLUserRepository(db)
	sessionMgr := auth.NewMemorySessionManager()
	authSvc := auth.NewAuthService(userRepo, sessionMgr)
	userSvc := user.NewUserService(userRepo)

	handler := &Handler{
		authService: authSvc,
		userService: userSvc,
	}

	ctx := &testContext{
		db:         db,
		handler:    handler,
		userRepo:   userRepo,
		sessionMgr: sessionMgr,
		authSvc:    authSvc,
	}

	return ctx, db
}

// createTestUser 创建测试用户并返回其信息和 session
func createTestUserWithSession(t *testing.T, ctx *testContext, username string, role user.Role) (*user.User, string) {
	t.Helper()

	context := context.Background()

	// 使用 Register 创建用户（会自动哈希密码）
	password := "password123"
	registeredUser, err := ctx.authSvc.Register(context, username, password)
	if err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}

	// 更新用户角色（如果不是 customer）
	if role != user.RoleCustomer {
		// 需要直接更新数据库中的角色
		_, err = ctx.db.ExecContext(context, "UPDATE users SET role = ? WHERE id = ?", role, registeredUser.ID)
		if err != nil {
			t.Fatalf("failed to update user role: %v", err)
		}
		registeredUser.Role = role
	}

	// 登录获取 session
	session, err := ctx.authSvc.Login(context, username, password)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	return registeredUser, session.Token
}

// TestHandleListUsers 测试获取用户列表接口
func TestHandleListUsers(t *testing.T) {
	ctx, _ := setupUserTestHandler(t)
	handler := ctx.handler

	// 创建测试用户
	for i := 1; i <= 5; i++ {
		_, _ = createTestUserWithSession(t, ctx, "listuser"+strconv.Itoa(i), user.RoleCustomer)
	}

	// 创建管理员 session
	_, adminSession := createTestUserWithSession(t, ctx, "admin", user.RoleAdmin)

	tests := []struct {
		name       string
		session    string
		query      url.Values
		wantStatus int
		minCount   int
		maxCount   int
	}{
		{
			name:       "list all users as admin",
			session:    adminSession,
			query:      url.Values{},
			wantStatus: http.StatusOK,
			minCount:   6, // 5 customers + 1 admin
			maxCount:   6,
		},
		{
			name:       "list with pagination",
			session:    adminSession,
			query:      url.Values{"page": []string{"1"}, "page_size": []string{"3"}},
			wantStatus: http.StatusOK,
			minCount:   1,
			maxCount:   3,
		},
		{
			name:       "list without session",
			session:    "",
			query:      url.Values{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "list with invalid session",
			session:    "invalid_token",
			query:      url.Values{},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/users?"+tt.query.Encode(), nil)
			if tt.session != "" {
				req.AddCookie(&http.Cookie{Name: "session_token", Value: tt.session})
			}
			w := httptest.NewRecorder()

			handler.HandleListUsers(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleListUsers() status = %d, want %d", w.Code, tt.wantStatus)
				return
			}

			if tt.wantStatus == http.StatusOK && (tt.minCount > 0 || tt.maxCount > 0) {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}

				users, ok := resp["users"].([]interface{})
				if !ok {
					t.Fatal("response missing users array")
				}

				count := len(users)
				if count < tt.minCount || count > tt.maxCount {
					t.Errorf("HandleListUsers() count = %d, want between %d and %d", count, tt.minCount, tt.maxCount)
				}
			}
		})
	}
}

// TestHandleDeleteUser 测试删除用户接口
func TestHandleDeleteUser(t *testing.T) {
	ctx, _ := setupUserTestHandler(t)
	handler := ctx.handler

	// 创建操作用户
	adminUser, adminSession := createTestUserWithSession(t, ctx, "admin", user.RoleAdmin)
	clerkUser, clerkSession := createTestUserWithSession(t, ctx, "clerk", user.RoleClerk)
	_, customerSession := createTestUserWithSession(t, ctx, "customer", user.RoleCustomer)

	tests := []struct {
		name       string
		userID     string
		session    string
		wantStatus int
	}{
		{
			name:       "admin deletes user",
			userID:     "", // 将在测试中设置
			session:    adminSession,
			wantStatus: http.StatusOK,
		},
		{
			name:       "clerk deletes user",
			userID:     "", // 将在测试中设置
			session:    clerkSession,
			wantStatus: http.StatusOK,
		},
		{
			name:       "customer cannot delete user",
			userID:     strconv.Itoa(adminUser.ID),
			session:    customerSession,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "delete without session",
			userID:     strconv.Itoa(clerkUser.ID),
			session:    "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "delete non-existing user",
			userID:     "99999",
			session:    adminSession,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "delete with invalid user id",
			userID:     "invalid",
			session:    adminSession,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为需要创建目标用户的测试用例创建用户
			if tt.userID == "" {
				targetUser, _ := createTestUserWithSession(t, ctx, "target_"+tt.name, user.RoleCustomer)
				tt.userID = strconv.Itoa(targetUser.ID)
			}

			req := httptest.NewRequest("DELETE", "/api/users/"+tt.userID, nil)
			if tt.session != "" {
				req.AddCookie(&http.Cookie{Name: "session_token", Value: tt.session})
			}
			w := httptest.NewRecorder()

			handler.HandleDeleteUser(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleDeleteUser() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// TestHandleResetPassword 测试重置用户密码接口
func TestHandleResetPassword(t *testing.T) {
	ctx, _ := setupUserTestHandler(t)
	handler := ctx.handler

	// 创建测试用户
	targetUser, _ := createTestUserWithSession(t, ctx, "targetuser", user.RoleCustomer)
	adminUser, adminSession := createTestUserWithSession(t, ctx, "admin", user.RoleAdmin)
	clerkUser, clerkSession := createTestUserWithSession(t, ctx, "clerk", user.RoleClerk)
	_, customerSession := createTestUserWithSession(t, ctx, "customer", user.RoleCustomer)

	tests := []struct {
		name       string
		userID     string
		request    ResetPasswordRequest
		session    string
		wantStatus int
	}{
		{
			name:   "admin resets password",
			userID: strconv.Itoa(targetUser.ID),
			request: ResetPasswordRequest{
				NewPassword: "newpassword123",
			},
			session:    adminSession,
			wantStatus: http.StatusOK,
		},
		{
			name:   "clerk resets password",
			userID: strconv.Itoa(targetUser.ID),
			request: ResetPasswordRequest{
				NewPassword: "anotherpassword456",
			},
			session:    clerkSession,
			wantStatus: http.StatusOK,
		},
		{
			name:   "customer cannot reset password",
			userID: strconv.Itoa(adminUser.ID),
			request: ResetPasswordRequest{
				NewPassword: "hackedpassword",
			},
			session:    customerSession,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "reset without session",
			userID: strconv.Itoa(clerkUser.ID),
			request: ResetPasswordRequest{
				NewPassword: "newpass123",
			},
			session:    "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "reset with empty password",
			userID: strconv.Itoa(targetUser.ID),
			request: ResetPasswordRequest{
				NewPassword: "",
			},
			session:    adminSession,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "reset with short password",
			userID: strconv.Itoa(targetUser.ID),
			request: ResetPasswordRequest{
				NewPassword: "123",
			},
			session:    adminSession,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "reset non-existing user",
			userID: "99999",
			request: ResetPasswordRequest{
				NewPassword: "newpass123",
			},
			session:    adminSession,
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "reset with invalid user id",
			userID: "invalid",
			request: ResetPasswordRequest{
				NewPassword: "newpass123",
			},
			session:    adminSession,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)

			req := httptest.NewRequest("POST", "/api/users/"+tt.userID+"/reset-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.session != "" {
				req.AddCookie(&http.Cookie{Name: "session_token", Value: tt.session})
			}
			w := httptest.NewRecorder()

			handler.HandleResetPassword(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleResetPassword() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
