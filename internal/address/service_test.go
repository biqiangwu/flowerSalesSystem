package address

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
)

// setupTestService 创建测试服务和数据库
func setupTestService(t *testing.T) (AddressService, *sql.DB) {
	t.Helper()

	db := setupTestDB(t)
	repo := NewAddressRepository(db)
	service := NewAddressService(repo)

	return service, db
}

// TestAddressService_CreateAddress 测试创建地址
func TestAddressService_CreateAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		userID  int
		request *CreateAddressRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:   "create valid address with label",
			userID: 1,
			request: &CreateAddressRequest{
				Label:   stringPtr("家"),
				Address: "北京市朝阳区xxx街道xxx号",
				Contact: "张三，13800138000",
			},
			wantErr: false,
		},
		{
			name:   "create valid address without label",
			userID: 2,
			request: &CreateAddressRequest{
				Label:   nil,
				Address: "上海市浦东新区yyy路yyy号",
				Contact: "李四，13900139000",
			},
			wantErr: false,
		},
		{
			name:   "create with empty address",
			userID: 1,
			request: &CreateAddressRequest{
				Label:   stringPtr("公司"),
				Address: "",
				Contact: "张三，13800138000",
			},
			wantErr: true,
			errMsg:  "地址不能为空",
		},
		{
			name:   "create with empty contact",
			userID: 1,
			request: &CreateAddressRequest{
				Label:   stringPtr("公司"),
				Address: "北京市朝阳区xxx街道xxx号",
				Contact: "",
			},
			wantErr: true,
			errMsg:  "联系方式不能为空",
		},
		{
			name:   "create with short address",
			userID: 1,
			request: &CreateAddressRequest{
				Label:   stringPtr("公司"),
				Address: "abc",
				Contact: "张三，13800138000",
			},
			wantErr: true,
			errMsg:  "地址长度不能少于5个字符",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateAddress(ctx, tt.userID, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("CreateAddress() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestAddressService_GetAddress 测试获取地址
func TestAddressService_GetAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试地址
	createReq := &CreateAddressRequest{
		Label:   stringPtr("公司"),
		Address: "深圳市南山区zzz大厦",
		Contact: "王五，13800138000",
	}
	if err := service.CreateAddress(ctx, 1, createReq); err != nil {
		t.Fatalf("failed to create address: %v", err)
	}

	// 获取所有地址列表来找到创建的地址 ID
	addresses, err := service.ListAddresses(ctx, 1)
	if err != nil {
		t.Fatalf("failed to list addresses: %v", err)
	}
	if len(addresses) == 0 {
		t.Fatal("no addresses found")
	}
	addressID := addresses[0].ID

	tests := []struct {
		name    string
		userID  int
		id      int
		wantErr bool
	}{
		{
			name:    "get existing address",
			userID:  1,
			id:      addressID,
			wantErr: false,
		},
		{
			name:    "get non-existing address",
			userID:  1,
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := service.GetAddress(ctx, tt.userID, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if address == nil {
					t.Error("GetAddress() address is nil")
					return
				}
				if address.ID != tt.id {
					t.Errorf("GetAddress() ID = %d, want %d", address.ID, tt.id)
				}
			}
		})
	}
}

// TestAddressService_ListAddresses 测试获取地址列表
func TestAddressService_ListAddresses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 为用户1创建多个地址
	addresses := []*CreateAddressRequest{
		{
			Label:   stringPtr("家"),
			Address: "北京市朝阳区xxx街道xxx号",
			Contact: "张三，13800138000",
		},
		{
			Label:   stringPtr("公司"),
			Address: "深圳市南山区zzz大厦",
			Contact: "王五，13800138000",
		},
	}

	for _, req := range addresses {
		if err := service.CreateAddress(ctx, 1, req); err != nil {
			t.Fatalf("failed to create address: %v", err)
		}
	}

	// 为用户2创建地址
	req2 := &CreateAddressRequest{
		Label:   stringPtr("家"),
		Address: "上海市浦东新区yyy路yyy号",
		Contact: "李四，13900139000",
	}
	if err := service.CreateAddress(ctx, 2, req2); err != nil {
		t.Fatalf("failed to create address: %v", err)
	}

	tests := []struct {
		name     string
		userID   int
		minCount int
		maxCount int
		wantErr  bool
	}{
		{
			name:     "list addresses for user 1",
			userID:   1,
			minCount: 2,
			maxCount: 2,
			wantErr:  false,
		},
		{
			name:     "list addresses for user 2",
			userID:   2,
			minCount: 1,
			maxCount: 1,
			wantErr:  false,
		},
		{
			name:     "list addresses for non-existing user",
			userID:   999,
			minCount: 0,
			maxCount: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ListAddresses(ctx, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) < tt.minCount || len(result) > tt.maxCount {
					t.Errorf("ListAddresses() count = %d, want between %d and %d", len(result), tt.minCount, tt.maxCount)
				}
			}
		})
	}
}

