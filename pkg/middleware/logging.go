package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware 日志中间件，记录请求方法、路径、状态码、耗时
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 记录请求开始时间
		start := time.Now()

		// 创建响应记录器以捕获状态码
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// 调用下一个 handler
		next.ServeHTTP(rw, r)

		// 计算耗时
		duration := time.Since(start)

		// 记录日志
		log.Printf("[%s] %s - %d - %v", r.Method, r.URL.Path, rw.status, duration)
	})
}

// responseWriter 包装 http.ResponseWriter 以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader 捕获状态码
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
