package middleware

import (
	"log"
	"net/http"
)

// RecoveryMiddleware 恢复中间件，捕获 panic 并返回 500 错误
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 使用 defer + recover 捕获 panic
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 日志
				log.Printf("panic recovered: %v", err)

				// 返回 500 错误
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		// 调用下一个 handler
		next.ServeHTTP(w, r)
	})
}
