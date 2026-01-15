package order

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
)

// OrderRepository 定义订单数据访问接口
type OrderRepository interface {
	Create(ctx context.Context, order *Order, items []*OrderItem) error
	GetByID(ctx context.Context, id int) (*Order, []*OrderItem, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*Order, []*OrderItem, error)
	List(ctx context.Context, filter OrderFilter) ([]*Order, error)
	UpdateStatus(ctx context.Context, id int, status OrderStatus) error
}

// orderRepository 实现 OrderRepository 接口
type orderRepository struct {
	db *sql.DB
}

// NewOrderRepository 创建 OrderRepository 实例
func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Create 创建订单及订单项（需要事务处理）
func (r *orderRepository) Create(ctx context.Context, order *Order, items []*OrderItem) error {
	// 开启事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 插入订单
	orderQuery := `
		INSERT INTO orders (order_no, user_id, address_id, total_amount, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.ExecContext(ctx, orderQuery,
		order.OrderNo, order.UserID, order.AddressID, order.TotalAmount.Value,
		string(order.Status), order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	order.ID = int(id)

	// 插入订单项
	for _, item := range items {
		item.OrderID = order.ID
		itemQuery := `
			INSERT INTO order_items (order_id, flower_sku, flower_name, quantity, unit_price, subtotal)
			VALUES (?, ?, ?, ?, ?, ?)
		`

		_, err := tx.ExecContext(ctx, itemQuery,
			item.OrderID, item.FlowerSKU, item.FlowerName, item.Quantity, item.UnitPrice.Value, item.Subtotal.Value,
		)
		if err != nil {
			return fmt.Errorf("create order item: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetByID 根据 ID 获取订单（包含订单项）
func (r *orderRepository) GetByID(ctx context.Context, id int) (*Order, []*OrderItem, error) {
	// 获取订单
	orderQuery := `
		SELECT id, order_no, user_id, address_id, total_amount, status, created_at, updated_at
		FROM orders WHERE id = ?
	`

	var order Order
	var totalAmount int64
	var status string

	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&order.ID, &order.OrderNo, &order.UserID, &order.AddressID, &totalAmount, &status,
		&order.CreatedAt, &order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("order not found: %d", id)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("get order by id: %w", err)
	}

	order.TotalAmount = flower.Decimal{Value: totalAmount}
	order.Status = OrderStatus(status)

	// 获取订单项
	itemsQuery := `
		SELECT id, order_id, flower_sku, flower_name, quantity, unit_price, subtotal
		FROM order_items WHERE order_id = ?
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	var items []*OrderItem
	for rows.Next() {
		var item OrderItem
		var unitPrice, subtotal int64

		err := rows.Scan(&item.ID, &item.OrderID, &item.FlowerSKU, &item.FlowerName,
			&item.Quantity, &unitPrice, &subtotal)
		if err != nil {
			return nil, nil, fmt.Errorf("scan order item: %w", err)
		}

		item.UnitPrice = flower.Decimal{Value: unitPrice}
		item.Subtotal = flower.Decimal{Value: subtotal}

		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate order items: %w", err)
	}

	return &order, items, nil
}

// GetByOrderNo 根据订单号获取订单（包含订单项）
func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*Order, []*OrderItem, error) {
	// 获取订单
	orderQuery := `
		SELECT id, order_no, user_id, address_id, total_amount, status, created_at, updated_at
		FROM orders WHERE order_no = ?
	`

	var order Order
	var totalAmount int64
	var status string

	err := r.db.QueryRowContext(ctx, orderQuery, orderNo).Scan(
		&order.ID, &order.OrderNo, &order.UserID, &order.AddressID, &totalAmount, &status,
		&order.CreatedAt, &order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("order not found: %s", orderNo)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("get order by order no: %w", err)
	}

	order.TotalAmount = flower.Decimal{Value: totalAmount}
	order.Status = OrderStatus(status)

	// 获取订单项
	itemsQuery := `
		SELECT id, order_id, flower_sku, flower_name, quantity, unit_price, subtotal
		FROM order_items WHERE order_id = ?
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, order.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	var items []*OrderItem
	for rows.Next() {
		var item OrderItem
		var unitPrice, subtotal int64

		err := rows.Scan(&item.ID, &item.OrderID, &item.FlowerSKU, &item.FlowerName,
			&item.Quantity, &unitPrice, &subtotal)
		if err != nil {
			return nil, nil, fmt.Errorf("scan order item: %w", err)
		}

		item.UnitPrice = flower.Decimal{Value: unitPrice}
		item.Subtotal = flower.Decimal{Value: subtotal}

		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate order items: %w", err)
	}

	return &order, items, nil
}

// List 根据筛选条件获取订单列表
func (r *orderRepository) List(ctx context.Context, filter OrderFilter) ([]*Order, error) {
	query := `
		SELECT id, order_no, user_id, address_id, total_amount, status, created_at, updated_at
		FROM orders WHERE 1=1
	`
	args := []interface{}{}

	// 用户筛选
	if filter.UserID > 0 {
		query += " AND user_id = ?"
		args = append(args, filter.UserID)
	}

	// 状态筛选
	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	// 订单号筛选
	if filter.OrderNo != "" {
		query += " AND order_no LIKE ?"
		args = append(args, "%"+filter.OrderNo+"%")
	}

	// 排序
	query += " ORDER BY created_at DESC"

	// 分页
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += " LIMIT ? OFFSET ?"
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		var totalAmount int64
		var status string

		err := rows.Scan(&order.ID, &order.OrderNo, &order.UserID, &order.AddressID, &totalAmount, &status,
			&order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}

		order.TotalAmount = flower.Decimal{Value: totalAmount}
		order.Status = OrderStatus(status)

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil
}

// UpdateStatus 更新订单状态
func (r *orderRepository) UpdateStatus(ctx context.Context, id int, status OrderStatus) error {
	query := `UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, string(status), time.Now(), id)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("order not found: %d", id)
	}

	return nil
}
