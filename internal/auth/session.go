package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/biqiangwu/flowerSalesSystem/internal/user"
)

// Session 用户会话
type Session struct {
	Token     string
	UserID    int
	Username  string
	Role      user.Role
	ExpiresAt time.Time
}

// SessionManager Session 管理接口
type SessionManager interface {
	CreateSession(ctx context.Context, userID int, username string, role user.Role) (*Session, error)
	ValidateSession(ctx context.Context, token string) (*Session, error)
	DeleteSession(ctx context.Context, token string) error
	CleanupExpiredSessions(ctx context.Context) error
}

// MemorySessionManager 内存 Session 管理实现
type MemorySessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewMemorySessionManager 创建内存 Session 管理器
func NewMemorySessionManager() *MemorySessionManager {
	return &MemorySessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession 创建新 Session
func (m *MemorySessionManager) CreateSession(ctx context.Context, userID int, username string, role user.Role) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	session := &Session{
		Token:     token,
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 默认 24 小时过期
	}

	m.mu.Lock()
	m.sessions[token] = session
	m.mu.Unlock()

	return session, nil
}

// ValidateSession 验证 Session
func (m *MemorySessionManager) ValidateSession(ctx context.Context, token string) (*Session, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}

	m.mu.RLock()
	session, exists := m.sessions[token]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// DeleteSession 删除 Session
func (m *MemorySessionManager) DeleteSession(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("empty token")
	}

	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()

	return nil
}

// CleanupExpiredSessions 清理过期的 Session
func (m *MemorySessionManager) CleanupExpiredSessions(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for token, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			delete(m.sessions, token)
		}
	}

	return nil
}

// generateToken 生成随机 Token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
