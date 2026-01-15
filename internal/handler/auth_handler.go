package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	// CookieName Session Cookie 名称
	CookieName = "session_token"
)

// HandleRegister 处理用户注册
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()

	// 注册用户
	u, err := h.authService.Register(ctx, req.Username, req.Password)
	if err != nil {
		if containsString(err.Error(), "already exists") {
			h.respondError(w, http.StatusConflict, "username already exists")
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "user registered successfully",
		"user": map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"role":     u.Role,
		},
	})
}

// HandleLogin 处理用户登录
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()
	session, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		if containsString(err.Error(), "invalid") {
			h.respondError(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 设置 Session Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    session.Token,
		Path:     "/",
		MaxAge:   86400, // 24 小时
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "login successful",
		"user": map[string]interface{}{
			"id":       session.UserID,
			"username": session.Username,
			"role":     session.Role,
		},
	})
}

// HandleLogout 处理用户登出
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 获取 Session Token
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "not logged in")
		return
	}

	ctx := context.Background()
	if err := h.authService.Logout(ctx, cookie.Value); err != nil {
		h.respondError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	// 清除 Session Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "logout successful",
	})
}

// containsString 检查字符串是否包含子串（忽略大小写）
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsIgnoreCase(s, substr))
}

func containsIgnoreCase(s, substr string) bool {
	// 简单实现，实际项目中可以使用 strings.ToLower
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
