package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// HandleGetOrderLogs 处理获取订单日志请求
// 支持两种 URL 格式：
// 1. GET /api/orders/logs?order_id=123
// 2. GET /api/orders/123/logs
func (h *Handler) HandleGetOrderLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 验证用户身份
	_, ok := h.authenticateRequest(r)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 获取订单 ID - 支持两种方式
	var orderID int
	var err error

	// 方式1: 优先从查询参数解析 ?order_id=123
	orderIDStr := r.URL.Query().Get("order_id")
	if orderIDStr != "" {
		orderID, err = strconv.Atoi(orderIDStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "订单ID格式错误")
			return
		}
	} else {
		// 方式2: 从 URL 路径解析 /api/orders/{orderID}/logs
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/orders/") && strings.HasSuffix(path, "/logs") {
			// 提取订单 ID: /api/orders/123/logs -> 123
			parts := strings.Split(path, "/")
			if len(parts) >= 4 {
				orderID, err = strconv.Atoi(parts[3])
				if err != nil {
					h.respondError(w, http.StatusBadRequest, "订单ID格式错误")
					return
				}
			}
		}
	}

	// 验证订单 ID
	if orderID <= 0 {
		h.respondError(w, http.StatusBadRequest, "订单ID无效")
		return
	}

	// 查询订单日志
	logs, err := h.orderLogService.GetOrderLogs(r.Context(), orderID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, fmt.Sprintf("查询订单日志失败: %v", err))
		return
	}

	// 返回日志列表
	h.respondJSON(w, http.StatusOK, logs)
}
