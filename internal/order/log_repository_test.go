package order

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
)

// setupTestDB 创建测试数据库连接
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// 创建订单日志表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS order_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER NOT NULL,
		operator_id INTEGER NOT NULL,
		action TEXT NOT NULL,
		old_status TEXT,
		new_status TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
		FOREIGN KEY (operator_id) REFERENCES users(id)
	);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		t.Fatalf("failed to create order_logs table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestOrderLogRepository_CreateLog 测试创建日志
func TestOrderLogRepository_CreateLog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewOrderLogRepository(db)

	tests := []struct {
		name    string
		log     *OrderLog
		wantErr bool
	}{
		{
			name: "create valid log without old status",
			log: &OrderLog{
				OrderID:    1,
				OperatorID: 2,
				Action:     "create_order",
				NewStatus:  StatusPending,
			},
			wantErr: false,
		},
		{
			name: "create valid log with old status",
			log: &OrderLog{
				OrderID:    1,
				OperatorID: 2,
				Action:     "complete_order",
				OldStatus:  StatusPending,
				NewStatus:  StatusCompleted,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := repo.CreateLog(ctx, tt.log)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.log.ID <= 0 {
				t.Errorf("CreateLog() ID = %d, want > 0", tt.log.ID)
			}
		})
	}
}

// TestOrderLogRepository_GetLogs 测试获取日志
func TestOrderLogRepository_GetLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewOrderLogRepository(db)
	ctx := context.Background()

	// 创建多条日志
	logs := []*OrderLog{
		{
			OrderID:    1,
			OperatorID: 2,
			Action:     "create_order",
			NewStatus:  StatusPending,
		},
		{
			OrderID:    1,
			OperatorID: 3,
			Action:     "complete_order",
			OldStatus:  StatusPending,
			NewStatus:  StatusCompleted,
		},
	}

	for _, log := range logs {
		if err := repo.CreateLog(ctx, log); err != nil {
			t.Fatalf("failed to create log: %v", err)
		}
	}

	tests := []struct {
		name     string
		orderID  int
		minCount int
		maxCount int
		wantErr  bool
	}{
		{
			name:     "get logs for order 1",
			orderID:  1,
			minCount: 2,
			maxCount: 2,
			wantErr:  false,
		},
		{
			name:     "get logs for non-existing order",
			orderID:  999,
			minCount: 0,
			maxCount: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetLogs(ctx, tt.orderID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) < tt.minCount || len(result) > tt.maxCount {
					t.Errorf("GetLogs() count = %d, want between %d and %d", len(result), tt.minCount, tt.maxCount)
				}
			}
		})
	}
}
