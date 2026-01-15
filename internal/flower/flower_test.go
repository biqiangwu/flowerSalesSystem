package flower

import (
	"testing"
)

func TestFlowerValidation(t *testing.T) {
	tests := []struct {
		name    string
		flower  Flower
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效的鲜花",
			flower: Flower{
				SKU:          "FLW001",
				Name:         "红玫瑰",
				Origin:       "云南",
				ShelfLife:    "7天",
				Preservation: "常温",
				PurchasePrice: Decimal{Value: 5000},  // 50.00
				SalePrice:     Decimal{Value: 10000},  // 100.00
				Stock:         100,
				IsActive:      true,
			},
			wantErr: false,
		},
		{
			name: "SKU为空",
			flower: Flower{
				SKU:          "",
				Name:         "红玫瑰",
				Origin:       "云南",
				ShelfLife:    "7天",
				Preservation: "常温",
				PurchasePrice: Decimal{Value: 5000},
				SalePrice:     Decimal{Value: 10000},
				Stock:         100,
				IsActive:      true,
			},
			wantErr: true,
			errMsg:  "SKU不能为空",
		},
		{
			name: "名称为空",
			flower: Flower{
				SKU:          "FLW001",
				Name:         "",
				Origin:       "云南",
				ShelfLife:    "7天",
				Preservation: "常温",
				PurchasePrice: Decimal{Value: 5000},
				SalePrice:     Decimal{Value: 10000},
				Stock:         100,
				IsActive:      true,
			},
			wantErr: true,
			errMsg:  "名称不能为空",
		},
		{
			name: "产地为空",
			flower: Flower{
				SKU:          "FLW001",
				Name:         "红玫瑰",
				Origin:       "",
				ShelfLife:    "7天",
				Preservation: "常温",
				PurchasePrice: Decimal{Value: 5000},
				SalePrice:     Decimal{Value: 10000},
				Stock:         100,
				IsActive:      true,
			},
			wantErr: true,
			errMsg:  "产地不能为空",
		},
		{
			name: "销售价格低于进货价格",
			flower: Flower{
				SKU:          "FLW001",
				Name:         "红玫瑰",
				Origin:       "云南",
				ShelfLife:    "7天",
				Preservation: "常温",
				PurchasePrice: Decimal{Value: 10000},  // 100.00
				SalePrice:     Decimal{Value: 5000},   // 50.00
				Stock:         100,
				IsActive:      true,
			},
			wantErr: true,
			errMsg:  "销售价格不能低于进货价格",
		},
		{
			name: "库存为负数",
			flower: Flower{
				SKU:          "FLW001",
				Name:         "红玫瑰",
				Origin:       "云南",
				ShelfLife:    "7天",
				Preservation: "常温",
				PurchasePrice: Decimal{Value: 5000},
				SalePrice:     Decimal{Value: 10000},
				Stock:         -10,
				IsActive:      true,
			},
			wantErr: true,
			errMsg:  "库存不能为负数",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flower.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Flower.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Flower.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestFlowerIsLowStock(t *testing.T) {
	tests := []struct {
		name           string
		flower         Flower
		threshold      int
		wantLowStock   bool
	}{
		{
			name: "库存充足",
			flower: Flower{
				SKU:   "FLW001",
				Stock: 100,
			},
			threshold:    10,
			wantLowStock: false,
		},
		{
			name: "库存等于阈值",
			flower: Flower{
				SKU:   "FLW001",
				Stock: 10,
			},
			threshold:    10,
			wantLowStock: true,
		},
		{
			name: "库存低于阈值",
			flower: Flower{
				SKU:   "FLW001",
				Stock: 5,
			},
			threshold:    10,
			wantLowStock: true,
		},
		{
			name: "库存为零",
			flower: Flower{
				SKU:   "FLW001",
				Stock: 0,
			},
			threshold:    10,
			wantLowStock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.flower.IsLowStock(tt.threshold)
			if got != tt.wantLowStock {
				t.Errorf("Flower.IsLowStock() = %v, want %v", got, tt.wantLowStock)
			}
		})
	}
}

func TestFlowerFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter FlowerFilter
		valid  bool
	}{
		{
			name: "有效的空筛选条件",
			filter: FlowerFilter{},
			valid: true,
		},
		{
			name: "有效的搜索条件",
			filter: FlowerFilter{
				Search: "玫瑰",
			},
			valid: true,
		},
		{
			name: "有效的产地筛选",
			filter: FlowerFilter{
				Origin: "云南",
			},
			valid: true,
		},
		{
			name: "有效的价格区间筛选",
			filter: FlowerFilter{
				MinPrice: 10,
				MaxPrice: 100,
			},
			valid: true,
		},
		{
			name: "有效的排序条件",
			filter: FlowerFilter{
				SortBy: "price_asc",
			},
			valid: true,
		},
		{
			name: "有效的分页条件",
			filter: FlowerFilter{
				Page:     1,
				PageSize: 20,
			},
			valid: true,
		},
		{
			name: "无效的价格区间-最小大于最大",
			filter: FlowerFilter{
				MinPrice: 100,
				MaxPrice: 10,
			},
			valid: false,
		},
		{
			name: "无效的负页码",
			filter: FlowerFilter{
				Page: -1,
			},
			valid: false,
		},
		{
			name: "无效的负每页数量",
			filter: FlowerFilter{
				Page:     1,
				PageSize: -1,
			},
			valid: false,
		},
		{
			name: "无效的分页-只设置了页码",
			filter: FlowerFilter{
				Page:     1,
				PageSize: 0,
			},
			valid: false,
		},
		{
			name: "无效的分页-只设置了每页数量",
			filter: FlowerFilter{
				Page:     0,
				PageSize: 20,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.valid && err != nil {
				t.Errorf("FlowerFilter.Validate() should not return error, got %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("FlowerFilter.Validate() should return error for invalid filter")
			}
		})
	}
}

func TestNewFlower(t *testing.T) {
	flower := NewFlower("FLW001", "红玫瑰", "云南", "7天", "常温", 50.00, 100.00, 100)

	if flower.SKU != "FLW001" {
		t.Errorf("NewFlower() SKU = %v, want %v", flower.SKU, "FLW001")
	}
	if flower.Name != "红玫瑰" {
		t.Errorf("NewFlower() Name = %v, want %v", flower.Name, "红玫瑰")
	}
	if flower.Origin != "云南" {
		t.Errorf("NewFlower() Origin = %v, want %v", flower.Origin, "云南")
	}
	if flower.PurchasePrice.Value != 5000 {
		t.Errorf("NewFlower() PurchasePrice = %v, want %v", flower.PurchasePrice.Value, 5000)
	}
	if flower.SalePrice.Value != 10000 {
		t.Errorf("NewFlower() SalePrice = %v, want %v", flower.SalePrice.Value, 10000)
	}
	if flower.Stock != 100 {
		t.Errorf("NewFlower() Stock = %v, want %v", flower.Stock, 100)
	}
	if !flower.IsActive {
		t.Errorf("NewFlower() IsActive = %v, want %v", flower.IsActive, true)
	}
	if flower.CreatedAt.IsZero() {
		t.Errorf("NewFlower() CreatedAt should be set")
	}
	if flower.UpdatedAt.IsZero() {
		t.Errorf("NewFlower() UpdatedAt should be set")
	}
}
