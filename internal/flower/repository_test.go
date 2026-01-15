package flower

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
)

// setupTestDB 创建测试数据库连接
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// 使用 sqlite 作为测试数据库
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// 创建测试表
	createTableSQL := `
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

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		t.Fatalf("failed to create flowers table: %v", err)
	}

	// 清理函数
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestFlowerRepository_Create 测试 Create 方法
func TestFlowerRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewFlowerRepository(db)

	tests := []struct {
		name    string
		flower  *Flower
		wantErr bool
	}{
		{
			name: "create valid flower",
			flower: &Flower{
				SKU:           "FLW001",
				Name:          "红玫瑰",
				Origin:        "云南",
				ShelfLife:     "7天",
				Preservation:  "常温",
				PurchasePrice: Decimal{Value: 5000},
				SalePrice:     Decimal{Value: 10000},
				Stock:         100,
				IsActive:      true,
			},
			wantErr: false,
		},
		{
			name: "create another flower",
			flower: &Flower{
				SKU:           "FLW002",
				Name:          "白百合",
				Origin:        "荷兰",
				ShelfLife:     "14天",
				Preservation:  "低温",
				PurchasePrice: Decimal{Value: 8000},
				SalePrice:     Decimal{Value: 15000},
				Stock:         50,
				IsActive:      true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := repo.Create(ctx, tt.flower)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.flower.CreatedAt.IsZero() {
				t.Errorf("Create() CreatedAt is zero")
			}

			if !tt.wantErr && tt.flower.UpdatedAt.IsZero() {
				t.Errorf("Create() UpdatedAt is zero")
			}
		})
	}
}

// TestFlowerRepository_Create_DuplicateSKU 测试重复 SKU
func TestFlowerRepository_Create_DuplicateSKU(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewFlowerRepository(db)
	ctx := context.Background()

	// 创建第一个鲜花
	flower1 := &Flower{
		SKU:           "DUP001",
		Name:          "红玫瑰",
		Origin:        "云南",
		ShelfLife:     "7天",
		Preservation:  "常温",
		PurchasePrice: Decimal{Value: 5000},
		SalePrice:     Decimal{Value: 10000},
		Stock:         100,
		IsActive:      true,
	}
	if err := repo.Create(ctx, flower1); err != nil {
		t.Fatalf("failed to create first flower: %v", err)
	}

	// 尝试创建相同 SKU 的鲜花
	flower2 := &Flower{
		SKU:           "DUP001",
		Name:          "白玫瑰",
		Origin:        "云南",
		ShelfLife:     "7天",
		Preservation:  "常温",
		PurchasePrice: Decimal{Value: 6000},
		SalePrice:     Decimal{Value: 12000},
		Stock:         80,
		IsActive:      true,
	}
	err := repo.Create(ctx, flower2)
	if err == nil {
		t.Error("Create() expected error for duplicate SKU, got nil")
	}
}

// TestFlowerRepository_GetBySKU 测试 GetBySKU 方法
func TestFlowerRepository_GetBySKU(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewFlowerRepository(db)
	ctx := context.Background()

	// 创建测试鲜花
	createdFlower := &Flower{
		SKU:           "GET001",
		Name:          "向日葵",
		Origin:        "美国",
		ShelfLife:     "10天",
		Preservation:  "常温",
		PurchasePrice: Decimal{Value: 3000},
		SalePrice:     Decimal{Value: 6000},
		Stock:         200,
		IsActive:      true,
	}
	if err := repo.Create(ctx, createdFlower); err != nil {
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
			flower, err := repo.GetBySKU(ctx, tt.sku)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBySKU() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if flower == nil {
					t.Error("GetBySKU() flower is nil")
					return
				}
				if flower.SKU != tt.sku {
					t.Errorf("GetBySKU() SKU = %q, want %q", flower.SKU, tt.sku)
				}
			}
		})
	}
}

// TestFlowerRepository_List 测试 List 方法
func TestFlowerRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewFlowerRepository(db)
	ctx := context.Background()

	// 创建多个测试鲜花
	flowers := []*Flower{
		{
			SKU:           "LIST001",
			Name:          "红玫瑰",
			Origin:        "云南",
			ShelfLife:     "7天",
			Preservation:  "常温",
			PurchasePrice: Decimal{Value: 5000},
			SalePrice:     Decimal{Value: 10000},
			Stock:         100,
			IsActive:      true,
		},
		{
			SKU:           "LIST002",
			Name:          "白百合",
			Origin:        "荷兰",
			ShelfLife:     "14天",
			Preservation:  "低温",
			PurchasePrice: Decimal{Value: 8000},
			SalePrice:     Decimal{Value: 15000},
			Stock:         50,
			IsActive:      true,
		},
		{
			SKU:           "LIST003",
			Name:          "康乃馨",
			Origin:        "哥伦比亚",
			ShelfLife:     "10天",
			Preservation:  "常温",
			PurchasePrice: Decimal{Value: 4000},
			SalePrice:     Decimal{Value: 8000},
			Stock:         150,
			IsActive:      true,
		},
	}

	for _, f := range flowers {
		if err := repo.Create(ctx, f); err != nil {
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
			minCount: 1,
			maxCount: 1,
			wantErr:  false,
		},
		{
			name: "list with pagination",
			filter: FlowerFilter{
				Page:     1,
				PageSize: 2,
			},
			minCount: 2,
			maxCount: 2,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.List(ctx, tt.filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) < tt.minCount || len(result) > tt.maxCount {
					t.Errorf("List() count = %d, want between %d and %d", len(result), tt.minCount, tt.maxCount)
				}
			}
		})
	}
}

// TestFlowerRepository_Update 测试 Update 方法
func TestFlowerRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewFlowerRepository(db)
	ctx := context.Background()

	// 创建测试鲜花
	flower := &Flower{
		SKU:           "UPD001",
		Name:          "红玫瑰",
		Origin:        "云南",
		ShelfLife:     "7天",
		Preservation:  "常温",
		PurchasePrice: Decimal{Value: 5000},
		SalePrice:     Decimal{Value: 10000},
		Stock:         100,
		IsActive:      true,
	}
	if err := repo.Create(ctx, flower); err != nil {
		t.Fatalf("failed to create flower: %v", err)
	}

	// 修改鲜花信息
	flower.Name = "红玫瑰（特大）"
	flower.SalePrice = Decimal{Value: 12000}
	flower.Stock = 150

	err := repo.Update(ctx, flower)
	if err != nil {
		t.Errorf("Update() error = %v", err)
		return
	}

	// 验证更新
	updated, err := repo.GetBySKU(ctx, "UPD001")
	if err != nil {
		t.Errorf("GetBySKU() after Update error = %v", err)
		return
	}

	if updated.Name != "红玫瑰（特大）" {
		t.Errorf("Update() Name = %q, want %q", updated.Name, "红玫瑰（特大）")
	}
	if updated.SalePrice.Value != 12000 {
		t.Errorf("Update() SalePrice = %d, want %d", updated.SalePrice.Value, 12000)
	}
	if updated.Stock != 150 {
		t.Errorf("Update() Stock = %d, want %d", updated.Stock, 150)
	}
}

// TestFlowerRepository_Delete 测试 Delete 方法
func TestFlowerRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewFlowerRepository(db)
	ctx := context.Background()

	// 创建测试鲜花
	flower := &Flower{
		SKU:           "DEL001",
		Name:          "红玫瑰",
		Origin:        "云南",
		ShelfLife:     "7天",
		Preservation:  "常温",
		PurchasePrice: Decimal{Value: 5000},
		SalePrice:     Decimal{Value: 10000},
		Stock:         100,
		IsActive:      true,
	}
	if err := repo.Create(ctx, flower); err != nil {
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
			err := repo.Delete(ctx, tt.sku)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 验证删除后鲜花不存在
			if !tt.wantErr {
				_, err := repo.GetBySKU(ctx, tt.sku)
				if err == nil {
					t.Error("Delete() flower still exists after deletion")
				}
			}
		})
	}
}

// TestFlowerRepository_UpdateStock 测试 UpdateStock 方法
func TestFlowerRepository_UpdateStock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewFlowerRepository(db)
	ctx := context.Background()

	// 创建测试鲜花
	flower := &Flower{
		SKU:           "STK001",
		Name:          "红玫瑰",
		Origin:        "云南",
		ShelfLife:     "7天",
		Preservation:  "常温",
		PurchasePrice: Decimal{Value: 5000},
		SalePrice:     Decimal{Value: 10000},
		Stock:         100,
		IsActive:      true,
	}
	if err := repo.Create(ctx, flower); err != nil {
		t.Fatalf("failed to create flower: %v", err)
	}

	tests := []struct {
		name    string
		sku     string
		delta   int
		wantStock int
		wantErr bool
	}{
		{
			name:     "increase stock",
			sku:      "STK001",
			delta:    50,
			wantStock: 150,
			wantErr:  false,
		},
		{
			name:     "decrease stock",
			sku:      "STK001",
			delta:    -30,
			wantStock: 120,
			wantErr:  false,
		},
		{
			name:     "update non-existing flower",
			sku:      "NONEXIST",
			delta:    10,
			wantStock: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateStock(ctx, tt.sku, tt.delta)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 验证库存已更新
				f, err := repo.GetBySKU(ctx, tt.sku)
				if err != nil {
					t.Errorf("GetBySKU() after UpdateStock error = %v", err)
					return
				}
				if f.Stock != tt.wantStock {
					t.Errorf("UpdateStock() Stock = %d, want %d", f.Stock, tt.wantStock)
				}
			}
		})
	}
}
