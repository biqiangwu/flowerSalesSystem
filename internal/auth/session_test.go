package auth

import (
	"context"
	"testing"
	"time"

	"github.com/biqiangwu/flowerSalesSystem/internal/user"
)

// TestMemorySessionManager_CreateSession 测试创建 Session
func TestMemorySessionManager_CreateSession(t *testing.T) {
	mgr := NewMemorySessionManager()
	ctx := context.Background()

	session, err := mgr.CreateSession(ctx, 1, "testuser", user.RoleCustomer)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if session == nil {
		t.Fatal("CreateSession() session is nil")
	}

	if session.Token == "" {
		t.Error("CreateSession() token is empty")
	}

	if session.UserID != 1 {
		t.Errorf("CreateSession() userID = %d, want 1", session.UserID)
	}

	if session.Username != "testuser" {
		t.Errorf("CreateSession() username = %q, want %q", session.Username, "testuser")
	}

	if session.Role != user.RoleCustomer {
		t.Errorf("CreateSession() role = %q, want %q", session.Role, user.RoleCustomer)
	}

	if session.ExpiresAt.Before(time.Now()) {
		t.Error("CreateSession() session already expired")
	}
}

// TestMemorySessionManager_CreateSession_UniqueTokens 测试生成唯一 Token
func TestMemorySessionManager_CreateSession_UniqueTokens(t *testing.T) {
	mgr := NewMemorySessionManager()
	ctx := context.Background()

	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		session, err := mgr.CreateSession(ctx, i, "user", user.RoleCustomer)
		if err != nil {
			t.Fatalf("CreateSession() error = %v", err)
		}

		if tokens[session.Token] {
			t.Errorf("CreateSession() duplicate token generated: %s", session.Token)
		}
		tokens[session.Token] = true
	}
}

// TestMemorySessionManager_ValidateSession 测试验证 Session
func TestMemorySessionManager_ValidateSession(t *testing.T) {
	mgr := NewMemorySessionManager()
	ctx := context.Background()

	// 创建 Session
	created, err := mgr.CreateSession(ctx, 1, "testuser", user.RoleAdmin)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// 验证有效 Session
	validated, err := mgr.ValidateSession(ctx, created.Token)
	if err != nil {
		t.Errorf("ValidateSession() error = %v", err)
		return
	}

	if validated == nil {
		t.Error("ValidateSession() session is nil")
		return
	}

	if validated.Token != created.Token {
		t.Errorf("ValidateSession() token = %q, want %q", validated.Token, created.Token)
	}

	if validated.UserID != created.UserID {
		t.Errorf("ValidateSession() userID = %d, want %d", validated.UserID, created.UserID)
	}
}

// TestMemorySessionManager_ValidateSession_Invalid 测试验证无效 Session
func TestMemorySessionManager_ValidateSession_Invalid(t *testing.T) {
	mgr := NewMemorySessionManager()
	ctx := context.Background()

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "non-existent token",
			token: "nonexistent_token_12345",
		},
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "random string",
			token: "random_string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.ValidateSession(ctx, tt.token)
			if err == nil {
				t.Error("ValidateSession() expected error for invalid token, got nil")
			}
		})
	}
}

// TestMemorySessionManager_DeleteSession 测试删除 Session
func TestMemorySessionManager_DeleteSession(t *testing.T) {
	mgr := NewMemorySessionManager()
	ctx := context.Background()

	// 创建 Session
	created, err := mgr.CreateSession(ctx, 1, "testuser", user.RoleCustomer)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// 删除 Session
	err = mgr.DeleteSession(ctx, created.Token)
	if err != nil {
		t.Errorf("DeleteSession() error = %v", err)
		return
	}

	// 验证 Session 已删除
	_, err = mgr.ValidateSession(ctx, created.Token)
	if err == nil {
		t.Error("DeleteSession() session still valid after deletion")
	}
}

// TestMemorySessionManager_DeleteSession_NonExistent 测试删除不存在的 Session
func TestMemorySessionManager_DeleteSession_NonExistent(t *testing.T) {
	mgr := NewMemorySessionManager()
	ctx := context.Background()

	err := mgr.DeleteSession(ctx, "nonexistent_token")
	// 删除不存在的 session 不应该报错（幂等操作）
	if err != nil {
		t.Errorf("DeleteSession() error = %v", err)
	}
}

// TestMemorySessionManager_CleanupExpiredSessions 测试清理过期 Session
func TestMemorySessionManager_CleanupExpiredSessions(t *testing.T) {
	mgr := NewMemorySessionManager()
	ctx := context.Background()

	// 创建一个 Session 并手动设置为已过期
	session, err := mgr.CreateSession(ctx, 1, "testuser", user.RoleCustomer)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// 手动修改 session 使其过期（通过 Delete 然后手动添加过期 session）
	mgr.DeleteSession(ctx, session.Token)
	expiredSession := &Session{
		Token:     session.Token,
		UserID:    session.UserID,
		Username:  session.Username,
		Role:      session.Role,
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	mgr.sessions[session.Token] = expiredSession

	// 清理过期 Session
	err = mgr.CleanupExpiredSessions(ctx)
	if err != nil {
		t.Errorf("CleanupExpiredSessions() error = %v", err)
	}

	// 验证过期 Session 已被清理
	_, err = mgr.ValidateSession(ctx, session.Token)
	if err == nil {
		t.Error("CleanupExpiredSessions() expired session still valid")
	}
}

// TestMemorySessionManager_ConcurrentAccess 测试并发访问
func TestMemorySessionManager_ConcurrentAccess(t *testing.T) {
	mgr := NewMemorySessionManager()
	ctx := context.Background()

	// 并发创建多个 Session
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			_, err := mgr.CreateSession(ctx, idx, "user", user.RoleCustomer)
			if err != nil {
				t.Errorf("CreateSession() concurrent error = %v", err)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证所有 Session 都可以被访问
	sessions := make(map[string]bool)
	for i := 0; i < 10; i++ {
		session, err := mgr.CreateSession(ctx, i+100, "user2", user.RoleCustomer)
		if err != nil {
			t.Errorf("CreateSession() error = %v", err)
		}
		if sessions[session.Token] {
			t.Error("CreateSession() duplicate token found")
		}
		sessions[session.Token] = true
	}
}
