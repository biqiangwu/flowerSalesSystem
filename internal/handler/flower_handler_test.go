package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
	_ "github.com/mattn/go-sqlite3"
)

// setupFlowerTestDB 创建鲜花测试数据库
func setupFlowerTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS flowers (
		sku TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		origin TEXT NOT NULL,
		shelf_life TEXT,
		preservation TEXT,
		purchase_price INTEGER NOT NULL,
		sale_price INTEGER NOT NULL,
		stock INTEGER NOT NULL DEFAULT 0,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		t.Fatalf("failed to create flowers table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// setupFlowerTestHandler 创建测试用的 Handler
func setupFlowerTestHandler(t *testing.T) *Handler {
	t.Helper()

	db := setupFlowerTestDB(t)
	flowerRepo := flower.NewFlowerRepository(db)
	flowerSvc := flower.NewFlowerService(flowerRepo)

	h := &Handler{
		flowerService: flowerSvc,
	}

	return h
}

// TestHandleListFlowers 测试获取鲜花列表
func TestHandleListFlowers(t *testing.T) {
	handler := setupFlowerTestHandler(t)

	ctx := t.Context()
	for i := 1; i <= 3; i++ {
		req := &flower.CreateFlowerRequest{
			SKU:           fmt.Sprintf("FLW%03d", i),
			Name:          fmt.Sprintf("测试鲜花%d", i),
			Origin:        "云南",
			ShelfLife:     "7天",
			Preservation:  "冷藏",
			PurchasePrice: 10.0,
			SalePrice:     15.0,
			Stock:         100,
		}
		if err := handler.flowerService.CreateFlower(ctx, req); err != nil {
			t.Fatalf("failed to create test flower: %v", err)
		}
	}

	req := httptest.NewRequest("GET", "/api/flowers", nil)
	w := httptest.NewRecorder()

	handler.HandleListFlowers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleListFlowers() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp []json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp) != 3 {
		t.Errorf("got %d flowers, want 3", len(resp))
	}
}
