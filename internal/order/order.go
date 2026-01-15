package order

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
)

// OrderStatus 订单状态类型
type OrderStatus string

// 订单状态常量
const (
	StatusPending   OrderStatus = "pending"   // 待处理
	StatusCompleted OrderStatus = "completed" // 已完成
	StatusCancelled OrderStatus = "cancelled" // 已取消
)

// Validate 验证订单状态是否有效
func (s OrderStatus) Validate() error {
	switch s {
	case StatusPending, StatusCompleted, StatusCancelled:
		return nil
	default:
		return fmt.Errorf("无效的订单状态: %s", s)
	}
}

// Order 订单实体
type Order struct {
	ID          int                  `json:"id"`
	OrderNo     string               `json:"order_no"`
	UserID      int                  `json:"user_id"`
	AddressID   int                  `json:"address_id"`
	TotalAmount flower.Decimal       `json:"total_amount"`
	Status      OrderStatus          `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Items       []*OrderItem         `json:"items,omitempty"` // 订单项（可选）
}

// OrderItem 订单项实体
type OrderItem struct {
	ID          int            `json:"id"`
	OrderID     int            `json:"order_id"`
	FlowerSKU   string         `json:"flower_sku"`
	FlowerName  string         `json:"flower_name"`
	Quantity    int            `json:"quantity"`
	UnitPrice   flower.Decimal `json:"unit_price"`
	Subtotal    flower.Decimal `json:"subtotal"`
}

// OrderFilter 订单筛选条件
type OrderFilter struct {
	UserID  int        // 按用户筛选
	Status  string     // 按状态筛选
	OrderNo string     // 按订单号筛选
	Page    int
	PageSize int
}

// NewOrder 创建新订单
func NewOrder(userID, addressID int) *Order {
	now := time.Now()
	return &Order{
		OrderNo:     GenerateOrderNo(),
		UserID:      userID,
		AddressID:   addressID,
		TotalAmount: flower.Decimal{Value: 0},
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewOrderItem 创建新订单项
func NewOrderItem(orderID int, flowerSKU, flowerName string, quantity int, unitPrice int64) *OrderItem {
	return &OrderItem{
		OrderID:    orderID,
		FlowerSKU:  flowerSKU,
		FlowerName: flowerName,
		Quantity:   quantity,
		UnitPrice:  flower.Decimal{Value: unitPrice},
		Subtotal:   flower.Decimal{Value: unitPrice * int64(quantity)},
	}
}

// Validate 验证订单数据
func (o *Order) Validate() error {
	if o.OrderNo == "" {
		return fmt.Errorf("订单编号不能为空")
	}
	if o.UserID <= 0 {
		return fmt.Errorf("用户ID不能为空")
	}
	if o.AddressID <= 0 {
		return fmt.Errorf("地址ID不能为空")
	}
	if o.TotalAmount.Value < 0 {
		return fmt.Errorf("订单金额不能为负数")
	}
	if o.Status == "" {
		return fmt.Errorf("订单状态不能为空")
	}
	if err := o.Status.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate 验证订单项数据
func (i *OrderItem) Validate() error {
	if i.FlowerSKU == "" {
		return fmt.Errorf("鲜花SKU不能为空")
	}
	if i.FlowerName == "" {
		return fmt.Errorf("鲜花名称不能为空")
	}
	if i.Quantity <= 0 {
		return fmt.Errorf("数量必须大于0")
	}
	if i.UnitPrice.Value < 0 {
		return fmt.Errorf("单价不能为负数")
	}
	// 验证小计是否正确
	expectedSubtotal := i.UnitPrice.Value * int64(i.Quantity)
	if i.Subtotal.Value != expectedSubtotal {
		return fmt.Errorf("小计金额不正确: expected %d, got %d", expectedSubtotal, i.Subtotal.Value)
	}
	return nil
}

// GenerateOrderNo 生成订单编号
// 格式: ORD + YYYYMMDD + 6位随机数
func GenerateOrderNo() string {
	date := time.Now().Format("20060102")
	random := rand.Intn(999999)
	return fmt.Sprintf("ORD%s%06d", date, random)
}
