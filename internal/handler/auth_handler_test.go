package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/biqiangwu/flowerSalesSystem/internal/auth"
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

// setupTestHandler 创建测试用的 Handler
func setupTestHandler(t *testing.T) *Handler {
	t.Helper()

	db := setupTestDB(t)
	userRepo := user.NewMySQLUserRepository(db)
	sessionMgr := auth.NewMemorySessionManager()
	authSvc := auth.NewAuthService(userRepo, sessionMgr)

	return NewHandler(authSvc, nil)
}

// TestHandleRegister 测试注册接口
func TestHandleRegister(t *testing.T) {
	handler := setupTestHandler(t)

	tests := []struct {
		name       string
		request    RegisterRequest
		wantStatus int
		checkBody  func(*testing.T, []byte)
	}{
		{
			name: "register new user successfully",
			request: RegisterRequest{
				Username: "newuser",
				Password: "password123",
			},
			wantStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				user, ok := resp["user"].(map[string]interface{})
				if !ok {
					t.Fatal("response missing user object")
				}
				if user["id"] == nil {
					t.Error("response missing id")
				}
				if user["username"] != "newuser" {
					t.Errorf("username = %v, want newuser", user["username"])
				}
			},
		},
		{
			name: "register with empty username",
			request: RegisterRequest{
				Username: "",
				Password: "password123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "register with empty password",
			request: RegisterRequest{
				Username: "testuser",
				Password: "",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "register with short password",
			request: RegisterRequest{
				Username: "shortpwd",
				Password: "123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "register with invalid json",
			request: RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.name != "register with invalid json" {
				var err error
				body, err = json.Marshal(tt.request)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleRegister(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleRegister() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.Bytes())
			}
		})
	}
}

// TestHandleRegister_DuplicateUsername 测试注册重复用户名
func TestHandleRegister_DuplicateUsername(t *testing.T) {
	handler := setupTestHandler(t)

	// 注册第一个用户
	reqBody := RegisterRequest{
		Username: "duplicate",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.HandleRegister(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first registration failed: status = %d", w.Code)
	}

	// 尝试注册相同用户名
	req = httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.HandleRegister(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("HandleRegister() duplicate username status = %d, want %d", w.Code, http.StatusConflict)
	}
}

// TestHandleLogin 测试登录接口
func TestHandleLogin(t *testing.T) {
	handler := setupTestHandler(t)

	// 先注册一个用户
	registerReq := RegisterRequest{
		Username: "loginuser",
		Password: "loginpass123",
	}
	body, _ := json.Marshal(registerReq)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.HandleRegister(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("registration failed: status = %d", w.Code)
	}

	tests := []struct {
		name       string
		request    LoginRequest
		wantStatus int
		checkCookie func(*testing.T, []*http.Cookie)
	}{
		{
			name: "login with correct credentials",
			request: LoginRequest{
				Username: "loginuser",
				Password: "loginpass123",
			},
			wantStatus: http.StatusOK,
			checkCookie: func(t *testing.T, cookies []*http.Cookie) {
				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "session_token" {
						sessionCookie = c
						break
					}
				}
				if sessionCookie == nil {
					t.Error("login response missing session_token cookie")
				}
				if sessionCookie != nil && !sessionCookie.HttpOnly {
					t.Error("session_token cookie should be HttpOnly")
				}
			},
		},
		{
			name: "login with wrong password",
			request: LoginRequest{
				Username: "loginuser",
				Password: "wrongpassword",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "login with non-existing user",
			request: LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "login with empty username",
			request: LoginRequest{
				Username: "",
				Password: "password123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "login with empty password",
			request: LoginRequest{
				Username: "loginuser",
				Password: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)

			req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleLogin(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleLogin() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.checkCookie != nil {
				tt.checkCookie(t, w.Result().Cookies())
			}
		})
	}
}

// TestHandleLogout 测试登出接口
func TestHandleLogout(t *testing.T) {
	handler := setupTestHandler(t)

	// 先注册并登录用户
	registerReq := RegisterRequest{
		Username: "logoutuser",
		Password: "logoutpass123",
	}
	body, _ := json.Marshal(registerReq)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.HandleRegister(w, req)

	loginReq := LoginRequest{
		Username: "logoutuser",
		Password: "logoutpass123",
	}
	body, _ = json.Marshal(loginReq)

	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.HandleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: status = %d", w.Code)
	}

	// 获取 session token
	var sessionToken string
	for _, c := range w.Result().Cookies() {
		if c.Name == "session_token" {
			sessionToken = c.Value
			break
		}
	}

	if sessionToken == "" {
		t.Fatal("login failed to set session_token cookie")
	}

	// 登出
	req = httptest.NewRequest("POST", "/api/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	w = httptest.NewRecorder()
	handler.HandleLogout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleLogout() status = %d, want %d", w.Code, http.StatusOK)
	}

	// 验证登出后 cookie 被清除
	cookies := w.Result().Cookies()
	var foundClearedCookie bool
	for _, c := range cookies {
		if c.Name == "session_token" {
			if c.Value == "" || c.MaxAge < 0 {
				foundClearedCookie = true
			}
		}
	}
	if !foundClearedCookie {
		t.Error("HandleLogout() session_token cookie not cleared")
	}
}

// TestHandleLogout_MissingCookie 测试登出时缺少 cookie
func TestHandleLogout_MissingCookie(t *testing.T) {
	handler := setupTestHandler(t)

	req := httptest.NewRequest("POST", "/api/logout", nil)
	w := httptest.NewRecorder()
	handler.HandleLogout(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HandleLogout() missing cookie status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestRespondJSON 测试 JSON 响应辅助函数
func TestRespondJSON(t *testing.T) {
	handler := NewHandler(nil, nil)

	testData := map[string]string{
		"message": "success",
		"status":  "ok",
	}

	w := httptest.NewRecorder()
	handler.respondJSON(w, http.StatusOK, testData)

	if w.Code != http.StatusOK {
		t.Errorf("respondJSON() status = %d, want %d", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("respondJSON() Content-Type = %q, want application/json", contentType)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["message"] != "success" {
		t.Errorf("respondJSON() message = %q, want success", resp["message"])
	}
}

// TestRespondError 测试错误响应辅助函数
func TestRespondError(t *testing.T) {
	handler := NewHandler(nil, nil)

	w := httptest.NewRecorder()
	handler.respondError(w, http.StatusBadRequest, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("respondError() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error != "invalid input" {
		t.Errorf("respondError() error = %q, want invalid input", resp.Error)
	}
}
