package user

import (
	"testing"
	"time"
)

// TestRoleConstants 测试 Role 类型常量的值
func TestRoleConstants(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected string
	}{
		{
			name:     "RoleCustomer",
			role:     RoleCustomer,
			expected: "customer",
		},
		{
			name:     "RoleClerk",
			role:     RoleClerk,
			expected: "clerk",
		},
		{
			name:     "RoleAdmin",
			role:     RoleAdmin,
			expected: "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.expected {
				t.Errorf("Role = %q, want %q", tt.role, tt.expected)
			}
		})
	}
}

// TestUserCreation 测试 User 结构体创建
func TestUserCreation(t *testing.T) {
	tests := []struct {
		name  string
		user  *User
		check func(*testing.T, *User)
	}{
		{
			name: "valid user with all fields",
			user: &User{
				ID:           1,
				Username:     "testuser",
				PasswordHash: "hashedpassword",
				Role:         RoleCustomer,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			check: func(t *testing.T, u *User) {
				if u.ID != 1 {
					t.Errorf("ID = %d, want 1", u.ID)
				}
				if u.Username != "testuser" {
					t.Errorf("Username = %q, want %q", u.Username, "testuser")
				}
				if u.PasswordHash != "hashedpassword" {
					t.Errorf("PasswordHash = %q, want %q", u.PasswordHash, "hashedpassword")
				}
				if u.Role != RoleCustomer {
					t.Errorf("Role = %q, want %q", u.Role, RoleCustomer)
				}
			},
		},
		{
			name: "user with admin role",
			user: &User{
				ID:           2,
				Username:     "admin",
				PasswordHash: "adminhash",
				Role:         RoleAdmin,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			check: func(t *testing.T, u *User) {
				if u.Role != RoleAdmin {
					t.Errorf("Role = %q, want %q", u.Role, RoleAdmin)
				}
			},
		},
		{
			name: "user with clerk role",
			user: &User{
				ID:           3,
				Username:     "clerk",
				PasswordHash: "clerkhash",
				Role:         RoleClerk,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			check: func(t *testing.T, u *User) {
				if u.Role != RoleClerk {
					t.Errorf("Role = %q, want %q", u.Role, RoleClerk)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.check != nil {
				tt.check(t, tt.user)
			}
		})
	}
}

// TestUserZeroValue 测试 User 结构体的零值
func TestUserZeroValue(t *testing.T) {
	var u User

	if u.ID != 0 {
		t.Errorf("ID zero value = %d, want 0", u.ID)
	}
	if u.Username != "" {
		t.Errorf("Username zero value = %q, want empty string", u.Username)
	}
	if u.PasswordHash != "" {
		t.Errorf("PasswordHash zero value = %q, want empty string", u.PasswordHash)
	}
	if u.Role != "" {
		t.Errorf("Role zero value = %q, want empty string", u.Role)
	}
	if !u.CreatedAt.IsZero() {
		t.Errorf("CreatedAt zero value = %v, want zero time", u.CreatedAt)
	}
	if !u.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt zero value = %v, want zero time", u.UpdatedAt)
	}
}
