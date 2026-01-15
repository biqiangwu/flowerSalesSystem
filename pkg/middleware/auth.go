package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/biqiangwu/flowerSalesSystem/internal/auth"
	"github.com/biqiangwu/flowerSalesSystem/internal/user"
)

// contextKey 是用于上下文键的自定义类型，避免键冲突
type contextKey string

// userKey 是用于存储用户信息的上下文键
const userKey contextKey = "user"

// AuthMiddleware 认证中间件，验证用户的 session token
func AuthMiddleware(authService auth.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从 Cookie 读取 session_token
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(w, "unauthorized: missing session token", http.StatusUnauthorized)
				return
			}
			http.Error(w, "unauthorized: invalid cookie", http.StatusUnauthorized)
			return
		}

		// 验证 Session
		u, err := authService.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// 将用户信息注入上下文
		ctx := context.WithValue(r.Context(), userKey, u)
		next(w, r.WithContext(ctx))
	}
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(r *http.Request) (*user.User, bool) {
	u, ok := r.Context().Value(userKey).(*user.User)
	return u, ok
}
