package flower

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
)

// setupTestService 创建测试服务和数据库
func setupTestService(t *testing.T) (FlowerService, *sql.DB) {
	t.Helper()

	db := setupTestDB(t)
	repo := NewFlowerRepository(db)
	service := NewFlowerService(repo)

	return service, db
}

// TestFlowerService_CreateFlower 测试创建鲜花
func TestFlowerService_CreateFlower(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		request *CreateFlowerRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "create valid flower",
			request: &CreateFlowerRequest{
				SKU:          "SVC001",
				Name:         "红玫瑰",
				Origin:       "云南",
				ShelfLife:    "7天",
				Preservation: "常温",
				PurchasePrice: 50.00,
				SalePrice:     100.00,
				Stock:         100,
			},
			wantErr: false,
		},
		{
			name: "create with empty SKU",
			request: &CreateFlowerRequest{
				SKU:          "",
				Name:         "红玫瑰",
				Origin:       "云南",
				ShelfLife:    "7天",
				Preservation: "常温",
				PurchasePrice: 50.00,
				SalePrice:     100.00,
				Stock:         100,
			},
			wantErr: true,
			errMsg:  "SKU不能为空",
		},
		{
			name: "create with sale price below purchase price",
			request: &CreateFlowerRequest{
				SKU:          "SVC002",
				Name:         "红玫瑰",
				Origin:       "云南",
				ShelfLife:    "7天",
				Preservation: "常温",
				PurchasePrice: 100.00,
				SalePrice:     50.00,
				Stock:         100,
			},
			wantErr: true,
			errMsg:  "销售价格不能低于进货价格",
		},
		{
			name: "create with negative stock",
			request: &CreateFlowerRequest{
				SKU:          "SVC003",
				Name:         "红玫瑰",
				Origin:       "云南",
				ShelfLife:    "7天",
				Preservation: "常温",
				PurchasePrice: 50.00,
				SalePrice:     100.00,
				Stock:         -10,
			},
			wantErr: true,
			errMsg:  "库存不能为负数",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateFlower(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFlower() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("CreateFlower() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestFlowerService_GetFlower 测试获取鲜花
func TestFlowerService_GetFlower(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试鲜花
	createReq := &CreateFlowerRequest{
		SKU:          "GET001",
		Name:         "向日葵",
		Origin:       "美国",
		ShelfLife:    "10天",
		Preservation: "常温",
		PurchasePrice: 30.00,
		SalePrice:     60.00,
		Stock:         200,
	}
	if err := service.CreateFlower(ctx, createReq); err != nil {
		t.Fatalf("failed to create flower: %v", err)
	}

	tests := []struct {
		name    string
		sku     string
		wantErr bool
	}{
		{
			name:    "get existing flower",
			sku:     "GET001",
			wantErr: false,
		},
		{
			name:    "get non-existing flower",
			sku:     "NONEXIST",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flower, err := service.GetFlower(ctx, tt.sku)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetFlower() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if flower == nil {
					t.Error("GetFlower() flower is nil")
					return
				}
				if flower.SKU != tt.sku {
					t.Errorf("GetFlower() SKU = %q, want %q", flower.SKU, tt.sku)
				}
			}
		})
	}
}

// TestFlowerService_ListFlowers 测试获取鲜花列表
func TestFlowerService_ListFlowers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 创建多个测试鲜花
	flowers := []*CreateFlowerRequest{
		{
			SKU:          "LST001",
			Name:         "红玫瑰",
			Origin:       "云南",
			ShelfLife:    "7天",
			Preservation: "常温",
			PurchasePrice: 50.00,
			SalePrice:     100.00,
			Stock:         100,
		},
		{
			SKU:          "LST002",
			Name:         "白百合",
			Origin:       "荷兰",
			ShelfLife:    "14天",
			Preservation: "低温",
			PurchasePrice: 80.00,
			SalePrice:     150.00,
			Stock:         50,
		},
		{
			SKU:          "LST003",
			Name:         "康乃馨",
			Origin:       "云南",
			ShelfLife:    "10天",
			Preservation: "常温",
			PurchasePrice: 40.00,
			SalePrice:     80.00,
			Stock:         5, // 低库存
		},
	}

	for _, f := range flowers {
		if err := service.CreateFlower(ctx, f); err != nil {
			t.Fatalf("failed to create flower: %v", err)
		}
	}

	tests := []struct {
		name     string
		filter   FlowerFilter
		minCount int
		maxCount int
		wantErr  bool
	}{
		{
			name:     "list all flowers",
			filter:   FlowerFilter{},
			minCount: 3,
			maxCount: 3,
			wantErr:  false,
		},
		{
			name: "list with search",
			filter: FlowerFilter{
				Search: "玫瑰",
			},
			minCount: 1,
			maxCount: 1,
			wantErr:  false,
		},
		{
			name: "list with origin filter",
			filter: FlowerFilter{
				Origin: "云南",
			},
			minCount: 2,
			maxCount: 2,
			wantErr:  false,
		},
		{
			name: "list with price range",
			filter: FlowerFilter{
				MinPrice: 50,
				MaxPrice: 120,
			},
			minCount: 2,
			maxCount: 2,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ListFlowers(ctx, tt.filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListFlowers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) < tt.minCount || len(result) > tt.maxCount {
					t.Errorf("ListFlowers() count = %d, want between %d and %d", len(result), tt.minCount, tt.maxCount)
				}
			}
		})
	}
}

