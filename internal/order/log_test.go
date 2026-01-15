package order

import (
	"testing"
)

func TestOrderLogValidation(t *testing.T) {
	tests := []struct {
		name    string
		log     *OrderLog
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid order log",
			log: &OrderLog{
				OrderID:    1,
				OperatorID: 2,
				Action:     "create_order",
				NewStatus:  StatusPending,
			},
			wantErr: false,
		},
		{
			name: "valid order log with old status",
			log: &OrderLog{
				OrderID:    1,
				OperatorID: 2,
				Action:     "complete_order",
				OldStatus:  StatusPending,
				NewStatus:  StatusCompleted,
			},
			wantErr: false,
		},
		{
			name: "invalid - order ID is 0",
			log: &OrderLog{
				OrderID:    0,
				OperatorID: 2,
				Action:     "create_order",
				NewStatus:  StatusPending,
			},
			wantErr: true,
			errMsg:  "订单ID不能为空",
		},
		{
			name: "invalid - operator ID is 0",
			log: &OrderLog{
				OrderID:    1,
				OperatorID: 0,
				Action:     "create_order",
				NewStatus:  StatusPending,
			},
			wantErr: true,
			errMsg:  "操作人ID不能为空",
		},
		{
			name: "invalid - empty action",
			log: &OrderLog{
				OrderID:    1,
				OperatorID: 2,
				Action:     "",
				NewStatus:  StatusPending,
			},
			wantErr: true,
			errMsg:  "操作类型不能为空",
		},
		{
			name: "invalid - empty new status",
			log: &OrderLog{
				OrderID:    1,
				OperatorID: 2,
				Action:     "create_order",
				NewStatus:  "",
			},
			wantErr: true,
			errMsg:  "新状态不能为空",
		},
		{
			name: "invalid - invalid new status",
			log: &OrderLog{
				OrderID:    1,
				OperatorID: 2,
				Action:     "change_status",
				NewStatus:  OrderStatus("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.log.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderLog.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err.Error() != tt.errMsg && err.Error()[:len(tt.errMsg)] != tt.errMsg {
				t.Errorf("OrderLog.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestNewOrderLog(t *testing.T) {
	log := NewOrderLog(1, 2, "create_order", StatusPending, "")

	if log.OrderID != 1 {
		t.Errorf("NewOrderLog() OrderID = %d, want 1", log.OrderID)
	}
	if log.OperatorID != 2 {
		t.Errorf("NewOrderLog() OperatorID = %d, want 2", log.OperatorID)
	}
	if log.Action != "create_order" {
		t.Errorf("NewOrderLog() Action = %s, want create_order", log.Action)
	}
	if log.NewStatus != StatusPending {
		t.Errorf("NewOrderLog() NewStatus = %s, want %s", log.NewStatus, StatusPending)
	}
	if log.CreatedAt.IsZero() {
		t.Error("NewOrderLog() CreatedAt should be set")
	}
}

func TestOrderLogWithStatusChange(t *testing.T) {
	log := NewOrderLog(1, 2, "complete_order", StatusCompleted, StatusPending)

	if log.OldStatus != StatusPending {
		t.Errorf("OrderLog OldStatus = %s, want %s", log.OldStatus, StatusPending)
	}
	if log.NewStatus != StatusCompleted {
		t.Errorf("OrderLog NewStatus = %s, want %s", log.NewStatus, StatusCompleted)
	}
}
