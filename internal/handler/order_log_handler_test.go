package handler

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/biqiangwu/flowerSalesSystem/internal/address"
	"github.com/biqiangwu/flowerSalesSystem/internal/auth"
	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
	"github.com/biqiangwu/flowerSalesSystem/internal/order"
	"github.com/biqiangwu/flowerSalesSystem/internal/user"
)

// setupOrderLogTestHandler 创建订单日志测试用的 Handler，同时返回 db
func setupOrderLogTestHandler(t *testing.T) (*Handler, *sql.DB) {
	t.Helper()

	db := setupOrderTestDB(t)

	// 初始化各个层
	userRepo := user.NewMySQLUserRepository(db)
	sessionMgr := auth.NewMemorySessionManager()
	authSvc := auth.NewAuthService(userRepo, sessionMgr)

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := order.NewOrderRepository(db)
	orderLogRepo := order.NewOrderLogRepository(db)
	orderLogSvc := order.NewOrderLogService(orderLogRepo)

	orderSvc := order.NewOrderService(orderRepo, flowerRepo, orderLogRepo)

	handler := &Handler{
		authService:     authSvc,
		orderService:    orderSvc,
		orderLogService: orderLogSvc,
	}

	return handler, db
}

// TestHandleGetOrderLogs 测试获取订单日志
func TestHandleGetOrderLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, h *Handler, db *sql.DB) string // 返回 sessionToken
		urlParams   string
		wantStatus  int
		validateResp func(t *testing.T, body string)
	}{
		{
			name: "正常查询订单日志",
			setupFunc: func(t *testing.T, h *Handler, db *sql.DB) string {
				ctx := context.Background()

				// 登录用户
				sessionToken := loginUser(t, h, "testuser", "password123")

				// 获取用户ID
				userRepo := user.NewMySQLUserRepository(db)
				u, _ := userRepo.GetByUsername(ctx, "testuser")

				// 创建地址和鲜花
				addressRepo := address.NewAddressRepository(db)
				flowerRepo := flower.NewFlowerRepository(db)
				orderRepo := order.NewOrderRepository(db)
				orderLogRepo := order.NewOrderLogRepository(db)
				orderLogSvc := order.NewOrderLogService(orderLogRepo)

				addr := &address.Address{
					UserID:  u.ID,
					Label:   "家",
					Address: "北京市朝阳区",
					Contact: "张三",
				}
				addressRepo.Create(ctx, addr)

				flw := &flower.Flower{
					SKU:           "FLW001",
					Name:          "红玫瑰",
					Origin:        "云南",
					ShelfLife:     "7天",
					Preservation:  "冷藏",
					PurchasePrice: flower.Decimal{Value: 5000},
					SalePrice:     flower.Decimal{Value: 10000},
					Stock:         100,
					IsActive:      true,
				}
				flowerRepo.Create(ctx, flw)

				// 创建订单
				createReq := &order.CreateOrderRequest{
					AddressID: addr.ID,
					Items: []*order.CreateOrderItemRequest{
						{FlowerSKU: "FLW001", Quantity: 5},
					},
				}

				orderSvc := order.NewOrderService(orderRepo, flowerRepo, orderLogRepo)
				orderNo, _ := orderSvc.CreateOrder(ctx, u.ID, createReq)

				// 通过订单号获取订单
				o, _, _ := orderRepo.GetByOrderNo(ctx, orderNo)

				// 添加一些日志
				orderLogSvc.LogOrderAction(ctx, o.ID, u.ID, "create_order", "", order.StatusPending)
				orderLogSvc.LogOrderAction(ctx, o.ID, u.ID, "complete_order", order.StatusPending, order.StatusCompleted)

				// 保存订单 ID 到测试环境变量，供 URL 参数使用
				os.Setenv("TEST_ORDER_ID", strconv.Itoa(o.ID))

				return sessionToken
			},
			urlParams:  "?order_id=1", // 会从环境变量读取
			wantStatus: http.StatusOK,
			validateResp: func(t *testing.T, body string) {
				// 验证返回的是日志数组
				if body == "" || body == "null" || body == "[]" {
					t.Error("expected non-empty logs array")
				}
			},
		},
		{
			name: "查询无日志的订单",
			setupFunc: func(t *testing.T, h *Handler, db *sql.DB) string {
				ctx := context.Background()

				// 登录用户
				sessionToken := loginUser(t, h, "testuser2", "password123")

				// 获取用户ID
				userRepo := user.NewMySQLUserRepository(db)
				u, _ := userRepo.GetByUsername(ctx, "testuser2")

				// 创建地址和鲜花
				addressRepo := address.NewAddressRepository(db)
				flowerRepo := flower.NewFlowerRepository(db)
				orderRepo := order.NewOrderRepository(db)
				orderLogRepo := order.NewOrderLogRepository(db)

				addr := &address.Address{
					UserID:  u.ID,
					Label:   "家",
					Address: "北京市朝阳区",
					Contact: "张三",
				}
				addressRepo.Create(ctx, addr)

				flw := &flower.Flower{
					SKU:           "FLW002",
					Name:          "白玫瑰",
					Origin:        "云南",
					ShelfLife:     "7天",
					Preservation:  "冷藏",
					PurchasePrice: flower.Decimal{Value: 5000},
					SalePrice:     flower.Decimal{Value: 10000},
					Stock:         100,
					IsActive:      true,
				}
				flowerRepo.Create(ctx, flw)

				// 创建订单（不添加日志，模拟无日志订单）
				createReq := &order.CreateOrderRequest{
					AddressID: addr.ID,
					Items: []*order.CreateOrderItemRequest{
						{FlowerSKU: "FLW002", Quantity: 5},
					},
				}

				orderSvc := order.NewOrderService(orderRepo, flowerRepo, orderLogRepo)
				orderNo, _ := orderSvc.CreateOrder(ctx, u.ID, createReq)

				// 通过订单号获取订单
				o, _, _ := orderRepo.GetByOrderNo(ctx, orderNo)

				// 修改 URL 参数为实际订单 ID
				os.Setenv("TEST_ORDER_ID", strconv.Itoa(o.ID))

				return sessionToken
			},
			urlParams:  "?order_id=999",
			wantStatus: http.StatusOK,
			validateResp: func(t *testing.T, body string) {
				// 空数组也是有效的响应
				if body == "" || body == "null" {
					t.Error("expected valid response")
				}
			},
		},
		{
			name: "订单ID无效",
			setupFunc: func(t *testing.T, h *Handler, db *sql.DB) string {
				return loginUser(t, h, "testuser3", "password123")
			},
			urlParams:  "?order_id=0",
			wantStatus: http.StatusBadRequest,
			validateResp: nil,
		},
		{
			name: "订单ID格式错误",
			setupFunc: func(t *testing.T, h *Handler, db *sql.DB) string {
				return loginUser(t, h, "testuser4", "password123")
			},
			urlParams:  "?order_id=abc",
			wantStatus: http.StatusBadRequest,
			validateResp: nil,
		},
		{
			name: "未授权访问",
			setupFunc: func(t *testing.T, h *Handler, db *sql.DB) string {
				return "" // 不返回 sessionToken，模拟未登录
			},
			urlParams:  "?order_id=1",
			wantStatus: http.StatusUnauthorized,
			validateResp: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理环境变量
			defer os.Unsetenv("TEST_ORDER_ID")

			handler, db := setupOrderLogTestHandler(t)

			sessionToken := tt.setupFunc(t, handler, db)

			// 构建请求 URL - 如果设置了环境变量，则使用环境变量中的订单 ID
			reqURL := "/api/orders/logs" + tt.urlParams
			if orderID := os.Getenv("TEST_ORDER_ID"); orderID != "" {
				// 替换 URL 中的 order_id 参数
				u, _ := url.Parse(reqURL)
				q := u.Query()
				q.Set("order_id", orderID)
				u.RawQuery = q.Encode()
				reqURL = u.String()
			}

			req := httptest.NewRequest("GET", reqURL, nil)
			if sessionToken != "" {
				req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
			}
			w := httptest.NewRecorder()

			handler.HandleGetOrderLogs(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleGetOrderLogs() status = %d, want %d, body = %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.validateResp != nil && w.Code == http.StatusOK {
				tt.validateResp(t, w.Body.String())
			}
		})
	}
}

