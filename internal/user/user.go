package user

import "time"

// Role 用户角色类型
type Role string

// 用户角色常量
const (
	RoleCustomer Role = "customer"
	RoleClerk    Role = "clerk"
	RoleAdmin    Role = "admin"
)

// User 用户领域模型
type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
