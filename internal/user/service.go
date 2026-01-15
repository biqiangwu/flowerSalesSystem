package user

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// 错误定义
var (
	ErrUserNotFound          = errors.New("用户不存在")
	ErrInsufficientPermission = errors.New("权限不足")
	ErrInvalidPassword       = errors.New("密码无效")
)

// UserService 定义用户管理业务逻辑接口
type UserService interface {
	ListUsers(ctx context.Context, page, pageSize int) ([]*User, error)
	DeleteUser(ctx context.Context, userID int, operatorID int, operatorRole Role) error
	ResetPassword(ctx context.Context, userID int, newPassword string, operatorID int, operatorRole Role) error
}

// userService 实现 UserService 接口
type userService struct {
	repo UserRepository
}

// NewUserService 创建 UserService 实例
func NewUserService(repo UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*User, error) {
	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100 // 默认每页 100 条
	}

	// 调用 repository 获取用户列表
	users, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	return users, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, userID int, operatorID int, operatorRole Role) error {
	// 权限验证：只有 admin 和 clerk 可以删除用户
	if !s.canDeleteUser(operatorRole) {
		return ErrInsufficientPermission
	}

	// 检查用户是否存在
	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// 删除用户
	if err := s.repo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// ResetPassword 重置用户密码
func (s *userService) ResetPassword(ctx context.Context, userID int, newPassword string, operatorID int, operatorRole Role) error {
	// 权限验证：只有 admin 和 clerk 可以重置密码
	if !s.canResetPassword(operatorRole) {
		return ErrInsufficientPermission
	}

	// 验证新密码
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// 检查用户是否存在
	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// 哈希新密码
	passwordHash, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	// 更新密码
	if err := s.repo.UpdatePassword(ctx, userID, passwordHash); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// canDeleteUser 检查角色是否可以删除用户
func (s *userService) canDeleteUser(role Role) bool {
	return role == RoleAdmin || role == RoleClerk
}

// canResetPassword 检查角色是否可以重置密码
func (s *userService) canResetPassword(role Role) bool {
	return role == RoleAdmin || role == RoleClerk
}

// validatePassword 验证密码强度
func (s *userService) validatePassword(password string) error {
	if len(password) == 0 {
		return ErrInvalidPassword
	}
	if len(password) < 6 {
		return ErrInvalidPassword
	}
	return nil
}

// hashPassword 对密码进行哈希
func (s *userService) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
