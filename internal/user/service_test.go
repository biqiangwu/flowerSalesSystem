package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
)

// setupTestService 创建测试服务和数据库
func setupTestService(t *testing.T) (UserService, *sql.DB) {
	t.Helper()

	db := setupTestDB(t)
	repo := NewMySQLUserRepository(db)
	service := NewUserService(repo)

	return service, db
}

// createTestUser 创建测试用户辅助函数
func createTestUser(t *testing.T, ctx context.Context, repo UserRepository, username string, role Role) *User {
	t.Helper()
	user := &User{
		Username:     username,
		PasswordHash: "hashedpassword",
		Role:         role,
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

// TestUserService_ListUsers 测试获取用户列表
func TestUserService_ListUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, db := setupTestService(t)
	ctx := context.Background()

	// 创建多个测试用户
	_ = []*User{
		createTestUser(t, ctx, NewMySQLUserRepository(db), "listuser1", RoleCustomer),
		createTestUser(t, ctx, NewMySQLUserRepository(db), "listuser2", RoleClerk),
		createTestUser(t, ctx, NewMySQLUserRepository(db), "listuser3", RoleAdmin),
		createTestUser(t, ctx, NewMySQLUserRepository(db), "listuser4", RoleCustomer),
		createTestUser(t, ctx, NewMySQLUserRepository(db), "listuser5", RoleCustomer),
	}

	tests := []struct {
		name       string
		page       int
		pageSize   int
		minCount   int
		maxCount   int
		wantErr    bool
	}{
		{
			name:     "list first page with 3 items",
			page:     1,
			pageSize: 3,
			minCount: 3,
			maxCount: 3,
			wantErr:  false,
		},
		{
			name:     "list second page with 2 items",
			page:     2,
			pageSize: 3,
			minCount: 2,
			maxCount: 2,
			wantErr:  false,
		},
		{
			name:     "list page with default page size",
			page:     1,
			pageSize: 10,
			minCount: 5,
			maxCount: 5,
			wantErr:  false,
		},
		{
			name:     "list empty page beyond available users",
			page:     100,
			pageSize: 10,
			minCount: 0,
			maxCount: 0,
			wantErr:  false,
		},
		{
			name:     "list with zero page size should use default (100)",
			page:     1,
			pageSize: 0,
			minCount: 5,
			maxCount: 5,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ListUsers(ctx, tt.page, tt.pageSize)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) < tt.minCount || len(result) > tt.maxCount {
					t.Errorf("ListUsers() count = %d, want between %d and %d", len(result), tt.minCount, tt.maxCount)
				}
			}
		})
	}
}

// TestUserService_DeleteUser 测试删除用户
func TestUserService_DeleteUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, db := setupTestService(t)
	ctx := context.Background()

	// 创建测试用户
	user1 := createTestUser(t, ctx, NewMySQLUserRepository(db), "deleteuser1", RoleCustomer)
	user2 := createTestUser(t, ctx, NewMySQLUserRepository(db), "deleteuser2", RoleClerk)

	tests := []struct {
		name        string
		userID      int
		operatorID  int
		operatorRole Role
		wantErr     bool
		errType     error
	}{
		{
			name:        "admin deletes user",
			userID:      user1.ID,
			operatorID:  999, // 不存在的管理员ID
			operatorRole: RoleAdmin,
			wantErr:     false,
		},
		{
			name:        "clerk deletes user",
			userID:      user2.ID,
			operatorID:  998, // 不存在的店员ID
			operatorRole: RoleClerk,
			wantErr:     false,
		},
		{
			name:        "delete non-existing user",
			userID:      99999,
			operatorID:  997,
			operatorRole: RoleAdmin,
			wantErr:     true,
			errType:     ErrUserNotFound,
		},
		{
			name:        "customer cannot delete user - insufficient permission",
			userID:      user1.ID,
			operatorID:  996,
			operatorRole: RoleCustomer,
			wantErr:     true,
			errType:     ErrInsufficientPermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteUser(ctx, tt.userID, tt.operatorID, tt.operatorRole)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.Is(err, tt.errType) && err.Error() != tt.errType.Error() {
					t.Errorf("DeleteUser() error type = %v, want %v", err, tt.errType)
				}
			}

			// 验证删除后用户不存在
			if !tt.wantErr {
				repo := NewMySQLUserRepository(db)
				_, err := repo.GetByID(ctx, tt.userID)
				if err == nil {
					t.Error("DeleteUser() user still exists after deletion")
				}
			}
		})
	}
}

