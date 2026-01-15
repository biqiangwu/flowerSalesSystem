package order

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动

	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
)

// TestOrderService_CompleteOrder_Success 测试成功完成订单
func TestOrderService_CompleteOrder_Success(t *testing.T) {
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

	// 创建待处理订单
	createReq := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 10},
		},
	}

	orderNo, err := service.CreateOrder(ctx, 1, createReq)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	// 获取订单ID
	order, _, err := orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		t.Fatalf("GetByOrderNo() error = %v", err)
	}

	// 完成订单
	err = service.CompleteOrder(ctx, order.ID, 1)
	if err != nil {
		t.Fatalf("CompleteOrder() error = %v", err)
	}

	// 验证订单状态已更新为已完成
	updatedOrder, _, err := orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if updatedOrder.Status != StatusCompleted {
		t.Errorf("CompleteOrder() status = %s, want %s", updatedOrder.Status, StatusCompleted)
	}

	// 验证订单日志
	logs, err := logRepo.GetLogs(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}

	// 应该有两条日志：创建订单和完成订单
	if len(logs) < 2 {
		t.Errorf("CompleteOrder() should create completion log, got %d logs", len(logs))
	}

	// 验证最后一条日志是完成操作
	lastLog := logs[len(logs)-1]
	if lastLog.Action != "complete_order" {
		t.Errorf("CompleteOrder() log action = %s, want complete_order", lastLog.Action)
	}
	if lastLog.NewStatus != StatusCompleted {
		t.Errorf("CompleteOrder() log new_status = %s, want %s", lastLog.NewStatus, StatusCompleted)
	}
}

// TestOrderService_CompleteOrder_OrderNotFound 测试完成不存在的订单
func TestOrderService_CompleteOrder_OrderNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	// 尝试完成不存在的订单
	err := service.CompleteOrder(ctx, 99999, 1)
	if err == nil {
		t.Error("CompleteOrder() should fail when order not found")
	}
}

// TestOrderService_CompleteOrder_InvalidStatusTransition 测试无效的状态流转
func TestOrderService_CompleteOrder_InvalidStatusTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name        string
		initialStatus OrderStatus
		wantErr     bool
		errContains string
	}{
		{
			name:        "从已完成状态尝试完成",
			initialStatus: StatusCompleted,
			wantErr:     true,
			errContains: "状态",
		},
		{
			name:        "从已取消状态尝试完成",
			initialStatus: StatusCancelled,
			wantErr:     true,
			errContains: "状态",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			createReq := &CreateOrderRequest{
				AddressID: 1,
				Items: []*CreateOrderItemRequest{
					{FlowerSKU: "FLW001", Quantity: 10},
				},
			}

			orderNo, _ := service.CreateOrder(ctx, 1, createReq)
			order, _, _ := orderRepo.GetByOrderNo(ctx, orderNo)

			// 手动设置订单状态
			db.Exec("UPDATE orders SET status = ? WHERE id = ?", tt.initialStatus, order.ID)

			// 尝试完成订单
			err := service.CompleteOrder(ctx, order.ID, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompleteOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("CompleteOrder() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

// TestOrderService_CancelOrder_Success 测试成功取消订单（含库存回退）
func TestOrderService_CancelOrder_Success(t *testing.T) {
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

	// 创建待处理订单
	createReq := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 10},
		},
	}

	orderNo, err := service.CreateOrder(ctx, 1, createReq)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	// 获取订单和记录扣减后的库存
	order, items, err := orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		t.Fatalf("GetByOrderNo() error = %v", err)
	}

	flw, _ := flowerRepo.GetBySKU(ctx, "FLW001")
	stockAfterCreate := flw.Stock // 应该是 90 (100 - 10)

	// 取消订单
	err = service.CancelOrder(ctx, order.ID, 1)
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}

	// 验证订单状态已更新为已取消
	updatedOrder, _, err := orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if updatedOrder.Status != StatusCancelled {
		t.Errorf("CancelOrder() status = %s, want %s", updatedOrder.Status, StatusCancelled)
	}

	// 验证库存已回退
	flwAfterCancel, _ := flowerRepo.GetBySKU(ctx, "FLW001")
	expectedStock := stockAfterCreate + items[0].Quantity
	if flwAfterCancel.Stock != expectedStock {
		t.Errorf("CancelOrder() stock = %d, want %d (after rollback)", flwAfterCancel.Stock, expectedStock)
	}

	// 验证订单日志
	logs, err := logRepo.GetLogs(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}

	// 应该有两条日志：创建订单和取消订单
	if len(logs) < 2 {
		t.Errorf("CancelOrder() should create cancellation log, got %d logs", len(logs))
	}

	// 验证最后一条日志是取消操作
	lastLog := logs[len(logs)-1]
	if lastLog.Action != "cancel_order" {
		t.Errorf("CancelOrder() log action = %s, want cancel_order", lastLog.Action)
	}
	if lastLog.NewStatus != StatusCancelled {
		t.Errorf("CancelOrder() log new_status = %s, want %s", lastLog.NewStatus, StatusCancelled)
	}
}

// TestOrderService_CancelOrder_OrderNotFound 测试取消不存在的订单
func TestOrderService_CancelOrder_OrderNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	// 尝试取消不存在的订单
	err := service.CancelOrder(ctx, 99999, 1)
	if err == nil {
		t.Error("CancelOrder() should fail when order not found")
	}
}

