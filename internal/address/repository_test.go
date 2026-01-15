package address

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

	// 创建测试表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS addresses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		label TEXT,
		address TEXT NOT NULL,
		contact TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		t.Fatalf("failed to create addresses table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestAddressRepository_Create 测试 Create 方法
func TestAddressRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewAddressRepository(db)

	tests := []struct {
		name    string
		address *Address
		wantErr bool
	}{
		{
			name: "create valid address with label",
			address: &Address{
				UserID:  1,
				Label:   "家",
				Address: "北京市朝阳区xxx街道xxx号",
				Contact: "张三，13800138000",
			},
			wantErr: false,
		},
		{
			name: "create valid address without label",
			address: &Address{
				UserID:  2,
				Address: "上海市浦东新区yyy路yyy号",
				Contact: "李四，13900139000",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := repo.Create(ctx, tt.address)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.address.ID <= 0 {
				t.Errorf("Create() ID = %d, want > 0", tt.address.ID)
			}

			if !tt.wantErr && tt.address.CreatedAt.IsZero() {
				t.Errorf("Create() CreatedAt is zero")
			}

			if !tt.wantErr && tt.address.UpdatedAt.IsZero() {
				t.Errorf("Create() UpdatedAt is zero")
			}
		})
	}
}

// TestAddressRepository_GetByID 测试 GetByID 方法
func TestAddressRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewAddressRepository(db)
	ctx := context.Background()

	// 创建测试地址
	createdAddress := &Address{
		UserID:  1,
		Label:   "公司",
		Address: "深圳市南山区zzz大厦",
		Contact: "王五，13800138000",
	}
	if err := repo.Create(ctx, createdAddress); err != nil {
		t.Fatalf("failed to create address: %v", err)
	}

	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "get existing address",
			id:      createdAddress.ID,
			wantErr: false,
		},
		{
			name:    "get non-existing address",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := repo.GetByID(ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if address == nil {
					t.Error("GetByID() address is nil")
					return
				}
				if address.ID != tt.id {
					t.Errorf("GetByID() ID = %d, want %d", address.ID, tt.id)
				}
			}
		})
	}
}

// TestAddressRepository_ListByUserID 测试 ListByUserID 方法
func TestAddressRepository_ListByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewAddressRepository(db)
	ctx := context.Background()

	// 为用户1创建多个地址
	addresses := []*Address{
		{
			UserID:  1,
			Label:   "家",
			Address: "北京市朝阳区xxx街道xxx号",
			Contact: "张三，13800138000",
		},
		{
			UserID:  1,
			Label:   "公司",
			Address: "深圳市南山区zzz大厦",
			Contact: "王五，13800138000",
		},
		{
			UserID:  2,
			Label:   "家",
			Address: "上海市浦东新区yyy路yyy号",
			Contact: "李四，13900139000",
		},
	}

	for _, a := range addresses {
		if err := repo.Create(ctx, a); err != nil {
			t.Fatalf("failed to create address: %v", err)
		}
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
			result, err := repo.ListByUserID(ctx, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListByUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) < tt.minCount || len(result) > tt.maxCount {
					t.Errorf("ListByUserID() count = %d, want between %d and %d", len(result), tt.minCount, tt.maxCount)
				}
			}
		})
	}
}

// TestAddressRepository_Update 测试 Update 方法
func TestAddressRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewAddressRepository(db)
	ctx := context.Background()

	// 创建测试地址
	address := &Address{
		UserID:  1,
		Label:   "家",
		Address: "北京市朝阳区xxx街道xxx号",
		Contact: "张三，13800138000",
	}
	if err := repo.Create(ctx, address); err != nil {
		t.Fatalf("failed to create address: %v", err)
	}

	// 修改地址信息
	address.Label = "新家"
	address.Address = "北京市海淀区aaa路bbb号"
	address.Contact = "张三，13700137000"

	err := repo.Update(ctx, address)
	if err != nil {
		t.Errorf("Update() error = %v", err)
		return
	}

	// 验证更新
	updated, err := repo.GetByID(ctx, address.ID)
	if err != nil {
		t.Errorf("GetByID() after Update error = %v", err)
		return
	}

	if updated.Label != "新家" {
		t.Errorf("Update() Label = %q, want %q", updated.Label, "新家")
	}
	if updated.Address != "北京市海淀区aaa路bbb号" {
		t.Errorf("Update() Address = %q, want %q", updated.Address, "北京市海淀区aaa路bbb号")
	}
	if updated.Contact != "张三，13700137000" {
		t.Errorf("Update() Contact = %q, want %q", updated.Contact, "张三，13700137000")
	}
}

// TestAddressRepository_Delete 测试 Delete 方法
func TestAddressRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupTestDB(t)
	repo := NewAddressRepository(db)
	ctx := context.Background()

	// 创建测试地址
	address := &Address{
		UserID:  1,
		Label:   "家",
		Address: "北京市朝阳区xxx街道xxx号",
		Contact: "张三，13800138000",
	}
	if err := repo.Create(ctx, address); err != nil {
		t.Fatalf("failed to create address: %v", err)
	}

	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "delete existing address",
			id:      address.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existing address",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 验证删除后地址不存在
			if !tt.wantErr {
				_, err := repo.GetByID(ctx, tt.id)
				if err == nil {
					t.Error("Delete() address still exists after deletion")
				}
			}
		})
	}
}
