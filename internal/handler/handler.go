package handler

import (
	"encoding/json"
	"net/http"

	"github.com/biqiangwu/flowerSalesSystem/internal/address"
	"github.com/biqiangwu/flowerSalesSystem/internal/auth"
	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
	"github.com/biqiangwu/flowerSalesSystem/internal/order"
	"github.com/biqiangwu/flowerSalesSystem/internal/user"
)

// Handler HTTP 处理器
type Handler struct {
	authService     auth.AuthService
	orderService    order.OrderService
	orderLogService order.OrderLogService
	userService     user.UserService
	flowerService   flower.FlowerService
	addressService  address.AddressService
	userRepo        user.UserRepository // 用于测试时获取用户信息
}

// NewHandler 创建 Handler
func NewHandler(authService auth.AuthService, orderService order.OrderService) *Handler {
	return &Handler{
		authService:  authService,
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

// RegisterRoutes 注册所有路由
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// ========== 认证路由 ==========
	mux.HandleFunc("POST /api/register", h.HandleRegister)
	mux.HandleFunc("POST /api/login", h.HandleLogin)
	mux.HandleFunc("POST /api/logout", h.HandleLogout)

	// ========== 鲜花路由 ==========
	// 公开路由：所有用户可访问
	mux.HandleFunc("GET /api/flowers", h.HandleListFlowers)
	mux.HandleFunc("GET /api/flowers/", h.HandleGetFlower)

	// 需要认证的路由：店员和管理员
	mux.HandleFunc("POST /api/flowers", h.HandleCreateFlower)
	mux.HandleFunc("PUT /api/flowers/", h.HandleUpdateFlower)
	mux.HandleFunc("DELETE /api/flowers/", h.HandleDeleteFlower)
	mux.HandleFunc("POST /api/flowers/stock", h.HandleAddStock)

	// ========== 地址路由 ==========
	// 需要认证的路由：所有登录用户
	mux.HandleFunc("GET /api/addresses", h.HandleListAddresses)
	mux.HandleFunc("POST /api/addresses", h.HandleCreateAddress)
	mux.HandleFunc("PUT /api/addresses/", h.HandleUpdateAddress)
	mux.HandleFunc("DELETE /api/addresses/", h.HandleDeleteAddress)

	// ========== 订单路由 ==========
	// 需要认证的路由：所有登录用户
	mux.HandleFunc("POST /api/orders", h.HandleCreateOrder)
	mux.HandleFunc("GET /api/orders", h.HandleListOrders)
	mux.HandleFunc("GET /api/orders/", h.HandleGetOrder)

	// 订单状态流转路由
	mux.HandleFunc("POST /api/orders/complete", h.HandleCompleteOrder)
	mux.HandleFunc("POST /api/orders/cancel", h.HandleCancelOrder)

	// ========== 用户管理路由 ==========
	// 需要管理员权限的路由
	mux.HandleFunc("GET /api/users", h.HandleListUsers)
	mux.HandleFunc("DELETE /api/users/", h.HandleDeleteUser)
	mux.HandleFunc("POST /api/users/reset-password", h.HandleResetPassword)

	// ========== 订单日志路由 ==========
	// 需要认证的路由
	mux.HandleFunc("GET /api/orders/logs", h.HandleGetOrderLogs)
}

// SetServices 设置所有服务（用于依赖注入）
func (h *Handler) SetServices(
	authSvc auth.AuthService,
	orderSvc order.OrderService,
	orderLogSvc order.OrderLogService,
	userSvc user.UserService,
	flowerSvc flower.FlowerService,
	addressSvc address.AddressService,
	userRepo user.UserRepository,
) {
	h.authService = authSvc
	h.orderService = orderSvc
	h.orderLogService = orderLogSvc
	h.userService = userSvc
	h.flowerService = flowerSvc
	h.addressService = addressSvc
	h.userRepo = userRepo
}