// TestAddressService_UpdateAddress 测试更新地址
func TestAddressService_UpdateAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试地址
	createReq := &CreateAddressRequest{
		Label:   stringPtr("家"),
		Address: "北京市朝阳区xxx街道xxx号",
		Contact: "张三，13800138000",
	}
	if err := service.CreateAddress(ctx, 1, createReq); err != nil {
		t.Fatalf("failed to create address: %v", err)
	}

	// 获取创建的地址 ID
	addresses, err := service.ListAddresses(ctx, 1)
	if err != nil {
		t.Fatalf("failed to list addresses: %v", err)
	}
	if len(addresses) == 0 {
		t.Fatal("no addresses found")
	}
	addressID := addresses[0].ID

	tests := []struct {
		name    string
		userID  int
		id      int
		request *UpdateAddressRequest
		wantErr bool
	}{
		{
			name:   "update valid address",
			userID: 1,
			id:     addressID,
			request: &UpdateAddressRequest{
				Label:   stringPtr("新家"),
				Address: stringPtr("北京市海淀区aaa路bbb号"),
				Contact: stringPtr("张三，13700137000"),
			},
			wantErr: false,
		},
		{
			name:   "update non-existing address",
			userID: 1,
			id:     99999,
			request: &UpdateAddressRequest{
				Label:   stringPtr("新家"),
				Address: stringPtr("北京市海淀区aaa路bbb号"),
			},
			wantErr: true,
		},
		{
			name:   "update with empty address",
			userID: 1,
			id:     addressID,
			request: &UpdateAddressRequest{
				Address: stringPtr(""),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateAddress(ctx, tt.userID, tt.id, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAddressService_DeleteAddress 测试删除地址
func TestAddressService_DeleteAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试地址
	createReq := &CreateAddressRequest{
		Label:   stringPtr("家"),
		Address: "北京市朝阳区xxx街道xxx号",
		Contact: "张三，13800138000",
	}
	if err := service.CreateAddress(ctx, 1, createReq); err != nil {
		t.Fatalf("failed to create address: %v", err)
	}

	// 获取创建的地址 ID
	addresses, err := service.ListAddresses(ctx, 1)
	if err != nil {
		t.Fatalf("failed to list addresses: %v", err)
	}
	if len(addresses) == 0 {
		t.Fatal("no addresses found")
	}
	addressID := addresses[0].ID

	tests := []struct {
		name    string
		userID  int
		id      int
		wantErr bool
	}{
		{
			name:    "delete existing address",
			userID:  1,
			id:      addressID,
			wantErr: false,
		},
		{
			name:    "delete non-existing address",
			userID:  1,
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteAddress(ctx, tt.userID, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 验证删除后地址不存在
			if !tt.wantErr {
				_, err := service.GetAddress(ctx, tt.userID, tt.id)
				if err == nil {
					t.Error("DeleteAddress() address still exists after deletion")
				}
			}
		})
	}
}

// TestAddressService_UserCanOnlyAccessOwnAddresses 测试用户只能操作自己的地址
func TestAddressService_UserCanOnlyAccessOwnAddresses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, _ := setupTestService(t)
	ctx := context.Background()

	// 为用户1创建地址
	createReq1 := &CreateAddressRequest{
		Label:   stringPtr("家"),
		Address: "北京市朝阳区xxx街道xxx号",
		Contact: "张三，13800138000",
	}
	if err := service.CreateAddress(ctx, 1, createReq1); err != nil {
		t.Fatalf("failed to create address: %v", err)
	}

	// 获取用户1的地址 ID
	addresses1, err := service.ListAddresses(ctx, 1)
	if err != nil {
		t.Fatalf("failed to list addresses: %v", err)
	}
	if len(addresses1) == 0 {
		t.Fatal("no addresses found for user 1")
	}
	addressID := addresses1[0].ID

	// 用户2尝试访问用户1的地址
	_, err = service.GetAddress(ctx, 2, addressID)
	if err == nil {
		t.Error("Expected error when user tries to access another user's address")
	}

	// 用户2尝试删除用户1的地址
	err = service.DeleteAddress(ctx, 2, addressID)
	if err == nil {
		t.Error("Expected error when user tries to delete another user's address")
	}

	// 用户2尝试更新用户1的地址
	updateReq := &UpdateAddressRequest{
		Label: stringPtr("修改"),
	}
	err = service.UpdateAddress(ctx, 2, addressID, updateReq)
	if err == nil {
		t.Error("Expected error when user tries to update another user's address")
	}
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
