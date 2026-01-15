package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/biqiangwu/flowerSalesSystem/internal/user"
)

// TestRoleMiddleware 测试角色中间件
func TestRoleMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		user          *user.User
		allowedRoles  []user.Role
		wantStatus    int
		handlerCalled bool
	}{
		{
			name: "admin 用户有权限访问 admin, clerk 路由",
			user: &user.User{
				ID:       1,
				Username: "admin",
				Role:     user.RoleAdmin,
			},
			allowedRoles:  []user.Role{user.RoleAdmin, user.RoleClerk},
			wantStatus:    http.StatusOK,
			handlerCalled: true,
		},
		{
			name: "clerk 用户有权限访问 admin, clerk 路由",
			user: &user.User{
				ID:       2,
				Username: "clerk",
				Role:     user.RoleClerk,
			},
			allowedRoles:  []user.Role{user.RoleAdmin, user.RoleClerk},
			wantStatus:    http.StatusOK,
			handlerCalled: true,
		},
		{
			name: "customer 用户无权限访问 admin, clerk 路由 - 返回 403",
			user: &user.User{
				ID:       3,
				Username: "customer",
				Role:     user.RoleCustomer,
			},
			allowedRoles:  []user.Role{user.RoleAdmin, user.RoleClerk},
			wantStatus:    http.StatusForbidden,
			handlerCalled: false,
		},
		{
			name:          "无用户信息 - 返回 401",
			user:          nil,
			allowedRoles:  []user.Role{user.RoleAdmin, user.RoleClerk},
			wantStatus:    http.StatusUnauthorized,
			handlerCalled: false,
		},
		{
			name: "admin 用户有权限访问仅 admin 路由",
			user: &user.User{
				ID:       1,
				Username: "admin",
				Role:     user.RoleAdmin,
			},
			allowedRoles:  []user.Role{user.RoleAdmin},
			wantStatus:    http.StatusOK,
			handlerCalled: true,
		},
		{
			name: "clerk 用户无权限访问仅 admin 路由 - 返回 403",
			user: &user.User{
				ID:       2,
				Username: "clerk",
				Role:     user.RoleClerk,
			},
			allowedRoles:  []user.Role{user.RoleAdmin},
			wantStatus:    http.StatusForbidden,
			handlerCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个测试 handler
			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			// 应用角色中间件
			middleware := RoleMiddleware(tt.allowedRoles...)
			handler := middleware(nextHandler)

			// 创建请求
			req := httptest.NewRequest("GET", "/test", nil)
			// 将用户信息注入上下文
			if tt.user != nil {
				ctx := context.WithValue(req.Context(), userKey, tt.user)
				req = req.WithContext(ctx)
			}

			// 创建响应记录器
			rr := httptest.NewRecorder()

			// 执行请求
			handler.ServeHTTP(rr, req)

			// 检查状态码
			if rr.Code != tt.wantStatus {
				t.Errorf("状态码 = %d, 期望 %d", rr.Code, tt.wantStatus)
			}

			// 检查 handler 是否被调用
			if handlerCalled != tt.handlerCalled {
				t.Errorf("handler 被调用 = %v, 期望 %v", handlerCalled, tt.handlerCalled)
			}
		})
	}
}
