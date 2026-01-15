package auth

import (
	"context"
	"fmt"

	"github.com/biqiangwu/flowerSalesSystem/internal/user"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务接口
type AuthService interface {
	Register(ctx context.Context, username, password string) (*user.User, error)
	Login(ctx context.Context, username, password string) (*Session, error)
	Logout(ctx context.Context, sessionToken string) error
	ValidateSession(ctx context.Context, sessionToken string) (*user.User, error)
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) bool
}

// authService 认证服务实现
type authService struct {
	userRepo    user.UserRepository
	sessionMgr  SessionManager
	minPwdLen   int
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo user.UserRepository, sessionMgr SessionManager) AuthService {
	return &authService{
		userRepo:   userRepo,
		sessionMgr: sessionMgr,
		minPwdLen:  6, // 最小密码长度 6 位
	}
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, username, password string) (*user.User, error) {
	// 验证用户名
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	// 验证密码
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	if len(password) < s.minPwdLen {
		return nil, fmt.Errorf("password must be at least %d characters", s.minPwdLen)
	}

	// 检查用户名是否已存在
	_, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil {
		return nil, fmt.Errorf("username already exists")
	}

	// 哈希密码
	hash, err := s.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	u := &user.User{
		Username:     username,
		PasswordHash: hash,
		Role:         user.RoleCustomer, // 默认角色为 customer
	}

	err = s.userRepo.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return u, nil
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, username, password string) (*Session, error) {
	// 验证输入
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	// 获取用户
	u, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 验证密码
	if !s.VerifyPassword(password, u.PasswordHash) {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 创建 Session
	session, err := s.sessionMgr.CreateSession(ctx, u.ID, u.Username, u.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		return fmt.Errorf("session token cannot be empty")
	}

	err := s.sessionMgr.DeleteSession(ctx, sessionToken)
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}

// ValidateSession 验证 Session 并返回用户信息
func (s *authService) ValidateSession(ctx context.Context, sessionToken string) (*user.User, error) {
	if sessionToken == "" {
		return nil, fmt.Errorf("session token cannot be empty")
	}

	session, err := s.sessionMgr.ValidateSession(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// 从 Session 中获取用户信息（为了获取最新信息，可以从数据库重新查询）
	u, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return u, nil
}

// HashPassword 哈希密码
func (s *authService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if len(password) < s.minPwdLen {
		return "", fmt.Errorf("password must be at least %d characters", s.minPwdLen)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate hash: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword 验证密码
func (s *authService) VerifyPassword(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
