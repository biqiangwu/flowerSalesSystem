package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/biqiangwu/flowerSalesSystem/internal/user"
)

// HandleListUsers 处理获取用户列表请求
func (h *Handler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 session 中获取用户信息
	_, err := h.getUserFromSession(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "未登录或 session 无效")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	// 获取用户列表
	users, err := h.userService.ListUsers(ctx, page, pageSize)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	// 转换为响应格式
	type UserResponse struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
	}

	userResponses := make([]UserResponse, len(users))
	for i, u := range users {
		userResponses[i] = UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Role:      string(u.Role),
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"users": userResponses,
	})
}

// HandleDeleteUser 处理删除用户请求
func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 session 中获取操作者信息
	operator, err := h.getUserFromSession(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "未登录或 session 无效")
		return
	}

	// 从 URL 中提取用户 ID
	userIDStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的用户 ID")
		return
	}

	// 删除用户
	err = h.userService.DeleteUser(ctx, userID, operator.ID, operator.Role)
	if err != nil {
		if err == user.ErrUserNotFound {
			h.respondError(w, http.StatusNotFound, "用户不存在")
			return
		}
		if err == user.ErrInsufficientPermission {
			h.respondError(w, http.StatusForbidden, "权限不足")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "删除用户失败")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "用户已删除",
	})
}

// HandleResetPassword 处理重置用户密码请求
func (h *Handler) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 从 session 中获取操作者信息
	operator, err := h.getUserFromSession(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "未登录或 session 无效")
		return
	}

	// 从 URL 中提取用户 ID
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		h.respondError(w, http.StatusBadRequest, "无效的请求路径")
		return
	}
	userID, err := strconv.Atoi(parts[3])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的用户 ID")
		return
	}

	// 解析请求体
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的请求格式")
		return
	}

	// 重置密码
	err = h.userService.ResetPassword(ctx, userID, req.NewPassword, operator.ID, operator.Role)
	if err != nil {
		if err == user.ErrUserNotFound {
			h.respondError(w, http.StatusNotFound, "用户不存在")
			return
		}
		if err == user.ErrInsufficientPermission {
			h.respondError(w, http.StatusForbidden, "权限不足")
			return
		}
		if err == user.ErrInvalidPassword {
			h.respondError(w, http.StatusBadRequest, "密码无效（至少需要 6 个字符）")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "重置密码失败")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "密码已重置",
	})
}

// getUserFromSession 从请求中获取用户信息
func (h *Handler) getUserFromSession(r *http.Request) (*user.User, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}

	sessionUser, err := h.authService.ValidateSession(r.Context(), cookie.Value)
	if err != nil {
		return nil, err
	}

	return sessionUser, nil
}
