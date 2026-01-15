package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/biqiangwu/flowerSalesSystem/internal/config"
	"github.com/biqiangwu/flowerSalesSystem/internal/database"
)

//go:embed static
var staticFiles embed.FS

func main() {
	// 1. 加载配置
	cfg := config.Load()
	log.Printf("加载配置: DBHost=%s, DBName=%s", cfg.DBHost, cfg.DBName)

	// 2. 建立数据库连接
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()
	log.Println("数据库连接成功")

	// 3. 执行数据库迁移
	if err := database.Migrate(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")

	// 4. 创建 HTTP ServeMux
	mux := http.NewServeMux()

	// 5. 注册静态文件服务
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("静态文件系统配置失败: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// 6. 启动 HTTP 服务器
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("HTTP 服务器启动在 http://0.0.0.0%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("HTTP 服务器错误: %v", err)
	}
}

// init 用于日志初始化
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
