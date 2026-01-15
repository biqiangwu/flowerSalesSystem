package order

import (
	"testing"

	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
)

func TestOrderStatus(t *testing.T) {
	tests := []struct {
		name   string
		status OrderStatus
		valid  bool
	}{
		{
			name:   "valid pending status",
			status: StatusPending,
			valid:  true,
		},
		{
			name:   "valid completed status",
			status: StatusCompleted,
			valid:  true,
		},
		{
			name:   "valid cancelled status",
			status: StatusCancelled,
			valid:  true,
		},
		{
			name:   "invalid status",
			status: OrderStatus("invalid"),
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.status.Validate()
			if tt.valid && err != nil {
				t.Errorf("OrderStatus.Validate() should not return error, got %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("OrderStatus.Validate() should return error for invalid status")
			}
		})
	}
}

func TestGenerateOrderNo(t *testing.T) {
	orderNo := GenerateOrderNo()

	if len(orderNo) != 17 {
		t.Errorf("GenerateOrderNo() length = %d, want 17", len(orderNo))
	}

	if orderNo[:3] != "ORD" {
		t.Errorf("GenerateOrderNo() prefix = %s, want ORD", orderNo[:3])
	}

	// 测试多个订单编号应该不同
	orderNo2 := GenerateOrderNo()
	if orderNo == orderNo2 {
		t.Error("GenerateOrderNo() should generate unique order numbers")
	}
}

