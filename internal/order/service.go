package order

import (
	"context"
	"fmt"

	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
)

// OrderService 定义订单业务逻辑接口
type OrderService interface {
	CreateOrder(ctx context.Context, userID int, req *CreateOrderRequest) (string, error)
	GetOrder(ctx context.Context, userID int, orderNo string) (*OrderResponse, error)
	ListOrders(ctx context.Context, userID int, filter OrderListFilter) ([]*OrderResponse, error)
	CompleteOrder(ctx context.Context, orderID int, operatorID int) error
	CancelOrder(ctx context.Context, orderID int, operatorID int) error
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	AddressID int                       `json:"address_id"`
	Items     []*CreateOrderItemRequest `json:"items"`
}

// CreateOrderItemRequest 创建订单项请求
type CreateOrderItemRequest struct {
	FlowerSKU string `json:"flower_sku"`
	Quantity  int    `json:"quantity"`
}

// OrderResponse 订单响应
type OrderResponse struct {
	ID          int                `json:"id"`
	OrderNo     string             `json:"order_no"`
	UserID      int                `json:"user_id"`
	AddressID   int                `json:"address_id"`
	TotalAmount int64              `json:"total_amount"` // 以分为单位
	Status      string             `json:"status"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	Items       []*OrderItemResponse `json:"items,omitempty"`
}

// OrderItemResponse 订单项响应
type OrderItemResponse struct {
	ID         int    `json:"id"`
	FlowerSKU  string `json:"flower_sku"`
	FlowerName string `json:"flower_name"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"`  // 以分为单位
	Subtotal   int64  `json:"subtotal"`    // 以分为单位
}

// OrderListFilter 订单列表筛选条件
type OrderListFilter struct {
	Status   string
	OrderNo  string
	Page     int
	PageSize int
}

// orderService 实现 OrderService 接口
type orderService struct {
	orderRepo OrderRepository
	flowerRepo flower.FlowerRepository
	logRepo   OrderLogRepository
}

// NewOrderService 创建 OrderService 实例
func NewOrderService(orderRepo OrderRepository, flowerRepo flower.FlowerRepository, logRepo OrderLogRepository) OrderService {
	return &orderService{
		orderRepo:  orderRepo,
		flowerRepo: flowerRepo,
		logRepo:    logRepo,
	}
}

// CreateOrder 创建订单（含库存扣减事务处理）
func (s *orderService) CreateOrder(ctx context.Context, userID int, req *CreateOrderRequest) (string, error) {
	// 验证请求
	if err := s.validateCreateRequest(req); err != nil {
		return "", err
	}

	// 验证所有鲜花并计算总金额
	orderItems, totalAmount, err := s.validateAndPrepareItems(ctx, req.Items)
	if err != nil {
		return "", err
	}

	// 创建订单实体
	order := NewOrder(userID, req.AddressID)
	order.TotalAmount = flower.Decimal{Value: totalAmount}

	// 执行事务：创建订单 + 扣减库存
	err = s.executeCreateOrderTransaction(ctx, order, orderItems)
	if err != nil {
		return "", err
	}

	// 记录订单日志
	log := NewOrderLog(order.ID, userID, "create_order", StatusPending, "")
	if err := s.logRepo.CreateLog(ctx, log); err != nil {
		// 日志记录失败不影响订单创建
		fmt.Printf("warning: failed to create order log: %v\n", err)
	}

	return order.OrderNo, nil
}

// validateCreateRequest 验证创建订单请求
func (s *orderService) validateCreateRequest(req *CreateOrderRequest) error {
	if req.AddressID <= 0 {
		return fmt.Errorf("地址ID不能为空")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("订单项不能为空")
	}
	return nil
}

// validateAndPrepareItems 验证并准备订单项
func (s *orderService) validateAndPrepareItems(ctx context.Context, items []*CreateOrderItemRequest) ([]*OrderItem, int64, error) {
	orderItems := make([]*OrderItem, 0, len(items))
	var totalAmount int64

	for _, item := range items {
		// 验证数量
		if item.Quantity <= 0 {
			return nil, 0, fmt.Errorf("数量必须大于0")
		}

		// 获取鲜花信息
		flw, err := s.flowerRepo.GetBySKU(ctx, item.FlowerSKU)
		if err != nil {
			return nil, 0, fmt.Errorf("获取鲜花信息失败: %w", err)
		}

		// 验证鲜花是否上架
		if !flw.IsActive {
			return nil, 0, fmt.Errorf("鲜花 %s 已下架", flw.Name)
		}

		// 验证库存
		if flw.Stock < item.Quantity {
			return nil, 0, fmt.Errorf("库存不足: %s (库存: %d, 需要: %d)", flw.Name, flw.Stock, item.Quantity)
		}

		// 创建订单项
		orderItem := NewOrderItem(0, item.FlowerSKU, flw.Name, item.Quantity, flw.SalePrice.Value)
		orderItems = append(orderItems, orderItem)
		totalAmount += orderItem.Subtotal.Value
	}

	return orderItems, totalAmount, nil
}

