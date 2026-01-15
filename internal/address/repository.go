package address

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// AddressRepository 定义地址数据访问接口
type AddressRepository interface {
	Create(ctx context.Context, a *Address) error
	GetByID(ctx context.Context, id int) (*Address, error)
	ListByUserID(ctx context.Context, userID int) ([]*Address, error)
	Update(ctx context.Context, a *Address) error
	Delete(ctx context.Context, id int) error
}

// addressRepository 实现 AddressRepository 接口
type addressRepository struct {
	db *sql.DB
}

// NewAddressRepository 创建 AddressRepository 实例
func NewAddressRepository(db *sql.DB) AddressRepository {
	return &addressRepository{db: db}
}

// Create 创建地址
func (r *addressRepository) Create(ctx context.Context, a *Address) error {
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	query := `
		INSERT INTO addresses (user_id, label, address, contact, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		a.UserID, a.Label, a.Address, a.Contact, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create address: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	a.ID = int(id)
	return nil
}

// GetByID 根据 ID 获取地址
func (r *addressRepository) GetByID(ctx context.Context, id int) (*Address, error) {
	query := `
		SELECT id, user_id, label, address, contact, created_at, updated_at
		FROM addresses WHERE id = ?
	`

	var a Address
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.UserID, &a.Label, &a.Address, &a.Contact,
		&a.CreatedAt, &a.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("address not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get address by id: %w", err)
	}

	return &a, nil
}

// ListByUserID 根据用户 ID 获取地址列表
func (r *addressRepository) ListByUserID(ctx context.Context, userID int) ([]*Address, error) {
	query := `
		SELECT id, user_id, label, address, contact, created_at, updated_at
		FROM addresses WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	defer rows.Close()

	var addresses []*Address
	for rows.Next() {
		var a Address
		err := rows.Scan(
			&a.ID, &a.UserID, &a.Label, &a.Address, &a.Contact,
			&a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan address: %w", err)
		}
		addresses = append(addresses, &a)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate addresses: %w", err)
	}

	return addresses, nil
}

// Update 更新地址信息
func (r *addressRepository) Update(ctx context.Context, a *Address) error {
	a.UpdatedAt = time.Now()

	query := `
		UPDATE addresses SET
			label = ?, address = ?, contact = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		a.Label, a.Address, a.Contact, a.UpdatedAt, a.ID,
	)
	if err != nil {
		return fmt.Errorf("update address: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("address not found: %d", a.ID)
	}

	return nil
}

// Delete 删除地址
func (r *addressRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM addresses WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete address: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("address not found: %d", id)
	}

	return nil
}
