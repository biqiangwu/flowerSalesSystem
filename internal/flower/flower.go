package flower

import (
	"fmt"
	"time"
)

// Flower 表示鲜花实体
type Flower struct {
	SKU           string    `json:"sku"`
	Name          string    `json:"name"`
	Origin        string    `json:"origin"`
	ShelfLife     string    `json:"shelf_life"`
	Preservation  string    `json:"preservation"`
	PurchasePrice Decimal   `json:"purchase_price"`
	SalePrice     Decimal   `json:"sale_price"`
	Stock         int       `json:"stock"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// FlowerFilter 表示鲜花列表查询的筛选条件
type FlowerFilter struct {
	Search   string  // 按 sku/name 搜索
	Origin   string  // 按产地筛选
	MinPrice float64 // 最低价
	MaxPrice float64 // 最高价
	SortBy   string  // price_asc, price_desc, stock
	Page     int
	PageSize int
}

// NewFlower 创建一个新的鲜花实体
func NewFlower(sku, name, origin, shelfLife, preservation string, purchasePrice, salePrice float64, stock int) *Flower {
	now := time.Now()
	return &Flower{
		SKU:           sku,
		Name:          name,
		Origin:        origin,
		ShelfLife:     shelfLife,
		Preservation:  preservation,
		PurchasePrice: DecimalFromFloat64(purchasePrice),
		SalePrice:     DecimalFromFloat64(salePrice),
		Stock:         stock,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Validate 验证鲜花数据是否有效
func (f *Flower) Validate() error {
	if f.SKU == "" {
		return fmt.Errorf("SKU不能为空")
	}
	if f.Name == "" {
		return fmt.Errorf("名称不能为空")
	}
	if f.Origin == "" {
		return fmt.Errorf("产地不能为空")
	}
	if f.SalePrice.LessThan(f.PurchasePrice) {
		return fmt.Errorf("销售价格不能低于进货价格")
	}
	if f.Stock < 0 {
		return fmt.Errorf("库存不能为负数")
	}
	return nil
}

// IsLowStock 判断库存是否低于或等于阈值
func (f *Flower) IsLowStock(threshold int) bool {
	return f.Stock <= threshold
}

// Validate 验证筛选条件是否有效
func (f *FlowerFilter) Validate() error {
	if f.MinPrice > f.MaxPrice && f.MaxPrice > 0 {
		return fmt.Errorf("最低价格不能高于最高价格")
	}
	// Page 和 PageSize 为 0 时表示不分页（返回所有结果）
	// 如果设置了分页，则必须大于 0
	if f.Page < 0 || f.PageSize < 0 {
		return fmt.Errorf("页码和每页数量不能为负数")
	}
	// 如果其中一个设置了，另一个也必须设置
	if (f.Page > 0 && f.PageSize == 0) || (f.Page == 0 && f.PageSize > 0) {
		return fmt.Errorf("页码和每页数量必须同时设置")
	}
	return nil
}