// executeCreateOrderTransaction 执行创建订单事务
func (s *orderService) executeCreateOrderTransaction(ctx context.Context, order *Order, items []*OrderItem) error {
	// 注意：这里需要使用数据库事务，但由于 Repository 接口不支持事务传递，
	// 我们需要实现一个两阶段提交的方式，或者修改接口支持事务。

	// 为了简单起见，我们先扣减库存，再创建订单。
	// 如果创建订单失败，则需要回滚库存。

	// 1. 先扣减所有库存
	for _, item := range items {
		err := s.flowerRepo.UpdateStock(ctx, item.FlowerSKU, -item.Quantity)
		if err != nil {
			// 库存扣减失败，需要回滚已扣减的库存
			s.rollbackStock(ctx, items, item)
			return fmt.Errorf("扣减库存失败: %w", err)
		}
	}

	// 2. 创建订单
	err := s.orderRepo.Create(ctx, order, items)
	if err != nil {
		// 订单创建失败，回滚库存
		s.rollbackStock(ctx, items, nil)
		return fmt.Errorf("创建订单失败: %w", err)
	}

	return nil
}

// rollbackStock 回滚库存
func (s *orderService) rollbackStock(ctx context.Context, items []*OrderItem, failedItem *OrderItem) {
	for _, item := range items {
		if failedItem != nil && item.FlowerSKU == failedItem.FlowerSKU {
			break // 失败的项不需要回滚
		}
		if err := s.flowerRepo.UpdateStock(ctx, item.FlowerSKU, item.Quantity); err != nil {
			fmt.Printf("warning: failed to rollback stock for %s: %v\n", item.FlowerSKU, err)
		}
	}
}

// GetOrder 获取订单详情（验证用户权限）
func (s *orderService) GetOrder(ctx context.Context, userID int, orderNo string) (*OrderResponse, error) {
	order, items, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}

	// 验证用户只能访问自己的订单
	if order.UserID != userID {
		return nil, fmt.Errorf("无权访问该订单")
	}

	return s.toResponse(order, items), nil
}

// ListOrders 获取订单列表（验证用户权限）
func (s *orderService) ListOrders(ctx context.Context, userID int, filter OrderListFilter) ([]*OrderResponse, error) {
	// 构建筛选条件（强制只能查看自己的订单）
	orderFilter := OrderFilter{
		UserID:   userID,
		Status:   filter.Status,
		OrderNo:  filter.OrderNo,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}

	orders, err := s.orderRepo.List(ctx, orderFilter)
	if err != nil {
		return nil, err
	}

	responses := make([]*OrderResponse, len(orders))
	for i, o := range orders {
		// 获取订单项
		_, orderItems, err := s.orderRepo.GetByID(ctx, o.ID)
		if err != nil {
			// 获取订单项失败，跳过
			responses[i] = s.toResponse(o, nil)
			continue
		}
		responses[i] = s.toResponse(o, orderItems)
	}

	return responses, nil
}

// toResponse 将 Order 实体转换为响应 DTO
func (s *orderService) toResponse(order *Order, items []*OrderItem) *OrderResponse {
	response := &OrderResponse{
		ID:          order.ID,
		OrderNo:     order.OrderNo,
		UserID:      order.UserID,
		AddressID:   order.AddressID,
		TotalAmount: order.TotalAmount.Value,
		Status:      string(order.Status),
		CreatedAt:   order.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   order.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if items != nil {
		response.Items = make([]*OrderItemResponse, len(items))
		for i, item := range items {
			response.Items[i] = &OrderItemResponse{
				ID:         item.ID,
				FlowerSKU:  item.FlowerSKU,
				FlowerName: item.FlowerName,
				Quantity:   item.Quantity,
				UnitPrice:  item.UnitPrice.Value,
				Subtotal:   item.Subtotal.Value,
			}
		}
	}

	return response
}

// CompleteOrder 完成订单
func (s *orderService) CompleteOrder(ctx context.Context, orderID int, operatorID int) error {
	// 获取订单
	order, _, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}

	// 验证订单状态流转：只有待处理订单可以完成
	if order.Status != StatusPending {
		return fmt.Errorf("订单状态不正确，当前状态: %s, 只有待处理订单可以完成", order.Status)
	}

	// 更新订单状态为已完成
	if err := s.orderRepo.UpdateStatus(ctx, orderID, StatusCompleted); err != nil {
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	// 记录订单日志
	log := NewOrderLog(orderID, operatorID, "complete_order", StatusCompleted, order.Status)
	if err := s.logRepo.CreateLog(ctx, log); err != nil {
		// 日志记录失败不影响业务操作
		fmt.Printf("warning: failed to create order log: %v\n", err)
	}

	return nil
}

// CancelOrder 取消订单（含库存回退）
func (s *orderService) CancelOrder(ctx context.Context, orderID int, operatorID int) error {
	// 获取订单
	order, items, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}

	// 验证订单状态流转：只有待处理订单可以取消
	if order.Status != StatusPending {
		return fmt.Errorf("订单状态不正确，当前状态: %s, 只有待处理订单可以取消", order.Status)
	}

	// 回退库存
	for _, item := range items {
		if err := s.flowerRepo.UpdateStock(ctx, item.FlowerSKU, item.Quantity); err != nil {
			// 库存回退失败，记录警告但继续处理
			fmt.Printf("warning: failed to rollback stock for %s: %v\n", item.FlowerSKU, err)
		}
	}

	// 更新订单状态为已取消
	if err := s.orderRepo.UpdateStatus(ctx, orderID, StatusCancelled); err != nil {
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	// 记录订单日志
	log := NewOrderLog(orderID, operatorID, "cancel_order", StatusCancelled, order.Status)
	if err := s.logRepo.CreateLog(ctx, log); err != nil {
		// 日志记录失败不影响业务操作
		fmt.Printf("warning: failed to create order log: %v\n", err)
	}

	return nil
}
