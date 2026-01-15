package order

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动

	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
)

// setupOrderTestDB 创建测试数据库连接（包含订单表）
func setupOrderTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// 创建订单表
	createOrderTableSQL := `
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
	createOrderItemTableSQL := `
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

	if _, err := db.Exec(createOrderTableSQL); err != nil {
		db.Close()
		t.Fatalf("failed to create orders table: %v", err)
	}

	if _, err := db.Exec(createOrderItemTableSQL); err != nil {
		db.Close()
		t.Fatalf("failed to create order_items table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestOrderRepository_Create 测试创建订单
func TestOrderRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupOrderTestDB(t)
	repo := NewOrderRepository(db)

	items := []*OrderItem{
		NewOrderItem(0, "FLW001", "红玫瑰", 10, 1000),
		NewOrderItem(0, "FLW002", "白百合", 5, 1500),
	}

	order := NewOrder(1, 1)
	order.TotalAmount = flower.Decimal{Value: 17500} // 10*1000 + 5*1500

	ctx := context.Background()
	err := repo.Create(ctx, order, items)

	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if order.ID <= 0 {
		t.Errorf("Create() ID = %d, want > 0", order.ID)
	}

	if len(items) == 0 || items[0].OrderID != order.ID {
		t.Error("Create() items should have OrderID set")
	}
}

// TestOrderRepository_GetByID 测试根据 ID 获取订单
func TestOrderRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupOrderTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	// 创建测试订单
	items := []*OrderItem{
		NewOrderItem(0, "FLW001", "红玫瑰", 10, 1000),
	}
	order := NewOrder(1, 1)
	order.TotalAmount = flower.Decimal{Value: 10000}

	if err := repo.Create(ctx, order, items); err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "get existing order",
			id:      order.ID,
			wantErr: false,
		},
		{
			name:    "get non-existing order",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultOrder, resultItems, err := repo.GetByID(ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if resultOrder == nil {
					t.Error("GetByID() order is nil")
					return
				}
				if resultOrder.ID != tt.id {
					t.Errorf("GetByID() ID = %d, want %d", resultOrder.ID, tt.id)
				}
				if len(resultItems) == 0 {
					t.Error("GetByID() should return items")
				}
			}
		})
	}
}

// TestOrderRepository_GetByOrderNo 测试根据订单号获取订单
func TestOrderRepository_GetByOrderNo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupOrderTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	// 创建测试订单
	items := []*OrderItem{
		NewOrderItem(0, "FLW001", "红玫瑰", 10, 1000),
	}
	order := NewOrder(1, 1)
	order.TotalAmount = flower.Decimal{Value: 10000}

	if err := repo.Create(ctx, order, items); err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	tests := []struct {
		name    string
		orderNo string
		wantErr bool
	}{
		{
			name:    "get existing order",
			orderNo: order.OrderNo,
			wantErr: false,
		},
		{
			name:    "get non-existing order",
			orderNo: "NONEXIST123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultOrder, resultItems, err := repo.GetByOrderNo(ctx, tt.orderNo)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByOrderNo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if resultOrder == nil {
					t.Error("GetByOrderNo() order is nil")
					return
				}
				if resultOrder.OrderNo != tt.orderNo {
					t.Errorf("GetByOrderNo() OrderNo = %s, want %s", resultOrder.OrderNo, tt.orderNo)
				}
				if len(resultItems) == 0 {
					t.Error("GetByOrderNo() should return items")
				}
			}
		})
	}
}

// TestOrderRepository_List 测试获取订单列表
func TestOrderRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupOrderTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	// 创建多个测试订单
	for i := 1; i <= 3; i++ {
		items := []*OrderItem{
			NewOrderItem(0, fmt.Sprintf("FLW%03d", i), fmt.Sprintf("鲜花%d", i), 10, 1000),
		}
		order := NewOrder(1, 1)
		order.TotalAmount = flower.Decimal{Value: 10000}

		if err := repo.Create(ctx, order, items); err != nil {
			t.Fatalf("failed to create order %d: %v", i, err)
		}
	}

	filter := OrderFilter{
		UserID:   1,
		Page:     1,
		PageSize: 10,
	}

	orders, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(orders) < 3 {
		t.Errorf("List() count = %d, want >= 3", len(orders))
	}
}

// TestOrderRepository_UpdateStatus 测试更新订单状态
func TestOrderRepository_UpdateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupOrderTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	// 创建测试订单
	items := []*OrderItem{
		NewOrderItem(0, "FLW001", "红玫瑰", 10, 1000),
	}
	order := NewOrder(1, 1)
	order.TotalAmount = flower.Decimal{Value: 10000}

	if err := repo.Create(ctx, order, items); err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	// 更新状态
	err := repo.UpdateStatus(ctx, order.ID, StatusCompleted)
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	// 验证更新
	updatedOrder, _, err := repo.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetByID() after UpdateStatus error = %v", err)
	}

	if updatedOrder.Status != StatusCompleted {
		t.Errorf("UpdateStatus() Status = %s, want %s", updatedOrder.Status, StatusCompleted)
	}
}
