package user

import (
	"context"
	"database/sql"
	"fmt"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	List(ctx context.Context, page, pageSize int) ([]*User, error)
	Delete(ctx context.Context, id int) error
	UpdatePassword(ctx context.Context, id int, passwordHash string) error
}

// MySQLUserRepository MySQL 用户数据访问实现
type MySQLUserRepository struct {
	db *sql.DB
}

// NewMySQLUserRepository 创建 MySQL 用户 Repository
func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

// Create 创建新用户
func (r *MySQLUserRepository) Create(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (username, password_hash, role)
		VALUES (?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query, u.Username, u.PasswordHash, u.Role)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	u.ID = int(id)

	// 获取数据库设置的时间戳
	err = r.db.QueryRowContext(ctx, "SELECT created_at, updated_at FROM users WHERE id = ?", u.ID).
		Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to get timestamps: %w", err)
	}

	return nil
}

// GetByID 根据 ID 获取用户
func (r *MySQLUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	user := &User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: id=%d", id)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

// GetByUsername 根据用户名获取用户
func (r *MySQLUserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = ?
	`
	user := &User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: username=%s", username)
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

// List 分页获取用户列表
func (r *MySQLUserRepository) List(ctx context.Context, page, pageSize int) ([]*User, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM users
		ORDER BY id
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// Delete 删除用户
func (r *MySQLUserRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: id=%d", id)
	}

	return nil
}

// UpdatePassword 更新用户密码
func (r *MySQLUserRepository) UpdatePassword(ctx context.Context, id int, passwordHash string) error {
	query := `UPDATE users SET password_hash = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, passwordHash, id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: id=%d", id)
	}

	return nil
}
