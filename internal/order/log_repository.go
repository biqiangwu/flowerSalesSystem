package order

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// OrderLogRepository 定义订单日志数据访问接口
type OrderLogRepository interface {
	CreateLog(ctx context.Context, log *OrderLog) error
	GetLogs(ctx context.Context, orderID int) ([]*OrderLog, error)
}

// orderLogRepository 实现 OrderLogRepository 接口
type orderLogRepository struct {
	db *sql.DB
}

// NewOrderLogRepository 创建 OrderLogRepository 实例
func NewOrderLogRepository(db *sql.DB) OrderLogRepository {
	return &orderLogRepository{db: db}
}

// CreateLog 创建订单日志
func (r *orderLogRepository) CreateLog(ctx context.Context, log *OrderLog) error {
	log.CreatedAt = time.Now()

	query := `
		INSERT INTO order_logs (order_id, operator_id, action, old_status, new_status, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		log.OrderID, log.OperatorID, log.Action, string(log.OldStatus), string(log.NewStatus), log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create order log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	log.ID = int(id)
	return nil
}

// GetLogs 获取订单的所有日志
func (r *orderLogRepository) GetLogs(ctx context.Context, orderID int) ([]*OrderLog, error) {
	query := `
		SELECT id, order_id, operator_id, action, old_status, new_status, created_at
		FROM order_logs WHERE order_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order logs: %w", err)
	}
	defer rows.Close()

	var logs []*OrderLog
	for rows.Next() {
		var log OrderLog
		var oldStatus, newStatus string

		err := rows.Scan(&log.ID, &log.OrderID, &log.OperatorID, &log.Action, &oldStatus, &newStatus, &log.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan order log: %w", err)
		}

		log.OldStatus = OrderStatus(oldStatus)
		log.NewStatus = OrderStatus(newStatus)

		logs = append(logs, &log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order logs: %w", err)
	}

	return logs, nil
}
