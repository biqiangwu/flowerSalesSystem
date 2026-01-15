package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
)

// HandleListFlowers 处理获取鲜花列表
func (h *Handler) HandleListFlowers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 解析查询参数
	filter := flower.FlowerFilter{
		Search:   r.URL.Query().Get("search"),
		Origin:   r.URL.Query().Get("origin"),
		MinPrice: parseFloatQuery(r.URL.Query().Get("min_price")),
		MaxPrice: parseFloatQuery(r.URL.Query().Get("max_price")),
		SortBy:   r.URL.Query().Get("sort_by"),
		Page:     parseIntQuery(r.URL.Query().Get("page"), 1),
		PageSize: parseIntQuery(r.URL.Query().Get("page_size"), 10),
	}

	ctx := context.Background()
	flowers, err := h.flowerService.ListFlowers(ctx, filter)
	if err != nil {
		if strings.Contains(err.Error(), "最低价格") || strings.Contains(err.Error(), "页码") {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, flowers)
}

// HandleGetFlower 处理获取鲜花详情
func (h *Handler) HandleGetFlower(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sku := extractFlowerSKU(r.URL.Path)
	if sku == "" {
		h.respondError(w, http.StatusBadRequest, "invalid flower SKU")
		return
	}

	ctx := context.Background()
	flowerResp, err := h.flowerService.GetFlower(ctx, sku)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			h.respondError(w, http.StatusNotFound, "flower not found")
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, flowerResp)
}

// HandleCreateFlower 处理创建鲜花
func (h *Handler) HandleCreateFlower(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CreateFlowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()
	if err := h.flowerService.CreateFlower(ctx, &flower.CreateFlowerRequest{
		SKU:           req.SKU,
		Name:          req.Name,
		Origin:        req.Origin,
		ShelfLife:     req.ShelfLife,
		Preservation:  req.Preservation,
		PurchasePrice: req.PurchasePrice,
		SalePrice:     req.SalePrice,
		Stock:         req.Stock,
	}); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "flower created successfully",
		"sku":     req.SKU,
	})
}

// HandleUpdateFlower 处理更新鲜花
func (h *Handler) HandleUpdateFlower(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sku := extractFlowerSKU(r.URL.Path)
	if sku == "" {
		h.respondError(w, http.StatusBadRequest, "invalid flower SKU")
		return
	}

	var req UpdateFlowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()
	updateReq := &flower.UpdateFlowerRequest{}
	if req.Name != nil {
		updateReq.Name = req.Name
	}
	if req.Origin != nil {
		updateReq.Origin = req.Origin
	}
	if req.ShelfLife != nil {
		updateReq.ShelfLife = req.ShelfLife
	}
	if req.Preservation != nil {
		updateReq.Preservation = req.Preservation
	}
	if req.PurchasePrice != nil {
		updateReq.PurchasePrice = req.PurchasePrice
	}
	if req.SalePrice != nil {
		updateReq.SalePrice = req.SalePrice
	}

	if err := h.flowerService.UpdateFlower(ctx, sku, updateReq); err != nil {
		if strings.Contains(err.Error(), "不存在") {
			h.respondError(w, http.StatusNotFound, "flower not found")
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "flower updated successfully",
	})
}

// HandleDeleteFlower 处理删除鲜花
func (h *Handler) HandleDeleteFlower(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sku := extractFlowerSKU(r.URL.Path)
	if sku == "" {
		h.respondError(w, http.StatusBadRequest, "invalid flower SKU")
		return
	}

	ctx := context.Background()
	if err := h.flowerService.DeleteFlower(ctx, sku); err != nil {
		if strings.Contains(err.Error(), "不存在") {
			h.respondError(w, http.StatusNotFound, "flower not found")
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "flower deleted successfully",
	})
}

// HandleAddStock 处理进货入库
func (h *Handler) HandleAddStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sku := extractFlowerSKU(r.URL.Path)
	if sku == "" {
		h.respondError(w, http.StatusBadRequest, "invalid flower SKU")
		return
	}

	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()
	if err := h.flowerService.AddStock(ctx, sku, req.Quantity); err != nil {
		if strings.Contains(err.Error(), "不存在") {
			h.respondError(w, http.StatusNotFound, "flower not found")
			return
		}
		if strings.Contains(err.Error(), "不能为负") {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 获取更新后的库存
	flowerResp, _ := h.flowerService.GetFlower(ctx, sku)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "stock added successfully",
		"sku":     sku,
		"stock":   flowerResp.Stock,
	})
}

// extractFlowerSKU 从 URL 路径中提取鲜花 SKU
func extractFlowerSKU(path string) string {
	// 路径格式: /api/flowers/{sku} 或 /api/flowers/{sku}/stock
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "flowers" {
		return parts[2]
	}
	return ""
}

// parseFloatQuery 解析浮点数查询参数
func parseFloatQuery(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseIntQuery 解析整数查询参数
func parseIntQuery(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	i, _ := strconv.Atoi(s)
	if i <= 0 {
		return defaultVal
	}
	return i
}
