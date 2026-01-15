package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/biqiangwu/flowerSalesSystem/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

const (
	// DefaultMaxOpenConns 默认最大打开连接数
	DefaultMaxOpenConns = 25
	// DefaultMaxIdleConns 默认最大空闲连接数
	DefaultMaxIdleConns = 10
	// DefaultConnMaxLifetime 默认连接最大生命周期
	DefaultConnMaxLifetime = 5 * time.Minute
)

// Open 打开数据库连接并配置连接池
// 返回配置好的 *sql.DB 实例
func Open(cfg *config.Config) (*sql.DB, error) {
	dsn := buildDSN(cfg)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(DefaultMaxOpenConns)
	db.SetMaxIdleConns(DefaultMaxIdleConns)
	db.SetConnMaxLifetime(DefaultConnMaxLifetime)

	// 验证连接是否可用
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

// buildDSN 构建 MySQL Data Source Name (DSN)
// 格式: user:password@tcp(host:port)/dbname?parseTime=true
func buildDSN(cfg *config.Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
}
