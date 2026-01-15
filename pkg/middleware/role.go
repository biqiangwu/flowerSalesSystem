package middleware

import (
	"fmt"
	"net/http"

	"github.com/biqiangwu/flowerSalesSystem/internal/user"
)

// RoleMiddleware 角色中间件，验证用户角色是否有权限访问
// 返回一个中间件函数，该函数接受下一个 handler 并返回一个新的 handler
func RoleMiddleware(allowedRoles ...user.Role) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 从上下文获取用户信息
			u, ok := GetUserFromContext(r)
			if !ok {
				http.Error(w, "unauthorized: no user in context", http.StatusUnauthorized)
				return
			}

			// 检查用户角色是否在允许列表中
			allowed := false
			for _, role := range allowedRoles {
				if u.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, fmt.Sprintf("forbidden: user role %s not allowed", u.Role), http.StatusForbidden)
				return
			}

			// 权限验证通过，调用下一个 handler
			next(w, r)
		}
	}
}
