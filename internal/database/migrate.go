package database

import (
	"database/sql"
	"embed"
	"fmt"
)

//go:embed schema.sql
var schemaFS embed.FS

// Migrate 读取并执行数据库迁移脚本
// 创建所有必需的数据库表
func Migrate(db *sql.DB) error {
	// 从嵌入的文件系统中读取 schema.sql
	schemaContent, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read schema.sql: %w", err)
	}

	// 执行 SQL 脚本
	_, err = db.Exec(string(schemaContent))
	if err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}

	return nil
}
