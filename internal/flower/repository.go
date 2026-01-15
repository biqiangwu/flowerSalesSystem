package flower

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// FlowerRepository 定义鲜花数据访问接口
type FlowerRepository interface {
	Create(ctx context.Context, f *Flower) error
	GetBySKU(ctx context.Context, sku string) (*Flower, error)
	List(ctx context.Context, filter FlowerFilter) ([]*Flower, error)
	Update(ctx context.Context, f *Flower) error
	Delete(ctx context.Context, sku string) error
	UpdateStock(ctx context.Context, sku string, delta int) error
}

// flowerRepository 实现 FlowerRepository 接口
type flowerRepository struct {
	db *sql.DB
}

// NewFlowerRepository 创建 FlowerRepository 实例
func NewFlowerRepository(db *sql.DB) FlowerRepository {
	return &flowerRepository{db: db}
}

// Create 创建鲜花
func (r *flowerRepository) Create(ctx context.Context, f *Flower) error {
	now := time.Now()
	f.CreatedAt = now
	f.UpdatedAt = now

	query := `
		INSERT INTO flowers (sku, name, origin, shelf_life, preservation,
			purchase_price, sale_price, stock, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// SQLite 使用 1/0 表示布尔值，MySQL 使用 TRUE/FALSE
	isActive := 0
	if f.IsActive {
		isActive = 1
	}

	result, err := r.db.ExecContext(ctx, query,
		f.SKU, f.Name, f.Origin, f.ShelfLife, f.Preservation,
		f.PurchasePrice.Value, f.SalePrice.Value, f.Stock, isActive,
		f.CreatedAt, f.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create flower: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no rows inserted")
	}

	return nil
}

// GetBySKU 根据 SKU 获取鲜花
func (r *flowerRepository) GetBySKU(ctx context.Context, sku string) (*Flower, error) {
	query := `
		SELECT sku, name, origin, shelf_life, preservation,
			purchase_price, sale_price, stock, is_active, created_at, updated_at
		FROM flowers WHERE sku = ?
	`

	var f Flower
	var isActive int
	var purchasePrice, salePrice int64

	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&f.SKU, &f.Name, &f.Origin, &f.ShelfLife, &f.Preservation,
		&purchasePrice, &salePrice, &f.Stock, &isActive,
		&f.CreatedAt, &f.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flower not found: %s", sku)
	}
	if err != nil {
		return nil, fmt.Errorf("get flower by sku: %w", err)
	}

	f.PurchasePrice = Decimal{Value: purchasePrice}
	f.SalePrice = Decimal{Value: salePrice}
	f.IsActive = isActive != 0

	return &f, nil
}

// List 根据筛选条件获取鲜花列表
func (r *flowerRepository) List(ctx context.Context, filter FlowerFilter) ([]*Flower, error) {
	query := `
		SELECT sku, name, origin, shelf_life, preservation,
			purchase_price, sale_price, stock, is_active, created_at, updated_at
		FROM flowers WHERE 1=1
	`
	args := []interface{}{}

	// 搜索条件
	if filter.Search != "" {
		query += " AND (sku LIKE ? OR name LIKE ?)"
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// 产地筛选
	if filter.Origin != "" {
		query += " AND origin = ?"
		args = append(args, filter.Origin)
	}

	// 价格区间筛选
	if filter.MinPrice > 0 {
		// 将元转换为分
		query += " AND sale_price >= ?"
		args = append(args, int64(filter.MinPrice*100))
	}
	if filter.MaxPrice > 0 {
		query += " AND sale_price <= ?"
		args = append(args, int64(filter.MaxPrice*100))
	}

	// 排序
	switch filter.SortBy {
	case "price_asc":
		query += " ORDER BY sale_price ASC"
	case "price_desc":
		query += " ORDER BY sale_price DESC"
	case "stock":
		query += " ORDER BY stock ASC"
	default:
		query += " ORDER BY created_at DESC"
	}

	// 分页
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += " LIMIT ? OFFSET ?"
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list flowers: %w", err)
	}
	defer rows.Close()

	var flowers []*Flower
	for rows.Next() {
		var f Flower
		var isActive int
		var purchasePrice, salePrice int64

		err := rows.Scan(
			&f.SKU, &f.Name, &f.Origin, &f.ShelfLife, &f.Preservation,
			&purchasePrice, &salePrice, &f.Stock, &isActive,
			&f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan flower: %w", err)
		}

		f.PurchasePrice = Decimal{Value: purchasePrice}
		f.SalePrice = Decimal{Value: salePrice}
		f.IsActive = isActive != 0

		flowers = append(flowers, &f)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate flowers: %w", err)
	}

	return flowers, nil
}

// Update 更新鲜花信息
func (r *flowerRepository) Update(ctx context.Context, f *Flower) error {
	f.UpdatedAt = time.Now()

	query := `
		UPDATE flowers SET
			name = ?, origin = ?, shelf_life = ?, preservation = ?,
			purchase_price = ?, sale_price = ?, stock = ?, is_active = ?,
			updated_at = ?
		WHERE sku = ?
	`

	isActive := 0
	if f.IsActive {
		isActive = 1
	}

	result, err := r.db.ExecContext(ctx, query,
		f.Name, f.Origin, f.ShelfLife, f.Preservation,
		f.PurchasePrice.Value, f.SalePrice.Value, f.Stock, isActive,
		f.UpdatedAt, f.SKU,
	)
	if err != nil {
		return fmt.Errorf("update flower: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("flower not found: %s", f.SKU)
	}

	return nil
}

// Delete 删除鲜花
func (r *flowerRepository) Delete(ctx context.Context, sku string) error {
	query := `DELETE FROM flowers WHERE sku = ?`

	result, err := r.db.ExecContext(ctx, query, sku)
	if err != nil {
		return fmt.Errorf("delete flower: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("flower not found: %s", sku)
	}

	return nil
}

// UpdateStock 更新库存（增量更新）
func (r *flowerRepository) UpdateStock(ctx context.Context, sku string, delta int) error {
	query := `UPDATE flowers SET stock = stock + ?, updated_at = ? WHERE sku = ?`

	result, err := r.db.ExecContext(ctx, query, delta, time.Now(), sku)
	if err != nil {
		return fmt.Errorf("update stock: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("flower not found: %s", sku)
	}

	return nil
}
