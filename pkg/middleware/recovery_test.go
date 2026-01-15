package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRecoveryMiddleware 测试恢复中间件
func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handlerFunc    http.HandlerFunc
		wantStatus     int
		wantBody       string
	}{
		{
			name: "正常请求 - 响应不变",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			},
			wantStatus: http.StatusOK,
			wantBody:   "OK",
		},
		{
			name: "Panic 请求 - 返回 500",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				panic("unexpected error")
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Internal server error",
		},
		{
			name: "Panic with string - 返回 500",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				panic("some error message")
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Internal server error",
		},
		{
			name: "Panic with error - 返回 500",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				panic(http.ErrHandlerTimeout)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Internal server error",
		},
		{
			name: "正常请求返回 404",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not Found"))
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "Not Found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 应用恢复中间件
			handler := RecoveryMiddleware(http.HandlerFunc(tt.handlerFunc))

			// 创建请求
			req := httptest.NewRequest("GET", "/test", nil)

			// 创建响应记录器
			rr := httptest.NewRecorder()

				// 执行请求（捕获 panic）
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Errorf("中间件应该捕获 panic，但 panic 逃逸了: %v", recovered)
					}
				}()
				handler.ServeHTTP(rr, req)
			}()

			// 检查状态码
			if rr.Code != tt.wantStatus {
				t.Errorf("状态码 = %d, 期望 %d", rr.Code, tt.wantStatus)
			}

			// 检查响应体
			body := rr.Body.String()
			if tt.wantBody != "" && !containsString(body, tt.wantBody) {
				t.Errorf("响应体应包含 %q，实际: %s", tt.wantBody, body)
			}
		})
	}
}

// containsString 检查字符串是否包含子字符串（忽略大小写）
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		containsIgnoreCase(s, substr))
}

func containsIgnoreCase(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
