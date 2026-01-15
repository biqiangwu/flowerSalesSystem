package order

import (
	"context"
	"fmt"
)

// OrderLogService 定义订单日志服务接口
type OrderLogService interface {
	LogOrderAction(ctx context.Context, orderID, operatorID int, action string, oldStatus, newStatus OrderStatus) error
	GetOrderLogs(ctx context.Context, orderID int) ([]*OrderLog, error)
}

// orderLogService 实现 OrderLogService 接口
type orderLogService struct {
	logRepo OrderLogRepository
}

// NewOrderLogService 创建 OrderLogService 实例
func NewOrderLogService(logRepo OrderLogRepository) OrderLogService {
	return &orderLogService{logRepo: logRepo}
}

// LogOrderAction 记录订单操作日志
func (s *orderLogService) LogOrderAction(ctx context.Context, orderID, operatorID int, action string, oldStatus, newStatus OrderStatus) error {
	// 创建日志对象
	log := NewOrderLog(orderID, operatorID, action, newStatus, oldStatus)

	// 验证日志数据
	if err := log.Validate(); err != nil {
		return err
	}

	// 持久化日志
	if err := s.logRepo.CreateLog(ctx, log); err != nil {
		return fmt.Errorf("记录订单日志: %w", err)
	}

	return nil
}

// GetOrderLogs 获取订单的所有操作日志
func (s *orderLogService) GetOrderLogs(ctx context.Context, orderID int) ([]*OrderLog, error) {
	// 验证订单ID
	if orderID <= 0 {
		return nil, fmt.Errorf("订单ID无效")
	}

	// 查询日志
	logs, err := s.logRepo.GetLogs(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("查询订单日志: %w", err)
	}

	return logs, nil
}