func TestOrderValidation(t *testing.T) {
	tests := []struct {
		name    string
		order   *Order
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid order",
			order: &Order{
				OrderNo:     "ORD20260114123456",
				UserID:      1,
				AddressID:   1,
				TotalAmount: flower.Decimal{Value: 10000},
				Status:      StatusPending,
			},
			wantErr: false,
		},
		{
			name: "invalid - empty order no",
			order: &Order{
				OrderNo:     "",
				UserID:      1,
				AddressID:   1,
				TotalAmount: flower.Decimal{Value: 10000},
				Status:      StatusPending,
			},
			wantErr: true,
			errMsg:  "订单编号不能为空",
		},
		{
			name: "invalid - user ID is 0",
			order: &Order{
				OrderNo:     "ORD20260114123456",
				UserID:      0,
				AddressID:   1,
				TotalAmount: flower.Decimal{Value: 10000},
				Status:      StatusPending,
			},
			wantErr: true,
			errMsg:  "用户ID不能为空",
		},
		{
			name: "invalid - address ID is 0",
			order: &Order{
				OrderNo:     "ORD20260114123456",
				UserID:      1,
				AddressID:   0,
				TotalAmount: flower.Decimal{Value: 10000},
				Status:      StatusPending,
			},
			wantErr: true,
			errMsg:  "地址ID不能为空",
		},
		{
			name: "invalid - total amount is negative",
			order: &Order{
				OrderNo:     "ORD20260114123456",
				UserID:      1,
				AddressID:   1,
				TotalAmount: flower.Decimal{Value: -100},
				Status:      StatusPending,
			},
			wantErr: true,
			errMsg:  "订单金额不能为负数",
		},
		{
			name: "invalid - empty status",
			order: &Order{
				OrderNo:     "ORD20260114123456",
				UserID:      1,
				AddressID:   1,
				TotalAmount: flower.Decimal{Value: 10000},
				Status:      "",
			},
			wantErr: true,
			errMsg:  "订单状态不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Order.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Order.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestOrderItemValidation(t *testing.T) {
	tests := []struct {
		name     string
		item     *OrderItem
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid order item",
			item: &OrderItem{
				OrderID:     1,
				FlowerSKU:   "FLW001",
				FlowerName:  "红玫瑰",
				Quantity:    10,
				UnitPrice:   flower.Decimal{Value: 1000},
				Subtotal:    flower.Decimal{Value: 10000},
			},
			wantErr: false,
		},
		{
			name: "invalid - empty flower SKU",
			item: &OrderItem{
				OrderID:     1,
				FlowerSKU:   "",
				FlowerName:  "红玫瑰",
				Quantity:    10,
				UnitPrice:   flower.Decimal{Value: 1000},
				Subtotal:    flower.Decimal{Value: 10000},
			},
			wantErr: true,
			errMsg:  "鲜花SKU不能为空",
		},
		{
			name: "invalid - empty flower name",
			item: &OrderItem{
				OrderID:     1,
				FlowerSKU:   "FLW001",
				FlowerName:  "",
				Quantity:    10,
				UnitPrice:   flower.Decimal{Value: 1000},
				Subtotal:    flower.Decimal{Value: 10000},
			},
			wantErr: true,
			errMsg:  "鲜花名称不能为空",
		},
		{
			name: "invalid - quantity is 0",
			item: &OrderItem{
				OrderID:     1,
				FlowerSKU:   "FLW001",
				FlowerName:  "红玫瑰",
				Quantity:    0,
				UnitPrice:   flower.Decimal{Value: 1000},
				Subtotal:    flower.Decimal{Value: 0},
			},
			wantErr: true,
			errMsg:  "数量必须大于0",
		},
		{
			name: "invalid - negative quantity",
			item: &OrderItem{
				OrderID:     1,
				FlowerSKU:   "FLW001",
				FlowerName:  "红玫瑰",
				Quantity:    -10,
				UnitPrice:   flower.Decimal{Value: 1000},
				Subtotal:    flower.Decimal{Value: -10000},
			},
			wantErr: true,
			errMsg:  "数量必须大于0",
		},
		{
			name: "invalid - unit price is negative",
			item: &OrderItem{
				OrderID:     1,
				FlowerSKU:   "FLW001",
				FlowerName:  "红玫瑰",
				Quantity:    10,
				UnitPrice:   flower.Decimal{Value: -1000},
				Subtotal:    flower.Decimal{Value: -10000},
			},
			wantErr: true,
			errMsg:  "单价不能为负数",
		},
		{
			name: "invalid - subtotal mismatch",
			item: &OrderItem{
				OrderID:     1,
				FlowerSKU:   "FLW001",
				FlowerName:  "红玫瑰",
				Quantity:    10,
				UnitPrice:   flower.Decimal{Value: 1000},
				Subtotal:    flower.Decimal{Value: 5000}, // 应该是 10000
			},
			wantErr: true,
			errMsg:  "小计金额不正确: expected 10000, got 5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderItem.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("OrderItem.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestNewOrder(t *testing.T) {
	order := NewOrder(1, 1)

	if order.OrderNo == "" {
		t.Error("NewOrder() OrderNo should not be empty")
	}
	if order.UserID != 1 {
		t.Errorf("NewOrder() UserID = %d, want 1", order.UserID)
	}
	if order.AddressID != 1 {
		t.Errorf("NewOrder() AddressID = %d, want 1", order.AddressID)
	}
	if order.Status != StatusPending {
		t.Errorf("NewOrder() Status = %s, want %s", order.Status, StatusPending)
	}
	if order.CreatedAt.IsZero() {
		t.Error("NewOrder() CreatedAt should be set")
	}
	if order.UpdatedAt.IsZero() {
		t.Error("NewOrder() UpdatedAt should be set")
	}
}

func TestNewOrderItem(t *testing.T) {
	item := NewOrderItem(1, "FLW001", "红玫瑰", 10, 1000)

	if item.FlowerSKU != "FLW001" {
		t.Errorf("NewOrderItem() FlowerSKU = %s, want FLW001", item.FlowerSKU)
	}
	if item.FlowerName != "红玫瑰" {
		t.Errorf("NewOrderItem() FlowerName = %s, want 红玫瑰", item.FlowerName)
	}
	if item.Quantity != 10 {
		t.Errorf("NewOrderItem() Quantity = %d, want 10", item.Quantity)
	}
	if item.UnitPrice.Value != 1000 {
		t.Errorf("NewOrderItem() UnitPrice = %d, want 1000", item.UnitPrice.Value)
	}
	// 小计应该自动计算：10 * 1000 = 10000
	if item.Subtotal.Value != 10000 {
		t.Errorf("NewOrderItem() Subtotal = %d, want 10000", item.Subtotal.Value)
	}
}