// TestHandleGetOrderLogs_ParseOrderIDFromURL 测试从 URL 路径解析订单ID
func TestHandleGetOrderLogs_ParseOrderIDFromURL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, db := setupOrderLogTestHandler(t)
	ctx := context.Background()

	// 准备测试数据
	sessionToken := loginUser(t, handler, "testuser", "password123")

	userRepo := user.NewMySQLUserRepository(db)
	u, _ := userRepo.GetByUsername(ctx, "testuser")

	addressRepo := address.NewAddressRepository(db)
	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := order.NewOrderRepository(db)
	orderLogRepo := order.NewOrderLogRepository(db)
	orderLogSvc := order.NewOrderLogService(orderLogRepo)

	addr := &address.Address{
		UserID:  u.ID,
		Label:   "家",
		Address: "北京市朝阳区",
		Contact: "张三",
	}
	addressRepo.Create(ctx, addr)

	flw := &flower.Flower{
		SKU:           "FLW001",
		Name:          "红玫瑰",
		Origin:        "云南",
		ShelfLife:     "7天",
		Preservation:  "冷藏",
		PurchasePrice: flower.Decimal{Value: 5000},
		SalePrice:     flower.Decimal{Value: 10000},
		Stock:         100,
		IsActive:      true,
	}
	flowerRepo.Create(ctx, flw)

	// 创建订单
	createReq := &order.CreateOrderRequest{
		AddressID: addr.ID,
		Items: []*order.CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 5},
		},
	}

	orderSvc := order.NewOrderService(orderRepo, flowerRepo, orderLogRepo)
	orderNo, _ := orderSvc.CreateOrder(ctx, u.ID, createReq)

	// 通过订单号获取订单
	o, _, _ := orderRepo.GetByOrderNo(ctx, orderNo)

	// 添加日志
	orderLogSvc.LogOrderAction(ctx, o.ID, u.ID, "create_order", "", order.StatusPending)

	tests := []struct {
		name       string
		urlPattern string
		wantStatus int
	}{
		{
			name:       "使用 /api/orders/{orderID}/logs 路径",
			urlPattern: "/api/orders/" + strconv.Itoa(o.ID) + "/logs",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.urlPattern, nil)
			req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
			w := httptest.NewRecorder()

			handler.HandleGetOrderLogs(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleGetOrderLogs() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// TestHandleGetOrderLogs_QueryParams 测试查询参数解析
func TestHandleGetOrderLogs_QueryParams(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _ := setupOrderLogTestHandler(t)
	sessionToken := loginUser(t, handler, "testuser", "password123")

	tests := []struct {
		name       string
		queryParam string
		wantStatus int
	}{
		{
			name:       "有效的 order_id 参数",
			queryParam: "?order_id=1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "缺少 order_id 参数",
			queryParam: "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "order_id 为非数字",
			queryParam: "?order_id=invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/orders/logs"+tt.queryParam, nil)
			req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
			w := httptest.NewRecorder()

			handler.HandleGetOrderLogs(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleGetOrderLogs() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// TestHandleGetOrderLogs_Unauthorized 不同场景的未授权测试
func TestHandleGetOrderLogs_Unauthorized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name        string
		setupCookie func(*http.Request)
		wantStatus  int
	}{
		{
			name: "无 session_token cookie",
			setupCookie: func(req *http.Request) {
				// 不设置 cookie
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "空的 session_token",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "session_token", Value: ""})
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "无效的 session_token",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "session_token", Value: "invalid_token_12345"})
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := setupOrderLogTestHandler(t)

			req := httptest.NewRequest("GET", "/api/orders/logs?order_id=1", nil)
			tt.setupCookie(req)
			w := httptest.NewRecorder()

			handler.HandleGetOrderLogs(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleGetOrderLogs() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
