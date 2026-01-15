package flower

import (
	"context"
	"fmt"
)

// FlowerService 定义鲜花业务逻辑接口
type FlowerService interface {
	CreateFlower(ctx context.Context, req *CreateFlowerRequest) error
	GetFlower(ctx context.Context, sku string) (*FlowerResponse, error)
	ListFlowers(ctx context.Context, filter FlowerFilter) ([]*FlowerResponse, error)
	UpdateFlower(ctx context.Context, sku string, req *UpdateFlowerRequest) error
	DeleteFlower(ctx context.Context, sku string) error
	AddStock(ctx context.Context, sku string, quantity int) error
}

// CreateFlowerRequest 创建鲜花请求
type CreateFlowerRequest struct {
	SKU           string
	Name          string
	Origin        string
	ShelfLife     string
	Preservation  string
	PurchasePrice float64
	SalePrice     float64
	Stock         int
}

// UpdateFlowerRequest 更新鲜花请求
type UpdateFlowerRequest struct {
	Name          *string
	Origin        *string
	ShelfLife     *string
	Preservation  *string
	PurchasePrice *float64
	SalePrice     *float64
}

// FlowerResponse 鲜花响应（带库存预警标识）
type FlowerResponse struct {
	SKU           string  `json:"sku"`
	Name          string  `json:"name"`
	Origin        string  `json:"origin"`
	ShelfLife     string  `json:"shelf_life"`
	Preservation  string  `json:"preservation"`
	PurchasePrice float64 `json:"purchase_price"`
	SalePrice     float64 `json:"sale_price"`
	Stock         int     `json:"stock"`
	IsActive      bool    `json:"is_active"`
	LowStock      bool    `json:"low_stock"` // 库存预警标识
}

// flowerService 实现 FlowerService 接口
type flowerService struct {
	repo      FlowerRepository
	threshold int // 库存预警阈值
}

// NewFlowerService 创建 FlowerService 实例
func NewFlowerService(repo FlowerRepository) FlowerService {
	return &flowerService{
		repo:      repo,
		threshold: 10, // 默认库存预警阈值为 10
	}
}

// CreateFlower 创建鲜花
func (s *flowerService) CreateFlower(ctx context.Context, req *CreateFlowerRequest) error {
	// 创建 Flower 实体
	flower := &Flower{
		SKU:           req.SKU,
		Name:          req.Name,
		Origin:        req.Origin,
		ShelfLife:     req.ShelfLife,
		Preservation:  req.Preservation,
		PurchasePrice: DecimalFromFloat64(req.PurchasePrice),
		SalePrice:     DecimalFromFloat64(req.SalePrice),
		Stock:         req.Stock,
		IsActive:      true,
	}

	// 验证数据
	if err := flower.Validate(); err != nil {
		return err
	}

	// 保存到数据库
	return s.repo.Create(ctx, flower)
}

// GetFlower 获取鲜花详情
func (s *flowerService) GetFlower(ctx context.Context, sku string) (*FlowerResponse, error) {
	flower, err := s.repo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}

	return s.toResponse(flower), nil
}

// ListFlowers 获取鲜花列表
func (s *flowerService) ListFlowers(ctx context.Context, filter FlowerFilter) ([]*FlowerResponse, error) {
	// 验证筛选条件
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	flowers, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]*FlowerResponse, len(flowers))
	for i, f := range flowers {
		responses[i] = s.toResponse(f)
	}

	return responses, nil
}

// UpdateFlower 更新鲜花信息
func (s *flowerService) UpdateFlower(ctx context.Context, sku string, req *UpdateFlowerRequest) error {
	// 获取现有鲜花
	flower, err := s.repo.GetBySKU(ctx, sku)
	if err != nil {
		return err
	}

	// 更新字段
	if req.Name != nil {
		flower.Name = *req.Name
	}
	if req.Origin != nil {
		flower.Origin = *req.Origin
	}
	if req.ShelfLife != nil {
		flower.ShelfLife = *req.ShelfLife
	}
	if req.Preservation != nil {
		flower.Preservation = *req.Preservation
	}
	if req.PurchasePrice != nil {
		flower.PurchasePrice = DecimalFromFloat64(*req.PurchasePrice)
	}
	if req.SalePrice != nil {
		flower.SalePrice = DecimalFromFloat64(*req.SalePrice)
	}

	// 验证更新后的数据
	if err := flower.Validate(); err != nil {
		return err
	}

	// 保存到数据库
	return s.repo.Update(ctx, flower)
}

// DeleteFlower 删除鲜花
func (s *flowerService) DeleteFlower(ctx context.Context, sku string) error {
	return s.repo.Delete(ctx, sku)
}

// AddStock 进货入库
func (s *flowerService) AddStock(ctx context.Context, sku string, quantity int) error {
	// 验证数量
	if quantity < 0 {
		return fmt.Errorf("进货数量不能为负数")
	}

	// 更新库存
	return s.repo.UpdateStock(ctx, sku, quantity)
}

// toResponse 将 Flower 实体转换为响应 DTO
func (s *flowerService) toResponse(f *Flower) *FlowerResponse {
	return &FlowerResponse{
		SKU:           f.SKU,
		Name:          f.Name,
		Origin:        f.Origin,
		ShelfLife:     f.ShelfLife,
		Preservation:  f.Preservation,
		PurchasePrice: f.PurchasePrice.ToFloat64(),
		SalePrice:     f.SalePrice.ToFloat64(),
		Stock:         f.Stock,
		IsActive:      f.IsActive,
		LowStock:      f.IsLowStock(s.threshold),
	}
}