// TestFlowerService_UpdateFlower 测试更新鲜花
func TestFlowerService_UpdateFlower(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试鲜花
	createReq := &CreateFlowerRequest{
		SKU:          "UPD001",
		Name:         "红玫瑰",
		Origin:       "云南",
		ShelfLife:    "7天",
		Preservation: "常温",
		PurchasePrice: 50.00,
		SalePrice:     100.00,
		Stock:         100,
	}
	if err := service.CreateFlower(ctx, createReq); err != nil {
		t.Fatalf("failed to create flower: %v", err)
	}

	tests := []struct {
		name    string
		sku     string
		request *UpdateFlowerRequest
		wantErr bool
	}{
		{
			name: "update valid flower",
			sku:  "UPD001",
			request: &UpdateFlowerRequest{
				Name:         stringPtr("红玫瑰（特大）"),
				Origin:       stringPtr("云南"),
				ShelfLife:    stringPtr("7天"),
				Preservation: stringPtr("常温"),
				PurchasePrice: float64Ptr(60.00),
				SalePrice:     float64Ptr(120.00),
			},
			wantErr: false,
		},
		{
			name: "update non-existing flower",
			sku:  "NONEXIST",
			request: &UpdateFlowerRequest{
				Name:         stringPtr("红玫瑰"),
				Origin:       stringPtr("云南"),
				ShelfLife:    stringPtr("7天"),
				Preservation: stringPtr("常温"),
			},
			wantErr: true,
		},
		{
			name: "update with invalid sale price",
			sku:  "UPD001",
			request: &UpdateFlowerRequest{
				Name:         stringPtr("红玫瑰"),
				Origin:       stringPtr("云南"),
				ShelfLife:    stringPtr("7天"),
				Preservation: stringPtr("常温"),
				PurchasePrice: float64Ptr(100.00),
				SalePrice:     float64Ptr(50.00),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateFlower(ctx, tt.sku, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateFlower() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFlowerService_DeleteFlower 测试删除鲜花
func TestFlowerService_DeleteFlower(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试鲜花
	createReq := &CreateFlowerRequest{
		SKU:          "DEL001",
		Name:         "红玫瑰",
		Origin:       "云南",
		ShelfLife:    "7天",
		Preservation: "常温",
		PurchasePrice: 50.00,
		SalePrice:     100.00,
		Stock:         100,
	}
	if err := service.CreateFlower(ctx, createReq); err != nil {
		t.Fatalf("failed to create flower: %v", err)
	}

	tests := []struct {
		name    string
		sku     string
		wantErr bool
	}{
		{
			name:    "delete existing flower",
			sku:     "DEL001",
			wantErr: false,
		},
		{
			name:    "delete non-existing flower",
			sku:     "NONEXIST",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteFlower(ctx, tt.sku)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteFlower() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 验证删除后鲜花不存在
			if !tt.wantErr {
				_, err := service.GetFlower(ctx, tt.sku)
				if err == nil {
					t.Error("DeleteFlower() flower still exists after deletion")
				}
			}
		})
	}
}

// TestFlowerService_AddStock 测试进货入库
func TestFlowerService_AddStock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试鲜花
	createReq := &CreateFlowerRequest{
		SKU:          "STK001",
		Name:         "红玫瑰",
		Origin:       "云南",
		ShelfLife:    "7天",
		Preservation: "常温",
		PurchasePrice: 50.00,
		SalePrice:     100.00,
		Stock:         100,
	}
	if err := service.CreateFlower(ctx, createReq); err != nil {
		t.Fatalf("failed to create flower: %v", err)
	}

	tests := []struct {
		name       string
		sku        string
		quantity   int
		wantStock  int
		wantErr    bool
	}{
		{
			name:      "add stock - positive quantity",
			sku:       "STK001",
			quantity:  50,
			wantStock: 150,
			wantErr:   false,
		},
		{
			name:      "add stock - zero quantity",
			sku:       "STK001",
			quantity:  0,
			wantStock: 150,
			wantErr:   false,
		},
		{
			name:      "add stock - non-existing flower",
			sku:       "NONEXIST",
			quantity:  50,
			wantStock: 0,
			wantErr:   true,
		},
		{
			name:      "add stock - negative quantity",
			sku:       "STK001",
			quantity:  -10,
			wantStock: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.AddStock(ctx, tt.sku, tt.quantity)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddStock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 验证库存已更新
				flower, err := service.GetFlower(ctx, tt.sku)
				if err != nil {
					t.Errorf("GetFlower() after AddStock error = %v", err)
					return
				}
				if flower.Stock != tt.wantStock {
					t.Errorf("AddStock() Stock = %d, want %d", flower.Stock, tt.wantStock)
				}
			}
		})
	}
}

// TestFlowerService_LowStockWarning 测试库存预警
func TestFlowerService_LowStockWarning(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试鲜花（包括低库存的）
	flowers := []*CreateFlowerRequest{
		{
			SKU:          "LOW001",
			Name:         "红玫瑰",
			Origin:       "云南",
			ShelfLife:    "7天",
			Preservation: "常温",
			PurchasePrice: 50.00,
			SalePrice:     100.00,
			Stock:         100, // 正常库存
		},
		{
			SKU:          "LOW002",
			Name:         "白百合",
			Origin:       "荷兰",
			ShelfLife:    "14天",
			Preservation: "低温",
			PurchasePrice: 80.00,
			SalePrice:     150.00,
			Stock:         10, // 等于阈值
		},
		{
			SKU:          "LOW003",
			Name:         "康乃馨",
			Origin:       "哥伦比亚",
			ShelfLife:    "10天",
			Preservation: "常温",
			PurchasePrice: 40.00,
			SalePrice:     80.00,
			Stock:         5, // 低于阈值
		},
	}

	for _, f := range flowers {
		if err := service.CreateFlower(ctx, f); err != nil {
			t.Fatalf("failed to create flower: %v", err)
		}
	}

	// 获取鲜花列表并检查低库存标识
	result, err := service.ListFlowers(ctx, FlowerFilter{})
	if err != nil {
		t.Fatalf("ListFlowers() error = %v", err)
	}

	lowStockCount := 0
	for _, f := range result {
		if f.LowStock {
			lowStockCount++
		}
	}

	// 应该有2个低库存鲜花（LOW002 和 LOW003）
	if lowStockCount != 2 {
		t.Errorf("LowStockWarning() low stock count = %d, want %d", lowStockCount, 2)
	}
}

// 辅助函数
func float64Ptr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}
