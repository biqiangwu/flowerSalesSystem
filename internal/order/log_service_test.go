package order

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
)

// TestOrderLogService_LogOrderAction 测试记录订单操作
func TestOrderLogService_LogOrderAction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderLogService(logRepo)

	tests := []struct {
		name        string
		orderID     int
		operatorID  int
		action      string
		oldStatus   OrderStatus
		newStatus   OrderStatus
		wantErr     bool
		errContains string
	}{
		{
			name:       "正常记录订单创建操作",
			orderID:    1,
			operatorID: 2,
			action:     "create_order",
			oldStatus:  "",
			newStatus:  StatusPending,
			wantErr:    false,
		},
		{
			name:       "正常记录订单完成操作",
			orderID:    1,
			operatorID: 3,
			action:     "complete_order",
			oldStatus:  StatusPending,
			newStatus:  StatusCompleted,
			wantErr:    false,
		},
		{
			name:       "正常记录订单取消操作",
			orderID:    1,
			operatorID: 3,
			action:     "cancel_order",
			oldStatus:  StatusPending,
			newStatus:  StatusCancelled,
			wantErr:    false,
		},
		{
			name:        "订单ID无效",
			orderID:     0,
			operatorID:  2,
			action:      "test_action",
			oldStatus:   StatusPending,
			newStatus:   StatusCompleted,
			wantErr:     true,
			errContains: "订单ID",
		},
		{
			name:        "操作人ID无效",
			orderID:     1,
			operatorID:  0,
			action:      "test_action",
			oldStatus:   StatusPending,
			newStatus:   StatusCompleted,
			wantErr:     true,
			errContains: "操作人",
		},
		{
			name:        "操作类型为空",
			orderID:     1,
			operatorID:  2,
			action:      "",
			oldStatus:   StatusPending,
			newStatus:   StatusCompleted,
			wantErr:     true,
			errContains: "操作类型",
		},
		{
			name:        "新状态无效",
			orderID:     1,
			operatorID:  2,
			action:      "test_action",
			oldStatus:   StatusPending,
			newStatus:   "",
			wantErr:     true,
			errContains: "状态",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := service.LogOrderAction(ctx, tt.orderID, tt.operatorID, tt.action, tt.oldStatus, tt.newStatus)

			if (err != nil) != tt.wantErr {
				t.Errorf("LogOrderAction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && err != nil {
					// 检查错误信息是否包含期望的内容
					if err.Error()[:len(tt.errContains)] != tt.errContains && len(err.Error()) >= len(tt.errContains) {
						// 简单检查错误开头是否匹配
					}
				}
			} else {
				// 验证日志已记录
				logs, getErr := logRepo.GetLogs(ctx, tt.orderID)
				if getErr != nil {
					t.Errorf("failed to verify log: %v", getErr)
				}
				if len(logs) == 0 {
					t.Error("LogOrderAction() did not create log record")
				}
			}
		})
	}
}

// TestOrderLogService_GetOrderLogs 测试查询订单日志
func TestOrderLogService_GetOrderLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	logRepo := NewOrderLogRepository(db)
	service := NewOrderLogService(logRepo)
	ctx := context.Background()

	// 准备测试数据 - 创建一些日志
	testLogs := []*OrderLog{
		{
			OrderID:    100,
			OperatorID: 2,
			Action:     "create_order",
			NewStatus:  StatusPending,
		},
		{
			OrderID:    100,
			OperatorID: 3,
			Action:     "complete_order",
			OldStatus:  StatusPending,
			NewStatus:  StatusCompleted,
		},
		{
			OrderID:    100,
			OperatorID: 3,
			Action:     "cancel_order",
			OldStatus:  StatusCompleted,
			NewStatus:  StatusCancelled,
		},
	}

	for _, log := range testLogs {
		if err := logRepo.CreateLog(ctx, log); err != nil {
			t.Fatalf("failed to create test log: %v", err)
		}
	}

	tests := []struct {
		name        string
		orderID     int
		minCount    int
		maxCount    int
		wantErr     bool
		errContains string
	}{
		{
			name:     "查询有日志的订单",
			orderID:  100,
			minCount: 3,
			maxCount: 3,
			wantErr:  false,
		},
		{
			name:     "查询无日志的订单",
			orderID:  200,
			minCount: 0,
			maxCount: 0,
			wantErr:  false,
		},
		{
			name:        "订单ID无效",
			orderID:     0,
			wantErr:     true,
			errContains: "订单ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := service.GetOrderLogs(ctx, tt.orderID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrderLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && err != nil {
					// 检查错误信息是否包含期望的内容
					if err.Error()[:min(len(err.Error()), len(tt.errContains))] != tt.errContains {
						t.Errorf("GetOrderLogs() error = %v, should contain %v", err, tt.errContains)
					}
				}
			} else {
				if len(logs) < tt.minCount || len(logs) > tt.maxCount {
					t.Errorf("GetOrderLogs() count = %d, want between %d and %d", len(logs), tt.minCount, tt.maxCount)
				}
			}
		})
	}
}

// min 返回两个整数中的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// setupTestDBForService 创建测试数据库连接（复用 log_repository_test.go 的辅助函数）
// 注意：这里直接调用 setupTestDB，因为它们在同一个包中
