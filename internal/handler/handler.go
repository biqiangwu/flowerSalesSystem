package handler

import (
	"encoding/json"
	"net/http"

	"github.com/biqiangwu/flowerSalesSystem/internal/auth"
	"github.com/biqiangwu/flowerSalesSystem/internal/order"
)

// Handler HTTP 处理器
type Handler struct {
	authService auth.AuthService
	orderService order.OrderService
	// 其他 service 稍后添加
	// flowerService flower.FlowerService
	// userService   user.UserService
}

// NewHandler 创建 Handler
func NewHandler(authService auth.AuthService, orderService order.OrderService) *Handler {
	return &Handler{
		authService: authService,
		orderService: orderService,
	}
}

// respondJSON 返回 JSON 响应
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

// respondError 返回错误响应
func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, ErrorResponse{Error: message})
}
