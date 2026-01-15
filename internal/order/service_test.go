package order

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动

	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
)

// setupServiceTestDB 创建测试数据库连接（包含所有必需表）
func setupServiceTestDB(t *testing.T) *sql.DB {
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
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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
			shelf_life INTEGER NOT NULL,
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

// insertTestFlower 插入测试鲜花数据
func insertTestFlower(t *testing.T, db *sql.DB, sku, name string, salePrice, stock int) {
	t.Helper()
	query := `INSERT INTO flowers (sku, name, origin, shelf_life, preservation, purchase_price, sale_price, stock)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, sku, name, "云南", 7, "冷藏", 5000, salePrice, stock)
	if err != nil {
		t.Fatalf("failed to insert test flower: %v", err)
	}
}

// insertTestUser 插入测试用户
func insertTestUser(t *testing.T, db *sql.DB, id int, username string) {
	t.Helper()
	query := `INSERT INTO users (id, username, password_hash, role) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, id, username, "hash", "customer")
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
}

// insertTestAddress 插入测试地址
func insertTestAddress(t *testing.T, db *sql.DB, id, userID int) {
	t.Helper()
	query := `INSERT INTO addresses (id, user_id, label, address, contact) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, id, userID, "家", "北京市朝阳区", "张三")
	if err != nil {
		t.Fatalf("failed to insert test address: %v", err)
	}
}

// TestOrderService_CreateOrder_Success 测试成功创建订单
func TestOrderService_CreateOrder_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	// 插入测试数据
	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 100) // 单价10元，库存100
	insertTestFlower(t, db, "FLW002", "白百合", 1500, 50)  // 单价15元，库存50

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	req := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 10},
			{FlowerSKU: "FLW002", Quantity: 5},
		},
	}

	orderNo, err := service.CreateOrder(ctx, 1, req)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	if orderNo == "" {
		t.Error("CreateOrder() orderNo is empty")
	}

	// 验证订单已创建
	order, items, err := orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		t.Fatalf("GetByOrderNo() error = %v", err)
	}

	if order.UserID != 1 {
		t.Errorf("CreateOrder() UserID = %d, want 1", order.UserID)
	}

	// 验证订单项
	if len(items) != 2 {
		t.Errorf("CreateOrder() items count = %d, want 2", len(items))
	}

	// 验证总金额: 10*1000 + 5*1500 = 17500 (175.00元)
	expectedTotal := int64(17500)
	if order.TotalAmount.Value != expectedTotal {
		t.Errorf("CreateOrder() TotalAmount = %d, want %d", order.TotalAmount.Value, expectedTotal)
	}

	// 验证库存已扣减
	flw1, err := flowerRepo.GetBySKU(ctx, "FLW001")
	if err != nil {
		t.Fatalf("GetBySKU() error = %v", err)
	}
	if flw1.Stock != 90 { // 100 - 10 = 90
		t.Errorf("FLW001 stock = %d, want 90", flw1.Stock)
	}

	flw2, err := flowerRepo.GetBySKU(ctx, "FLW002")
	if err != nil {
		t.Fatalf("GetBySKU() error = %v", err)
	}
	if flw2.Stock != 45 { // 50 - 5 = 45
		t.Errorf("FLW002 stock = %d, want 45", flw2.Stock)
	}

	// 验证订单日志
	logs, err := logRepo.GetLogs(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}
	if len(logs) == 0 {
		t.Error("CreateOrder() should create order log")
	}
}

// TestOrderService_CreateOrder_InsufficientStock 测试库存不足时创建订单失败
func TestOrderService_CreateOrder_InsufficientStock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	// 插入测试数据
	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 5) // 库存只有5

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	req := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 10}, // 需要10个，库存不足
		},
	}

	_, err := service.CreateOrder(ctx, 1, req)
	if err == nil {
		t.Error("CreateOrder() should fail when stock is insufficient")
	}

	// 验证错误信息
	expectedErrMsg := "库存不足"
	if err != nil && err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
		t.Errorf("CreateOrder() error = %v, should contain %v", err, expectedErrMsg)
	}
}

// TestOrderService_CreateOrder_FlowerNotFound 测试鲜花不存在时创建订单失败
func TestOrderService_CreateOrder_FlowerNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	// 插入测试数据
	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)
	// 不插入鲜花

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	req := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "NONEXIST", Quantity: 1},
		},
	}

	_, err := service.CreateOrder(ctx, 1, req)
	if err == nil {
		t.Error("CreateOrder() should fail when flower not found")
	}
}

// TestOrderService_CreateOrder_EmptyItems 测试空订单项时创建订单失败
func TestOrderService_CreateOrder_EmptyItems(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	insertTestUser(t, db, 1, "user1")

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	req := &CreateOrderRequest{
		AddressID: 1,
		Items:     []*CreateOrderItemRequest{},
	}

	_, err := service.CreateOrder(ctx, 1, req)
	if err == nil {
		t.Error("CreateOrder() should fail when items is empty")
	}
}

// TestOrderService_CreateOrder_InvalidQuantity 测试无效数量时创建订单失败
func TestOrderService_CreateOrder_InvalidQuantity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 100)

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	req := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 0}, // 无效数量
		},
	}

	_, err := service.CreateOrder(ctx, 1, req)
	if err == nil {
		t.Error("CreateOrder() should fail when quantity is invalid")
	}
}

// TestOrderService_GetOrder_Success 测试成功获取订单
func TestOrderService_GetOrder_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	// 插入测试数据
	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 100)

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	// 先创建订单
	createReq := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 5},
		},
	}

	orderNo, err := service.CreateOrder(ctx, 1, createReq)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	// 获取订单
	response, err := service.GetOrder(ctx, 1, orderNo)
	if err != nil {
		t.Fatalf("GetOrder() error = %v", err)
	}

	if response.OrderNo != orderNo {
		t.Errorf("GetOrder() OrderNo = %s, want %s", response.OrderNo, orderNo)
	}

	if response.UserID != 1 {
		t.Errorf("GetOrder() UserID = %d, want 1", response.UserID)
	}

	if len(response.Items) != 1 {
		t.Errorf("GetOrder() items count = %d, want 1", len(response.Items))
	}
}

// TestOrderService_GetOrder_NotFound 测试获取不存在的订单
func TestOrderService_GetOrder_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	_, err := service.GetOrder(ctx, 1, "NONEXIST123")
	if err == nil {
		t.Error("GetOrder() should fail when order not found")
	}
}

// TestOrderService_GetOrder_Unauthorized 测试用户无权访问其他用户的订单
func TestOrderService_GetOrder_Unauthorized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	// 插入两个用户
	insertTestUser(t, db, 1, "user1")
	insertTestUser(t, db, 2, "user2")
	insertTestAddress(t, db, 1, 1)
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 100)

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	// user1 创建订单
	createReq := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 5},
		},
	}

	orderNo, err := service.CreateOrder(ctx, 1, createReq)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	// user2 尝试获取 user1 的订单
	_, err = service.GetOrder(ctx, 2, orderNo)
	if err == nil {
		t.Error("GetOrder() should fail when user is not authorized")
	}
}

// TestOrderService_ListOrders_Success 测试成功获取订单列表
func TestOrderService_ListOrders_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	// 插入测试数据
	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 100)

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	// 创建多个订单
	for i := 0; i < 3; i++ {
		req := &CreateOrderRequest{
			AddressID: 1,
			Items: []*CreateOrderItemRequest{
				{FlowerSKU: "FLW001", Quantity: 1},
			},
		}
		if _, err := service.CreateOrder(ctx, 1, req); err != nil {
			t.Fatalf("CreateOrder() error = %v", err)
		}
	}

	// 获取订单列表
	filter := OrderListFilter{
		Page:     1,
		PageSize: 10,
	}

	orders, err := service.ListOrders(ctx, 1, filter)
	if err != nil {
		t.Fatalf("ListOrders() error = %v", err)
	}

	if len(orders) < 3 {
		t.Errorf("ListOrders() count = %d, want >= 3", len(orders))
	}

	// 验证所有订单都属于用户1
	for _, o := range orders {
		if o.UserID != 1 {
			t.Errorf("ListOrders() returned order with UserID = %d, want 1", o.UserID)
		}
	}
}

// TestOrderService_ListOrders_WithStatusFilter 测试按状态筛选订单
func TestOrderService_ListOrders_WithStatusFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 100)

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	// 创建订单
	req := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 1},
		},
	}

	if _, err := service.CreateOrder(ctx, 1, req); err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	// 按待处理状态筛选
	filter := OrderListFilter{
		Status:   "pending",
		Page:     1,
		PageSize: 10,
	}

	orders, err := service.ListOrders(ctx, 1, filter)
	if err != nil {
		t.Fatalf("ListOrders() error = %v", err)
	}

	if len(orders) == 0 {
		t.Error("ListOrders() should return pending orders")
	}

	// 验证所有订单都是待处理状态
	for _, o := range orders {
		if o.Status != "pending" {
			t.Errorf("ListOrders() returned order with Status = %s, want pending", o.Status)
		}
	}
}

// TestOrderService_CreateOrder_TransactionRollback 测试事务回滚
func TestOrderService_CreateOrder_TransactionRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 100)
	insertTestFlower(t, db, "FLW002", "白百合", 1500, 5) // 库存只有5

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	// 记录初始库存
	flw1Before, _ := flowerRepo.GetBySKU(ctx, "FLW001")
	stock1Before := flw1Before.Stock

	flw2Before, _ := flowerRepo.GetBySKU(ctx, "FLW002")
	stock2Before := flw2Before.Stock

	// 创建订单，第二个鲜花库存不足
	req := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 10},
			{FlowerSKU: "FLW002", Quantity: 10}, // 库存不足
		},
	}

	_, err := service.CreateOrder(ctx, 1, req)
	if err == nil {
		t.Error("CreateOrder() should fail when stock is insufficient")
	}

	// 验证库存没有变化（事务回滚）
	flw1After, _ := flowerRepo.GetBySKU(ctx, "FLW001")
	if flw1After.Stock != stock1Before {
		t.Errorf("FLW001 stock changed from %d to %d, should be unchanged", stock1Before, flw1After.Stock)
	}

	flw2After, _ := flowerRepo.GetBySKU(ctx, "FLW002")
	if flw2After.Stock != stock2Before {
		t.Errorf("FLW002 stock changed from %d to %d, should be unchanged", stock2Before, flw2After.Stock)
	}
}

// TestOrderService_CreateOrder_InactiveFlower 测试购买已下架鲜花失败
func TestOrderService_CreateOrder_InactiveFlower(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)

	// 插入已下架的鲜花
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 100)
	db.Exec("UPDATE flowers SET is_active = 0 WHERE sku = ?", "FLW001")

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	req := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 1},
		},
	}

	_, err := service.CreateOrder(ctx, 1, req)
	if err == nil {
		t.Error("CreateOrder() should fail when flower is inactive")
	}
}
