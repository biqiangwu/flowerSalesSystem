package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/biqiangwu/flowerSalesSystem/internal/address"
	"github.com/biqiangwu/flowerSalesSystem/internal/auth"
	"github.com/biqiangwu/flowerSalesSystem/internal/config"
	"github.com/biqiangwu/flowerSalesSystem/internal/database"
	"github.com/biqiangwu/flowerSalesSystem/internal/flower"
	"github.com/biqiangwu/flowerSalesSystem/internal/handler"
	"github.com/biqiangwu/flowerSalesSystem/internal/order"
	"github.com/biqiangwu/flowerSalesSystem/internal/user"
	"github.com/biqiangwu/flowerSalesSystem/pkg/middleware"
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

	// 4. 初始化 Repository 层
	userRepo := user.NewMySQLUserRepository(db)
	addressRepo := address.NewAddressRepository(db)
	flowerRepo := flower.NewFlowerRepository(db)
	orderRepo := order.NewOrderRepository(db)
	orderLogRepo := order.NewOrderLogRepository(db)

	// 5. 初始化 Session 管理
	sessionMgr := auth.NewMemorySessionManager()

	// 6. 初始化服务层
	authSvc := auth.NewAuthService(userRepo, sessionMgr)
	flowerSvc := flower.NewFlowerService(flowerRepo)
	addressSvc := address.NewAddressService(addressRepo)
	orderSvc := order.NewOrderService(orderRepo, flowerRepo, orderLogRepo)
	orderLogSvc := order.NewOrderLogService(orderLogRepo)
	userSvc := user.NewUserService(userRepo)

	// 7. 创建 Handler 并注入所有服务
	h := handler.NewHandler(authSvc, orderSvc)
	h.SetServices(authSvc, orderSvc, orderLogSvc, userSvc, flowerSvc, addressSvc, userRepo)

	// 8. 创建 HTTP ServeMux
	mux := http.NewServeMux()

	// 9. 注册所有 API 路由
	h.RegisterRoutes(mux)
	log.Println("API 路由注册完成")

	// 10. 注册静态文件服务（SPA 模式）
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("静态文件系统配置失败: %v", err)
	}

	// 创建 SPA 处理器：未匹配的路由返回 index.html
	spaHandler := &spaHandler{http.FileServer(http.FS(staticFS))}
	mux.Handle("/", spaHandler)

	// 11. 应用中间件
	// 包装日志中间件和恢复中间件
	finalHandler := middleware.LoggingMiddleware(middleware.RecoveryMiddleware(mux))

	// 12. 启动 HTTP 服务器
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("HTTP 服务器启动在 http://0.0.0.0%s", addr)

	if err := http.ListenAndServe(addr, finalHandler); err != nil {
		log.Fatalf("HTTP 服务器错误: %v", err)
	}
}

// spaHandler 处理 SPA 路由，未匹配的路由返回 index.html
type spaHandler struct {
	handler http.Handler
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 如果是 API 请求，不处理
	if strings.HasPrefix(r.URL.Path, "/api/") {
		h.handler.ServeHTTP(w, r)
		return
	}

	// 尝试提供静态文件
	h.handler.ServeHTTP(w, r)
}

// init 用于日志初始化
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
