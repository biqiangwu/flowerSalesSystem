package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/biqiangwu/flowerSalesSystem/internal/auth"
	"github.com/biqiangwu/flowerSalesSystem/internal/user"
)

// mockAuthService 是 AuthService 的模拟实现，用于测试
type mockAuthService struct {
	validToken   string
	validUser    *user.User
	validateErr  error
}

func (m *mockAuthService) Register(ctx context.Context, username, password string) (*user.User, error) {
	return nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, username, password string) (*auth.Session, error) {
	return nil, nil
}

func (m *mockAuthService) Logout(ctx context.Context, sessionToken string) error {
	return nil
}

func (m *mockAuthService) ValidateSession(ctx context.Context, sessionToken string) (*user.User, error) {
	if m.validateErr != nil {
		return nil, m.validateErr
	}
	if sessionToken == m.validToken {
		return m.validUser, nil
	}
	return nil, &authError{"invalid session token"}
}

func (m *mockAuthService) HashPassword(password string) (string, error) {
	return "", nil
}

func (m *mockAuthService) VerifyPassword(password, hash string) bool {
	return true
}

// authError 用于模拟错误
type authError struct {
	msg string
}

func (e *authError) Error() string {
	return e.msg
}

// TestAuthMiddleware 测试认证中间件
func TestAuthMiddleware(t *testing.T) {
	// 创建测试用户
	testUser := &user.User{
		ID:       1,
		Username: "testuser",
		Role:     user.RoleCustomer,
	}

	tests := []struct {
		name           string
		validToken     string
		validateErr    error
		cookie         *http.Cookie
		wantStatus     int
		wantUserInCtx  bool
	}{
		{
			name:       "有效 Session - 调用 next handler，用户信息注入上下文",
			validToken: "valid-token-123",
			validateErr: nil,
			cookie: &http.Cookie{
				Name:  "session_token",
				Value: "valid-token-123",
			},
			wantStatus:    http.StatusOK,
			wantUserInCtx: true,
		},
		{
			name:       "无效 Session - 返回 401",
			validToken: "valid-token-123",
			validateErr: nil,
			cookie: &http.Cookie{
				Name:  "session_token",
				Value: "invalid-token",
			},
			wantStatus:    http.StatusUnauthorized,
			wantUserInCtx: false,
		},
		{
			name:       "缺失 Cookie - 返回 401",
			validToken: "valid-token-123",
			validateErr: nil,
			cookie:     nil,
			wantStatus:    http.StatusUnauthorized,
			wantUserInCtx: false,
		},
		{
			name:       "过期 Session - 返回 401",
			validToken: "valid-token-123",
			validateErr: &authError{"session expired"},
			cookie: &http.Cookie{
				Name:  "session_token",
				Value: "valid-token-123",
			},
			wantStatus:    http.StatusUnauthorized,
			wantUserInCtx: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 mock AuthService
			mockAuth := &mockAuthService{
				validToken:  tt.validToken,
				validUser:   testUser,
				validateErr: tt.validateErr,
			}

			// 创建一个测试 handler，检查上下文中的用户信息
			nextHandlerCalled := false
			var ctxUser *user.User
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextHandlerCalled = true
				// 尝试从上下文获取用户信息
				if u := r.Context().Value(userKey); u != nil {
					ctxUser = u.(*user.User)
				}
				w.WriteHeader(http.StatusOK)
			})

			// 应用中间件
			handler := AuthMiddleware(mockAuth, nextHandler)

			// 创建请求
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			// 创建响应记录器
			rr := httptest.NewRecorder()

			// 执行请求
			handler.ServeHTTP(rr, req)

			// 检查状态码
			if rr.Code != tt.wantStatus {
				t.Errorf("状态码 = %d, 期望 %d", rr.Code, tt.wantStatus)
			}

			// 检查 next handler 是否被调用
			if tt.wantUserInCtx && !nextHandlerCalled {
				t.Error("next handler 应该被调用但没有")
			}
			if !tt.wantUserInCtx && nextHandlerCalled {
				t.Error("next handler 不应该被调用但被调用了")
			}

			// 检查上下文中的用户信息
			if tt.wantUserInCtx {
				if ctxUser == nil {
					t.Error("期望用户信息在上下文中但没有")
				} else if ctxUser.ID != testUser.ID {
					t.Errorf("上下文中的用户 ID = %d, 期望 %d", ctxUser.ID, testUser.ID)
				}
			}
		})
	}
}
