package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/biqiangwu/flowerSalesSystem/internal/address"
	"github.com/biqiangwu/flowerSalesSystem/internal/auth"
	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
	"github.com/biqiangwu/flowerSalesSystem/internal/order"
	"github.com/biqiangwu/flowerSalesSystem/internal/user"
	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
)

// setupOrderTestDB 创建完整的测试数据库
func setupOrderTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// 创建用户表
	createUsersTable := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`

	// 创建地址表
	createAddressesTable := `
		CREATE TABLE IF NOT EXISTS addresses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			label TEXT,
			address TEXT NOT NULL,
			contact TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`

	// 创建鲜花表
	createFlowersTable := `
		CREATE TABLE IF NOT EXISTS flowers (
			sku TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			origin TEXT NOT NULL,
			shelf_life TEXT NOT NULL,
			preservation TEXT NOT NULL,
			purchase_price INTEGER NOT NULL,
			sale_price INTEGER NOT NULL,
			stock INTEGER NOT NULL DEFAULT 0,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`

	// 创建订单表
	createOrdersTable := `
		CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_no TEXT UNIQUE NOT NULL,
			user_id INTEGER NOT NULL,
			address_id INTEGER NOT NULL,
			total_amount INTEGER NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (address_id) REFERENCES addresses(id)
		);
	`

	// 创建订单项表
	createOrderItemsTable := `
		CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			flower_sku TEXT NOT NULL,
			flower_name TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			unit_price INTEGER NOT NULL,
			subtotal INTEGER NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
		);
	`

	// 创建订单日志表
	createOrderLogsTable := `
		CREATE TABLE IF NOT EXISTS order_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			operator_id INTEGER NOT NULL,
			action TEXT NOT NULL,
			old_status TEXT,
			new_status TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
			FOREIGN KEY (operator_id) REFERENCES users(id)
		);
	`

	tables := []string{
		createUsersTable, createAddressesTable, createFlowersTable,
		createOrdersTable, createOrderItemsTable, createOrderLogsTable,
	}

	for _, tableSQL := range tables {
		if _, err := db.Exec(tableSQL); err != nil {
			db.Close()
			t.Fatalf("failed to create table: %v", err)
		}
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// setupOrderTestHandler 创建订单测试用的 Handler，同时返回 db
func setupOrderTestHandler(t *testing.T) (*Handler, *sql.DB) {
	t.Helper()

	db := setupOrderTestDB(t)

	// 初始化各个层
	userRepo := user.NewMySQLUserRepository(db)
	sessionMgr := auth.NewMemorySessionManager()
	authSvc := auth.NewAuthService(userRepo, sessionMgr)

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := order.NewOrderRepository(db)
	orderLogRepo := order.NewOrderLogRepository(db)

	orderSvc := order.NewOrderService(orderRepo, flowerRepo, orderLogRepo)

	handler := &Handler{
		authService:  authSvc,
		orderService: orderSvc,
	}

	return handler, db
}

// insertTestData 插入测试数据（用户、地址、鲜花）并返回用户ID
func insertTestData(t *testing.T, db *sql.DB) (userID, addressID int) {
	t.Helper()

	ctx := context.Background()

	userRepo := user.NewMySQLUserRepository(db)
	addressRepo := address.NewAddressRepository(db)
	flowerRepo := flower.NewFlowerRepository(db)

	// 创建用户
	u := &user.User{
		Username:     "testuser",
		PasswordHash: "hash",
		Role:         user.RoleCustomer,
	}
	if err := userRepo.Create(ctx, u); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// 创建地址
	addr := &address.Address{
		UserID:  u.ID,
		Label:   "家",
		Address: "北京市朝阳区",
		Contact: "张三",
	}
	if err := addressRepo.Create(ctx, addr); err != nil {
		t.Fatalf("failed to create address: %v", err)
	}

	// 创建鲜花
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
	if err := flowerRepo.Create(ctx, flw); err != nil {
		t.Fatalf("failed to create flower: %v", err)
	}

	return u.ID, addr.ID
}

// loginUser 通过 handler 登录并返回 session token
func loginUser(t *testing.T, h *Handler, username, password string) string {
	t.Helper()

	// 先注册
	registerReq := RegisterRequest{
		Username: username,
		Password: password,
	}
	body, _ := json.Marshal(registerReq)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleRegister(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("registration failed: status = %d, body = %s", w.Code, w.Body.String())
	}

	// 再登录
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}
	body, _ = json.Marshal(loginReq)

	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.HandleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: status = %d", w.Code)
	}

	// 获取 session token
	for _, c := range w.Result().Cookies() {
		if c.Name == "session_token" {
			return c.Value
		}
	}

	t.Fatal("login failed to set session_token cookie")
	return ""
}

// TestHandleCreateOrder_Success 测试成功创建订单
func TestHandleCreateOrder_Success(t *testing.T) {
	handler, db := setupOrderTestHandler(t)
	ctx := context.Background()

	// 登录获取 token（这会注册一个新用户）
	sessionToken := loginUser(t, handler, "testuser", "password123")

	// 获取用户ID
	userRepo := user.NewMySQLUserRepository(db)
	u, _ := userRepo.GetByUsername(ctx, "testuser")

	// 为用户创建地址和鲜花
	addressRepo := address.NewAddressRepository(db)
	flowerRepo := flower.NewFlowerRepository(db)

	addr := &address.Address{
		UserID:  u.ID,
		Label:   "家",
		Address: "北京市朝阳区",
		Contact: "张三",
	}
	if err := addressRepo.Create(ctx, addr); err != nil {
		t.Fatalf("failed to create address: %v", err)
	}

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
	if err := flowerRepo.Create(ctx, flw); err != nil {
		t.Fatalf("failed to create flower: %v", err)
	}

	// 创建订单请求
	createReq := CreateOrderRequest{
		AddressID: addr.ID,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 5},
		},
	}
	body, _ := json.Marshal(createReq)

	req := httptest.NewRequest("POST", "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	w := httptest.NewRecorder()

	handler.HandleCreateOrder(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("HandleCreateOrder() status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	// 验证响应
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["order_no"] == nil {
		t.Error("response missing order_no")
	}
}

// TestHandleCreateOrder_Unauthorized 测试未授权创建订单
func TestHandleCreateOrder_Unauthorized(t *testing.T) {
	handler, _ := setupOrderTestHandler(t)

	createReq := CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 1},
		},
	}
	body, _ := json.Marshal(createReq)

	req := httptest.NewRequest("POST", "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleCreateOrder(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HandleCreateOrder() unauthorized status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestHandleCreateOrder_EmptyItems 测试空订单项
func TestHandleCreateOrder_EmptyItems(t *testing.T) {
	handler, _ := setupOrderTestHandler(t)

	sessionToken := loginUser(t, handler, "testuser", "password123")

	createReq := CreateOrderRequest{
		AddressID: 1,
		Items:     []*CreateOrderItemRequest{},
	}
	body, _ := json.Marshal(createReq)

	req := httptest.NewRequest("POST", "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	w := httptest.NewRecorder()

	handler.HandleCreateOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleCreateOrder() empty items status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestHandleCreateOrder_InvalidJSON 测试无效 JSON
func TestHandleCreateOrder_InvalidJSON(t *testing.T) {
	handler, _ := setupOrderTestHandler(t)

	sessionToken := loginUser(t, handler, "testuser", "password123")

	req := httptest.NewRequest("POST", "/api/orders", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	w := httptest.NewRecorder()

	handler.HandleCreateOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleCreateOrder() invalid json status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestHandleGetOrder_Success 测试成功获取订单
func TestHandleGetOrder_Success(t *testing.T) {
	handler, db := setupOrderTestHandler(t)
	ctx := context.Background()

	// 登录获取 token
	sessionToken := loginUser(t, handler, "testuser", "password123")

	// 获取用户ID
	userRepo := user.NewMySQLUserRepository(db)
	u, _ := userRepo.GetByUsername(ctx, "testuser")

	// 为用户创建地址和鲜花
	addressRepo := address.NewAddressRepository(db)
	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := order.NewOrderRepository(db)
	orderLogRepo := order.NewOrderLogRepository(db)
	orderSvc := order.NewOrderService(orderRepo, flowerRepo, orderLogRepo)

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

	// 通过服务层创建订单
	createReq := &order.CreateOrderRequest{
		AddressID: addr.ID,
		Items: []*order.CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 5},
		},
	}
	orderNo, _ := orderSvc.CreateOrder(ctx, u.ID, createReq)

	// 获取订单
	req := httptest.NewRequest("GET", "/api/orders/"+orderNo, nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	w := httptest.NewRecorder()

	handler.HandleGetOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleGetOrder() status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	// 验证响应
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["order_no"] != orderNo {
		t.Errorf("order_no = %v, want %s", resp["order_no"], orderNo)
	}
}

// TestHandleGetOrder_NotFound 测试获取不存在的订单
func TestHandleGetOrder_NotFound(t *testing.T) {
	handler, _ := setupOrderTestHandler(t)

	sessionToken := loginUser(t, handler, "testuser", "password123")

	req := httptest.NewRequest("GET", "/api/orders/NONEXIST123", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	w := httptest.NewRecorder()

	handler.HandleGetOrder(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("HandleGetOrder() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// TestHandleGetOrder_Unauthorized 测试用户无权访问其他用户的订单
func TestHandleGetOrder_Unauthorized(t *testing.T) {
	handler, db := setupOrderTestHandler(t)
	ctx := context.Background()

	// 创建用户1（通过 insertTestData）
	userID1, addressID := insertTestData(t, db)

	// 通过服务层为用户1创建订单
	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := order.NewOrderRepository(db)
	orderLogRepo := order.NewOrderLogRepository(db)
	orderSvc := order.NewOrderService(orderRepo, flowerRepo, orderLogRepo)

	createReq := &order.CreateOrderRequest{
		AddressID: addressID,
		Items: []*order.CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 5},
		},
	}
	orderNo, _ := orderSvc.CreateOrder(ctx, userID1, createReq)

	// 用户2登录尝试获取用户1的订单
	sessionToken := loginUser(t, handler, "user2", "password123")

	req := httptest.NewRequest("GET", "/api/orders/"+orderNo, nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	w := httptest.NewRecorder()

	handler.HandleGetOrder(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("HandleGetOrder() unauthorized status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

// TestHandleListOrders_Success 测试成功获取订单列表
func TestHandleListOrders_Success(t *testing.T) {
	handler, db := setupOrderTestHandler(t)
	ctx := context.Background()

	// 登录用户
	sessionToken := loginUser(t, handler, "testuser", "password123")

	// 获取用户ID
	userRepo := user.NewMySQLUserRepository(db)
	u, _ := userRepo.GetByUsername(ctx, "testuser")

	// 为用户创建地址
	addressRepo := address.NewAddressRepository(db)
	addr := &address.Address{
		UserID:  u.ID,
		Label:   "家",
		Address: "北京市朝阳区",
		Contact: "张三",
	}
	addressRepo.Create(ctx, addr)

	// 创建鲜花
	flowerRepo := flower.NewFlowerRepository(db)
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

	// 创建订单服务并创建多个订单
	orderRepo := order.NewOrderRepository(db)
	orderLogRepo := order.NewOrderLogRepository(db)
	orderSvc := order.NewOrderService(orderRepo, flowerRepo, orderLogRepo)

	// 创建多个订单
	for i := 0; i < 3; i++ {
		createReq := &order.CreateOrderRequest{
			AddressID: addr.ID,
			Items: []*order.CreateOrderItemRequest{
				{FlowerSKU: "FLW001", Quantity: 1},
			},
		}
		orderSvc.CreateOrder(ctx, u.ID, createReq)
	}

	// 获取订单列表
	req := httptest.NewRequest("GET", "/api/orders?page=1&page_size=10", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	w := httptest.NewRecorder()

	handler.HandleListOrders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleListOrders() status = %d, want %d", w.Code, http.StatusOK)
	}

	// 验证响应
	var resp []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp) < 3 {
		t.Errorf("HandleListOrders() count = %d, want >= 3", len(resp))
	}
}

// TestHandleListOrders_Unauthorized 测试未授权获取订单列表
func TestHandleListOrders_Unauthorized(t *testing.T) {
	handler, _ := setupOrderTestHandler(t)

	req := httptest.NewRequest("GET", "/api/orders", nil)
	w := httptest.NewRecorder()

	handler.HandleListOrders(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HandleListOrders() unauthorized status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
