package order

import (
	"fmt"
	"time"
)

// OrderLog 订单操作日志实体
type OrderLog struct {
	ID         int       `json:"id"`
	OrderID    int       `json:"order_id"`
	OperatorID int       `json:"operator_id"`
	Action     string    `json:"action"`     // 操作类型：create_order, complete_order, cancel_order 等
	OldStatus  OrderStatus `json:"old_status"` // 变更前状态
	NewStatus  OrderStatus `json:"new_status"` // 变更后状态
	CreatedAt  time.Time `json:"created_at"`
}

// NewOrderLog 创建订单日志
func NewOrderLog(orderID, operatorID int, action string, newStatus, oldStatus OrderStatus) *OrderLog {
	return &OrderLog{
		OrderID:    orderID,
		OperatorID: operatorID,
		Action:     action,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		CreatedAt:  time.Now(),
	}
}

// Validate 验证订单日志数据
func (l *OrderLog) Validate() error {
	if l.OrderID <= 0 {
		return fmt.Errorf("订单ID不能为空")
	}
	if l.OperatorID <= 0 {
		return fmt.Errorf("操作人ID不能为空")
	}
	if l.Action == "" {
		return fmt.Errorf("操作类型不能为空")
	}
	if l.NewStatus == "" {
		return fmt.Errorf("新状态不能为空")
	}
	if err := l.NewStatus.Validate(); err != nil {
		return err
	}
	// OldStatus 可以为空（创建订单时），但如果设置了需要验证
	if l.OldStatus != "" {
		if err := l.OldStatus.Validate(); err != nil {
			return err
		}
	}
	return nil
}
