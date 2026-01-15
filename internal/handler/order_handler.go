package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/biqiangwu/flowerSalesSystem/internal/order"
)

// HandleCreateOrder 处理创建订单
func (h *Handler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 验证用户身份
	userID, ok := h.authenticateRequest(r)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 解析请求
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 转换为服务层请求
	serviceReq := &order.CreateOrderRequest{
		AddressID: req.AddressID,
		Items:     make([]*order.CreateOrderItemRequest, len(req.Items)),
	}
	for i, item := range req.Items {
		serviceReq.Items[i] = &order.CreateOrderItemRequest{
			FlowerSKU: item.FlowerSKU,
			Quantity:  item.Quantity,
		}
	}

	ctx := context.Background()
	orderNo, err := h.orderService.CreateOrder(ctx, userID, serviceReq)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "order created successfully",
		"order_no": orderNo,
	})
}

// HandleGetOrder 处理获取订单详情
func (h *Handler) HandleGetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 验证用户身份
	userID, ok := h.authenticateRequest(r)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 从 URL 获取订单号
	orderNo := extractOrderNo(r.URL.Path)
	if orderNo == "" {
		h.respondError(w, http.StatusBadRequest, "invalid order number")
		return
	}

	ctx := context.Background()
	orderResp, err := h.orderService.GetOrder(ctx, userID, orderNo)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "order not found")
			return
		}
		if strings.Contains(err.Error(), "无权") {
			h.respondError(w, http.StatusForbidden, "access denied")
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, orderResp)
}

// HandleListOrders 处理获取订单列表
func (h *Handler) HandleListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 验证用户身份
	userID, ok := h.authenticateRequest(r)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 解析查询参数
	filter := order.OrderListFilter{
		Status: r.URL.Query().Get("status"),
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			filter.PageSize = ps
		}
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	ctx := context.Background()
	orders, err := h.orderService.ListOrders(ctx, userID, filter)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, orders)
}

// authenticateRequest 验证请求并返回用户 ID
func (h *Handler) authenticateRequest(r *http.Request) (int, bool) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return 0, false
	}

	ctx := context.Background()
	u, err := h.authService.ValidateSession(ctx, cookie.Value)
	if err != nil {
		return 0, false
	}

	return u.ID, true
}

// extractOrderNo 从 URL 路径中提取订单号
func extractOrderNo(path string) string {
	// 路径格式: /api/orders/{order_no}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "orders" {
		return parts[2]
	}
	return ""
}