// TestUserService_ResetPassword 测试重置用户密码
func TestUserService_ResetPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, db := setupTestService(t)
	ctx := context.Background()

	// 创建测试用户
	user1 := createTestUser(t, ctx, NewMySQLUserRepository(db), "resetuser1", RoleCustomer)
	user2 := createTestUser(t, ctx, NewMySQLUserRepository(db), "resetuser2", RoleClerk)

	oldPasswordHash := user1.PasswordHash

	tests := []struct {
		name         string
		userID       int
		newPassword  string
		operatorID   int
		operatorRole Role
		wantErr      bool
		errType      error
	}{
		{
			name:         "admin resets user password",
			userID:       user1.ID,
			newPassword:  "newpassword123",
			operatorID:   999,
			operatorRole: RoleAdmin,
			wantErr:      false,
		},
		{
			name:         "clerk resets user password",
			userID:       user2.ID,
			newPassword:  "anotherpassword456",
			operatorID:   998,
			operatorRole: RoleClerk,
			wantErr:      false,
		},
		{
			name:         "reset non-existing user password",
			userID:       99999,
			newPassword:  "newpassword789",
			operatorID:   997,
			operatorRole: RoleAdmin,
			wantErr:      true,
			errType:      ErrUserNotFound,
		},
		{
			name:         "customer cannot reset password - insufficient permission",
			userID:       user1.ID,
			newPassword:  "hackedpassword",
			operatorID:   996,
			operatorRole: RoleCustomer,
			wantErr:      true,
			errType:      ErrInsufficientPermission,
		},
		{
			name:         "reset with empty password",
			userID:       user1.ID,
			newPassword:  "",
			operatorID:   999,
			operatorRole: RoleAdmin,
			wantErr:      true,
			errType:      ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 如果是第一次测试，需要重置 oldPasswordHash
			if tt.name == "admin resets user password" {
				oldPasswordHash = user1.PasswordHash
			}

			err := service.ResetPassword(ctx, tt.userID, tt.newPassword, tt.operatorID, tt.operatorRole)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResetPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.Is(err, tt.errType) && err.Error() != tt.errType.Error() {
					t.Errorf("ResetPassword() error type = %v, want %v", err, tt.errType)
				}
			}

			// 验证密码已更新
			if !tt.wantErr {
				repo := NewMySQLUserRepository(db)
				user, err := repo.GetByID(ctx, tt.userID)
				if err != nil {
					t.Errorf("GetByID() after ResetPassword error = %v", err)
					return
				}
				if user.PasswordHash == oldPasswordHash {
					t.Error("ResetPassword() password was not changed")
				}
			}
		})
	}
}

// TestUserService_PermissionValidation 测试权限验证
func TestUserService_PermissionValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	service, db := setupTestService(t)
	ctx := context.Background()

	tests := []struct {
		name         string
		operation    string
		operatorRole Role
		wantAllowed  bool
	}{
		{
			name:         "admin can delete user",
			operation:    "delete",
			operatorRole: RoleAdmin,
			wantAllowed:  true,
		},
		{
			name:         "clerk can delete user",
			operation:    "delete",
			operatorRole: RoleClerk,
			wantAllowed:  true,
		},
		{
			name:         "customer cannot delete user",
			operation:    "delete",
			operatorRole: RoleCustomer,
			wantAllowed:  false,
		},
		{
			name:         "admin can reset password",
			operation:    "reset_password",
			operatorRole: RoleAdmin,
			wantAllowed:  true,
		},
		{
			name:         "clerk can reset password",
			operation:    "reset_password",
			operatorRole: RoleClerk,
			wantAllowed:  true,
		},
		{
			name:         "customer cannot reset password",
			operation:    "reset_password",
			operatorRole: RoleCustomer,
			wantAllowed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			repo := NewMySQLUserRepository(db)

			// 为每个测试用例创建独立的操作者和目标用户
			operator := createTestUser(t, ctx, repo, "operator_"+tt.name, tt.operatorRole)
			target := createTestUser(t, ctx, repo, "target_"+tt.name, RoleCustomer)

			// 根据操作类型执行相应的操作
			switch tt.operation {
			case "delete":
				err = service.DeleteUser(ctx, target.ID, operator.ID, tt.operatorRole)
			case "reset_password":
				err = service.ResetPassword(ctx, target.ID, "newpass123", operator.ID, tt.operatorRole)
			}

			// 检查操作是否被允许
			if tt.wantAllowed {
				if err != nil {
					t.Errorf("%s: expected operation to be allowed, but got error: %v", tt.name, err)
				}
			} else {
				if err == nil {
					t.Errorf("%s: expected operation to be denied, but got no error", tt.name)
				}
				if !errors.Is(err, ErrInsufficientPermission) && err.Error() != ErrInsufficientPermission.Error() {
					t.Errorf("%s: expected ErrInsufficientPermission, got: %v", tt.name, err)
				}
			}
		})
	}
}
