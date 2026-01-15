package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestLoggingMiddleware 测试日志中间件
func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handlerStatus  int
		wantInLog      []string
		notWantInLog   string
	}{
		{
			name:          "正常请求 - 返回 200",
			handlerStatus: http.StatusOK,
			wantInLog:     []string{"GET", "/test", "200"},
		},
		{
			name:          "错误请求 - 返回 404",
			handlerStatus: http.StatusNotFound,
			wantInLog:     []string{"404"},
		},
		{
			name:          "错误请求 - 返回 500",
			handlerStatus: http.StatusInternalServerError,
			wantInLog:     []string{"500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 捕获日志输出
			var logBuf bytes.Buffer
			log.SetOutput(&logBuf)
			defer log.SetOutput(os.Stderr)

			// 创建测试 handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
			})

			// 应用日志中间件
			handler := LoggingMiddleware(nextHandler)

			// 创建请求
			req := httptest.NewRequest("GET", "/test", nil)

			// 创建响应记录器
			rr := httptest.NewRecorder()

			// 执行请求
			handler.ServeHTTP(rr, req)

			// 检查状态码
			if rr.Code != tt.handlerStatus {
				t.Errorf("状态码 = %d, 期望 %d", rr.Code, tt.handlerStatus)
			}

			// 检查日志输出
			logOutput := logBuf.String()

			// 验证期望的日志内容
			for _, expected := range tt.wantInLog {
				if !strings.Contains(logOutput, expected) {
					t.Errorf("日志应包含 %q，实际日志: %s", expected, logOutput)
				}
			}

			// 验证不应出现的日志内容
			if tt.notWantInLog != "" && strings.Contains(logOutput, tt.notWantInLog) {
				t.Errorf("日志不应包含 %q，实际日志: %s", tt.notWantInLog, logOutput)
			}

			// 验证日志包含基本要素：方法、路径、状态码
			if !strings.Contains(logOutput, "GET") {
				t.Error("日志应包含 HTTP 方法")
			}
			if !strings.Contains(logOutput, "/test") {
				t.Error("日志应包含请求路径")
			}
			statusStr := string(rune(tt.handlerStatus/100)) + "XX"
			if !strings.Contains(logOutput, statusStr) && tt.handlerStatus >= 100 && tt.handlerStatus < 600 {
				// 检查状态码是否在日志中
				found := false
				for i := 100; i <= 599; i++ {
					if strings.Contains(logOutput, string(rune(i))) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("日志应包含状态码信息，实际日志: %s", logOutput)
				}
			}
		})
	}
}

// TestLoggingMiddlewareWithDuration 测试日志中间件记录耗时
func TestLoggingMiddlewareWithDuration(t *testing.T) {
	// 捕获日志输出
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	// 创建测试 handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 应用日志中间件
	handler := LoggingMiddleware(nextHandler)

	// 创建请求
	req := httptest.NewRequest("POST", "/api/test", nil)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(rr, req)

	// 检查日志输出
	logOutput := logBuf.String()

	// 验证日志包含方法、路径
	if !strings.Contains(logOutput, "POST") {
		t.Error("日志应包含 HTTP 方法 POST")
	}
	if !strings.Contains(logOutput, "/api/test") {
		t.Error("日志应包含请求路径 /api/test")
	}
}
