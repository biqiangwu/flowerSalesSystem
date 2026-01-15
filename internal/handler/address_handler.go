package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/biqiangwu/flowerSalesSystem/internal/address"
	"github.com/biqiangwu/flowerSalesSystem/pkg/middleware"
)

// HandleListAddresses 处理获取地址列表
func (h *Handler) HandleListAddresses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从上下文获取用户信息
	u, ok := middleware.GetUserFromContext(r)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := context.Background()
	addresses, err := h.addressService.ListAddresses(ctx, u.ID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, addresses)
}

// HandleCreateAddress 处理创建地址
func (h *Handler) HandleCreateAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从上下文获取用户信息
	u, ok := middleware.GetUserFromContext(r)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()
	if err := h.addressService.CreateAddress(ctx, u.ID, &address.CreateAddressRequest{
		Label:   req.Label,
		Address: req.Address,
		Contact: req.Contact,
	}); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "address created successfully",
	})
}

// HandleUpdateAddress 处理更新地址
func (h *Handler) HandleUpdateAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从上下文获取用户信息
	u, ok := middleware.GetUserFromContext(r)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 从 URL 获取地址ID
	addressID := extractAddressID(r.URL.Path)
	if addressID <= 0 {
		h.respondError(w, http.StatusBadRequest, "invalid address id")
		return
	}

	var req UpdateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()
	updateReq := &address.UpdateAddressRequest{}
	if req.Label != nil {
		updateReq.Label = req.Label
	}
	if req.Address != nil {
		updateReq.Address = req.Address
	}
	if req.Contact != nil {
		updateReq.Contact = req.Contact
	}

	if err := h.addressService.UpdateAddress(ctx, u.ID, addressID, updateReq); err != nil {
		if strings.Contains(err.Error(), "不存在") {
			h.respondError(w, http.StatusNotFound, "address not found")
			return
		}
		if strings.Contains(err.Error(), "无权") {
			h.respondError(w, http.StatusForbidden, err.Error())
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "address updated successfully",
	})
}

// HandleDeleteAddress 处理删除地址
func (h *Handler) HandleDeleteAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从上下文获取用户信息
	u, ok := middleware.GetUserFromContext(r)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 从 URL 获取地址ID
	addressID := extractAddressID(r.URL.Path)
	if addressID <= 0 {
		h.respondError(w, http.StatusBadRequest, "invalid address id")
		return
	}

	ctx := context.Background()
	if err := h.addressService.DeleteAddress(ctx, u.ID, addressID); err != nil {
		if strings.Contains(err.Error(), "不存在") {
			h.respondError(w, http.StatusNotFound, "address not found")
			return
		}
		if strings.Contains(err.Error(), "无权") {
			h.respondError(w, http.StatusForbidden, err.Error())
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "address deleted successfully",
	})
}

// extractAddressID 从 URL 路径中提取地址ID
func extractAddressID(path string) int {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "addresses" {
		if id, err := strconv.Atoi(parts[2]); err == nil {
			return id
		}
	}
	return 0
}
