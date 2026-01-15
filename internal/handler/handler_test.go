package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewHandler 测试 Handler 创建
func TestNewHandler(t *testing.T) {
	h := NewHandler(nil, nil)

	if h == nil {
		t.Fatal("NewHandler() returned nil")
	}

	if h.authService != nil {
		t.Error("authService should be nil when passed nil")
	}

	if h.orderService != nil {
		t.Error("orderService should be nil when passed nil")
	}
}

// TestHandler_respondJSON 测试 respondJSON 方法
func TestHandler_respondJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       interface{}
		wantStatus int
		wantHeader string
	}{
		{
			name:       "respond with JSON data",
			status:     http.StatusOK,
			data:       map[string]string{"message": "success"},
			wantStatus: http.StatusOK,
			wantHeader: "application/json",
		},
		{
			name:       "respond with created status",
			status:     http.StatusCreated,
			data:       map[string]int{"id": 123},
			wantStatus: http.StatusCreated,
			wantHeader: "application/json",
		},
		{
			name:       "respond with nil data",
			status:     http.StatusNoContent,
			data:       nil,
			wantStatus: http.StatusNoContent,
			wantHeader: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(nil, nil)
			w := httptest.NewRecorder()

			h.respondJSON(w, tt.status, tt.data)

			if w.Code != tt.wantStatus {
				t.Errorf("respondJSON() status = %d, want %d", w.Code, tt.wantStatus)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != tt.wantHeader {
				t.Errorf("respondJSON() Content-Type = %q, want %q", contentType, tt.wantHeader)
			}
		})
	}
}

// TestHandler_respondError 测试 respondError 方法
func TestHandler_respondError(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		message    string
		wantStatus int
		wantError  string
	}{
		{
			name:       "respond with bad request error",
			status:     http.StatusBadRequest,
			message:    "invalid input",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid input",
		},
		{
			name:       "respond with unauthorized error",
			status:     http.StatusUnauthorized,
			message:    "unauthorized access",
			wantStatus: http.StatusUnauthorized,
			wantError:  "unauthorized access",
		},
		{
			name:       "respond with not found error",
			status:     http.StatusNotFound,
			message:    "resource not found",
			wantStatus: http.StatusNotFound,
			wantError:  "resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(nil, nil)
			w := httptest.NewRecorder()

			h.respondError(w, tt.status, tt.message)

			if w.Code != tt.wantStatus {
				t.Errorf("respondError() status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp.Error != tt.wantError {
				t.Errorf("respondError() error = %q, want %q", resp.Error, tt.wantError)
			}
		})
	}
}

// TestCookieName 测试 Cookie 名称常量
func TestCookieName(t *testing.T) {
	if CookieName != "session_token" {
		t.Errorf("CookieName = %q, want session_token", CookieName)
	}
}