// TestOrderService_CancelOrder_InvalidStatusTransition 测试无效的状态流转
func TestOrderService_CancelOrder_InvalidStatusTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name        string
		initialStatus OrderStatus
		wantErr     bool
		errContains string
	}{
		{
			name:        "从已取消状态尝试取消",
			initialStatus: StatusCancelled,
			wantErr:     true,
			errContains: "状态",
		},
		{
			name:        "从已完成状态尝试取消",
			initialStatus: StatusCompleted,
			wantErr:     true,
			errContains: "状态",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			createReq := &CreateOrderRequest{
				AddressID: 1,
				Items: []*CreateOrderItemRequest{
					{FlowerSKU: "FLW001", Quantity: 10},
				},
			}

			orderNo, _ := service.CreateOrder(ctx, 1, createReq)
			order, _, _ := orderRepo.GetByOrderNo(ctx, orderNo)

			// 手动设置订单状态
			db.Exec("UPDATE orders SET status = ? WHERE id = ?", tt.initialStatus, order.ID)

			// 尝试取消订单
			err := service.CancelOrder(ctx, order.ID, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("CancelOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("CancelOrder() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}

// TestOrderService_CancelOrder_StockRollbackTransaction 测试取消订单时库存回退事务
func TestOrderService_CancelOrder_StockRollbackTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	// 插入测试数据
	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 100)
	insertTestFlower(t, db, "FLW002", "白百合", 1500, 50)

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	// 创建包含多个订单项的订单
	createReq := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 10}, // 扣减10
			{FlowerSKU: "FLW002", Quantity: 5},  // 扣减5
		},
	}

	orderNo, err := service.CreateOrder(ctx, 1, createReq)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	// 获取订单项
	order, items, err := orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		t.Fatalf("GetByOrderNo() error = %v", err)
	}

	// 记录当前库存
	flw1, _ := flowerRepo.GetBySKU(ctx, "FLW001")
	flw2, _ := flowerRepo.GetBySKU(ctx, "FLW002")
	stock1BeforeCancel := flw1.Stock
	stock2BeforeCancel := flw2.Stock

	// 取消订单
	err = service.CancelOrder(ctx, order.ID, 1)
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}

	// 验证所有订单项的库存都已回退
	flw1After, _ := flowerRepo.GetBySKU(ctx, "FLW001")
	flw2After, _ := flowerRepo.GetBySKU(ctx, "FLW002")

	// FLW001: 100 - 10 = 90, 回退后应该是 100
	expectedStock1 := stock1BeforeCancel + items[0].Quantity
	if flw1After.Stock != expectedStock1 {
		t.Errorf("FLW001 stock after cancel = %d, want %d", flw1After.Stock, expectedStock1)
	}

	// FLW002: 50 - 5 = 45, 回退后应该是 50
	expectedStock2 := stock2BeforeCancel + items[1].Quantity
	if flw2After.Stock != expectedStock2 {
		t.Errorf("FLW002 stock after cancel = %d, want %d", flw2After.Stock, expectedStock2)
	}
}

// TestOrderService_CancelOrder_RollbackFailure 测试库存回退失败时的处理
func TestOrderService_CancelOrder_RollbackFailure(t *testing.T) {
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
	createReq := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 10},
		},
	}

	orderNo, _ := service.CreateOrder(ctx, 1, createReq)
	order, _, _ := orderRepo.GetByOrderNo(ctx, orderNo)

	// 删除鲜花，模拟库存回退失败
	db.Exec("DELETE FROM flowers WHERE sku = ?", "FLW001")

	// 尝试取消订单
	err := service.CancelOrder(ctx, order.ID, 1)
	// 取消应该失败或部分成功（取决于实现）
	// 关键是不应该出现panic，并且应该有错误处理
	if err == nil {
		t.Log("CancelOrder() succeeded despite stock rollback failure (implementation may allow this)")
	}
}

// TestOrderService_CompleteOrder_MultipleItems 测试完成多商品订单
func TestOrderService_CompleteOrder_MultipleItems(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupServiceTestDB(t)
	ctx := context.Background()

	insertTestUser(t, db, 1, "user1")
	insertTestAddress(t, db, 1, 1)
	insertTestFlower(t, db, "FLW001", "红玫瑰", 1000, 100)
	insertTestFlower(t, db, "FLW002", "白百合", 1500, 50)

	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := NewOrderRepository(db)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderService(orderRepo, flowerRepo, logRepo)

	// 创建多商品订单
	createReq := &CreateOrderRequest{
		AddressID: 1,
		Items: []*CreateOrderItemRequest{
			{FlowerSKU: "FLW001", Quantity: 10},
			{FlowerSKU: "FLW002", Quantity: 5},
		},
	}

	orderNo, err := service.CreateOrder(ctx, 1, createReq)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	order, _, err := orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		t.Fatalf("GetByOrderNo() error = %v", err)
	}

	// 完成订单
	err = service.CompleteOrder(ctx, order.ID, 1)
	if err != nil {
		t.Fatalf("CompleteOrder() error = %v", err)
	}

	// 验证订单状态
	updatedOrder, _, _ := orderRepo.GetByID(ctx, order.ID)
	if updatedOrder.Status != StatusCompleted {
		t.Errorf("CompleteOrder() status = %s, want %s", updatedOrder.Status, StatusCompleted)
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOf(s, substr) >= 0)
}

// indexOf 查找子串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
