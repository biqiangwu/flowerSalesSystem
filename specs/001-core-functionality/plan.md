# 鲜花销售系统技术实现方案

## 文档信息
- **版本**: 1.0
- **创建日期**: 2026-01-14
- **项目名称**: flowerSalesSystem
- **对应规格**: `001-core-functionality/spec.md`

---

## 1. 技术上下文总结

### 1.1 技术选型

| 组件 | 选型 | 理由 |
|------|------|------|
| **编程语言** | Go 1.25+ | 静态类型、并发原生支持、部署简单 |
| **Web框架** | `net/http` (标准库) | 遵循简单性原则，减少外部依赖 |
| **数据库** | MySQL 8.0+ | 事务支持成熟、社区活跃 |
| **数据库驱动** | `go-sql-driver/mysql` | 官方推荐，稳定可靠 |
| **密码加密** | `golang.org/x/crypto/bcrypt` | 标准库扩展，安全可靠 |
| **部署平台** | Kubernetes v1.34 | 容器编排，生产级部署 |
| **容器测试** | Kind v1.34 | 本地开发环境模拟 |

### 1.2 架构原则

1. **单体架构**: 单一 Go 二进制程序，静态资源通过 `embed.FS` 嵌入
2. **标准库优先**: 仅引入必要的依赖（bcrypt、MySQL驱动）
3. **显式依赖注入**: 通过结构体成员传递依赖，拒绝全局变量
4. **测试先行**: TDD 驱动开发，表格驱动测试
5. **简单性优先**: 不进行微服务拆分，前端和后端在同一进程

---

## 2. "合宪性"审查

### 2.1 简单性原则 (Simplicity First)

| 宪法条款 | 审查结果 | 说明 |
|----------|----------|------|
| 1.1 YAGNI | ✅ 符合 | 仅实现 spec.md 中明确要求的功能，不做过度设计 |
| 1.2 标准库优先 | ✅ 符合 | Web服务使用 `net/http`，无 Gin/Echo 等框架 |
| 1.3 反过度工程 | ✅ 符合 | 使用简单函数和结构体，避免不必要的接口抽象 |

### 2.2 测试先行铁律 (Test-First Imperative)

| 宪法条款 | 审查结果 | 说明 |
|----------|----------|------|
| 2.1 TDD循环 | ✅ 符合 | 所有功能开发遵循 Red-Green-Refactor |
| 2.2 表格驱动 | ✅ 符合 | 单元测试优先采用表格驱动测试风格 |
| 2.3 拒绝Mocks | ✅ 符合 | 优先编写集成测试，使用真实数据库 |

### 2.3 明确性原则 (Clarity and Explicitness)

| 宪法条款 | 审查结果 | 说明 |
|----------|----------|------|
| 3.1 错误处理 | ✅ 符合 | 所有错误显式处理，使用 `fmt.Errorf` 包装错误链 |
| 3.2 无全局变量 | ✅ 符合 | 所有依赖通过函数参数或结构体成员注入 |

---

## 3. 项目结构细化

### 3.1 目录结构

```
flowerSalesSystem/
├── cmd/
│   └── server/
│       └── main.go                 # 程序入口
├── internal/
│   ├── auth/                       # 认证模块
│   │   ├── auth.go                 # 认证逻辑
│   │   ├── auth_test.go            # 认证测试
│   │   └── session.go              # Session管理
│   ├── flower/                     # 鲜花模块
│   │   ├── flower.go               # 鲜花CRUD
│   │   ├── flower_test.go
│   │   └── repository.go           # 鲜花数据访问
│   ├── address/                    # 地址模块
│   │   ├── address.go
│   │   ├── address_test.go
│   │   └── repository.go
│   ├── order/                      # 订单模块
│   │   ├── order.go                # 订单CRUD
│   │   ├── order_test.go
│   │   ├── repository.go
│   │   └── service.go              # 订单业务逻辑（库存扣减/回退）
│   ├── user/                       # 用户模块
│   │   ├── user.go                 # 用户CRUD
│   │   ├── user_test.go
│   │   └── repository.go
│   ├── config/                     # 配置管理
│   │   └── config.go               # 从环境变量读取配置
│   ├── database/                   # 数据库
│   │   ├── database.go             # 数据库连接
│   │   └── migrate.go              # 数据库迁移
│   └── handler/                    # HTTP处理器
│       ├── handler.go              # 公共处理器工具
│       ├── auth_handler.go         # 认证相关
│       ├── flower_handler.go       # 鲜花相关
│       ├── address_handler.go      # 地址相关
│       ├── order_handler.go        # 订单相关
│       └── user_handler.go         # 用户管理
├── pkg/
│   └── middleware/
│       ├── auth.go                 # 认证中间件
│       ├── logging.go              # 日志中间件
│       └── recovery.go             # 恢复中间件
├── web/                            # 前端静态文件（编译时嵌入）
│   └── static/                     # HTML/CSS/JS
│       ├── index.html
│       ├── css/
│       └── js/
├── k8s/                            # Kubernetes配置
│   ├── kind-cluster.yaml
│   ├── metallb-native.yaml
│   └── ingress-nginx.yaml
├── helm/                           # Helm Chart
│   └── flower-sales-system/
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
├── go.mod
├── go.sum
├── Makefile
└── constitution.md
```

**关键设计**：
- 静态文件放在 `web/static/` 目录
- 使用 Go 1.16+ 的 `embed` 指令将静态文件嵌入二进制
- 编译后的单一二进制文件包含所有资源

### 3.2 模块依赖关系

```
                    ┌─────────────────────────┐
                    │   Go Server (单一进程)   │
                    │                         │
                    │  ┌────────────────────┐  │
                    │  │   handler/         │  │ ← HTTP处理层
                    │  │  ┌──────────────┐  │  │
                    │  │  │ /api/*       │  │  │ ← API路由
                    │  │  │ /*           │  │  │ ← 静态文件
                    │  │  └──────────────┘  │  │
                    │  └─────────┬──────────┘  │
                    └────────────┼─────────────┘
                                 │
        ┌────────────────────────┼────────────────────────┐
        │                        │                        │
   ┌────▼─────┐          ┌─────▼─────┐          ┌─────▼─────┐
   │  auth/   │          │  flower/  │          │  order/   │
   └──────────┘          └───────────┘          └─────┬─────┘
                                                       │
                                                  ┌────▼─────┐
                                                  │   user/  │
                                                  └──────────┘
        ┌──────────────────────────┴─────────────────────────┐
        │                    database/                       │
        └──────────────────────────┬────────────────────────┘
                                   │
                           ┌───────▼────────┐
                           │     MySQL      │
                           └────────────────┘
```

**架构说明**：
- 单一 Go 进程同时处理 API 请求和静态文件服务
- 静态文件通过 `embed.FS` 嵌入二进制，无运行时文件依赖
- 符合"简单性原则"，减少部署和运维复杂度

### 3.3 包职责说明

| 包 | 职责 | 对外暴露 |
|---|------|----------|
| `internal/auth` | 用户认证、Session管理 | `Login()`, `Logout()`, `Register()` |
| `internal/flower` | 鲜花CRUD、库存管理 | `FlowerService` interface |
| `internal/address` | 收货地址管理 | `AddressService` interface |
| `internal/order` | 订单CRUD、状态流转、库存扣减 | `OrderService` interface |
| `internal/user` | 用户管理（含密码重置） | `UserService` interface |
| `internal/config` | 配置加载 | `Load() *Config` |
| `internal/database` | 数据库连接、迁移 | `Open() *sql.DB`, `Migrate()` |
| `internal/handler` | HTTP请求处理 | `RegisterRoutes()` |
| `pkg/middleware` | HTTP中间件 | `AuthMiddleware()`, `LoggingMiddleware()` |

---

## 4. 核心数据结构

### 4.1 领域模型 (Domain Models)

#### User (用户)
```go
// internal/user/user.go
type Role string

const (
    RoleCustomer Role = "customer"
    RoleClerk    Role = "clerk"
    RoleAdmin    Role = "admin"
)

type User struct {
    ID           int       `json:"id"`
    Username     string    `json:"username"`
    PasswordHash string    `json:"-"` // 不序列化
    Role         Role      `json:"role"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

#### Flower (鲜花)
```go
// internal/flower/flower.go
import "encoding/json"

type Decimal struct {
    value int64 // 存储为分（整数）
}

type Flower struct {
    SKU          string    `json:"sku"`
    Name         string    `json:"name"`
    Origin       string    `json:"origin"`
    ShelfLife    string    `json:"shelf_life"`
    Preservation string    `json:"preservation"`
    PurchasePrice Decimal  `json:"purchase_price"`
    SalePrice     Decimal  `json:"sale_price"`
    Stock         int      `json:"stock"`
    IsActive      bool     `json:"is_active"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

#### Address (收货地址)
```go
// internal/address/address.go
type Address struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id"`
    Label     string    `json:"label"`      // 可选
    Address   string    `json:"address"`
    Contact   string    `json:"contact"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### Order (订单)
```go
// internal/order/order.go
type OrderStatus string

const (
    StatusPending   OrderStatus = "pending"
    StatusCompleted OrderStatus = "completed"
    StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
    ID          int         `json:"id"`
    OrderNo     string      `json:"order_no"`
    UserID      int         `json:"user_id"`
    AddressID   int         `json:"address_id"`
    TotalAmount Decimal     `json:"total_amount"`
    Status      OrderStatus `json:"status"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
}

type OrderItem struct {
    ID          int     `json:"id"`
    OrderID     int     `json:"order_id"`
    FlowerSKU   string  `json:"flower_sku"`
    FlowerName  string  `json:"flower_name"`
    Quantity    int     `json:"quantity"`
    UnitPrice   Decimal `json:"unit_price"`
    Subtotal    Decimal `json:"subtotal"`
}
```

#### OrderLog (订单操作日志)
```go
// internal/order/log.go
type OrderLog struct {
    ID         int       `json:"id"`
    OrderID    int       `json:"order_id"`
    OperatorID int       `json:"operator_id"`
    Action     string    `json:"action"`
    OldStatus  string    `json:"old_status"`
    NewStatus  string    `json:"new_status"`
    CreatedAt  time.Time `json:"created_at"`
}
```

### 4.2 DTO (数据传输对象)

#### 请求DTO
```go
// internal/handler/dto.go
type RegisterRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type CreateFlowerRequest struct {
    SKU          string  `json:"sku"`
    Name         string  `json:"name"`
    Origin       string  `json:"origin"`
    ShelfLife    string  `json:"shelf_life"`
    Preservation string  `json:"preservation"`
    PurchasePrice float64 `json:"purchase_price"` // 元
    SalePrice     float64 `json:"sale_price"`     // 元
    Stock         int     `json:"stock"`
}

type CreateOrderRequest struct {
    AddressID int                      `json:"address_id"`
    Items     []CreateOrderItemRequest `json:"items"`
}

type CreateOrderItemRequest struct {
    FlowerSKU string `json:"flower_sku"`
    Quantity  int    `json:"quantity"`
}

type ResetPasswordRequest struct {
    NewPassword string `json:"new_password"`
}
```

#### 响应DTO
```go
type UserResponse struct {
    ID        int       `json:"id"`
    Username  string    `json:"username"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
}

type FlowerResponse struct {
    SKU          string  `json:"sku"`
    Name         string  `json:"name"`
    Origin       string  `json:"origin"`
    ShelfLife    string  `json:"shelf_life"`
    Preservation string  `json:"preservation"`
    PurchasePrice float64 `json:"purchase_price"`
    SalePrice     float64 `json:"sale_price"`
    Stock         int     `json:"stock"`
    IsActive      bool    `json:"is_active"`
    LowStock      bool    `json:"low_stock"` // 库存预警标识
}

type OrderResponse struct {
    ID          int       `json:"id"`
    OrderNo     string    `json:"order_no"`
    UserID      int       `json:"user_id"`
    Address     Address   `json:"address"`
    TotalAmount float64   `json:"total_amount"`
    Status      string    `json:"status"`
    Items       []OrderItem `json:"items"`
    CreatedAt   time.Time `json:"created_at"`
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Details string `json:"details,omitempty"`
}
```

---

## 5. 接口设计

### 5.1 Repository 接口 (数据访问层)

#### FlowerRepository
```go
// internal/flower/repository.go
type FlowerRepository interface {
    Create(ctx context.Context, f *Flower) error
    GetBySKU(ctx context.Context, sku string) (*Flower, error)
    List(ctx context.Context, filter FlowerFilter) ([]*Flower, error)
    Update(ctx context.Context, f *Flower) error
    Delete(ctx context.Context, sku string) error
    UpdateStock(ctx context.Context, sku string, delta int) error
}

type FlowerFilter struct {
    Search       string   // 按 sku/name 搜索
    Origin       string   // 按产地筛选
    MinPrice     float64  // 最低价
    MaxPrice     float64  // 最高价
    SortBy       string   // price_asc, price_desc, stock
    Page         int
    PageSize     int
}
```

#### OrderRepository
```go
// internal/order/repository.go
type OrderRepository interface {
    Create(ctx context.Context, o *Order, items []*OrderItem) error
    GetByID(ctx context.Context, id int) (*Order, []*OrderItem, error)
    GetByOrderNo(ctx context.Context, orderNo string) (*Order, []*OrderItem, error)
    List(ctx context.Context, filter OrderFilter) ([]*Order, error)
    UpdateStatus(ctx context.Context, id int, status OrderStatus) error
    CreateLog(ctx context.Context, log *OrderLog) error
    GetLogs(ctx context.Context, orderID int) ([]*OrderLog, error)
}

type OrderFilter struct {
    UserID     int
    Status     string
    OrderNo    string
    Page       int
    PageSize   int
}
```

#### UserRepository
```go
// internal/user/repository.go
type UserRepository interface {
    Create(ctx context.Context, u *User) error
    GetByID(ctx context.Context, id int) (*User, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
    List(ctx context.Context, page, pageSize int) ([]*User, error)
    Delete(ctx context.Context, id int) error
    UpdatePassword(ctx context.Context, id int, passwordHash string) error
}
```

#### AddressRepository
```go
// internal/address/repository.go
type AddressRepository interface {
    Create(ctx context.Context, a *Address) error
    GetByID(ctx context.Context, id int) (*Address, error)
    ListByUserID(ctx context.Context, userID int) ([]*Address, error)
    Update(ctx context.Context, a *Address) error
    Delete(ctx context.Context, id int) error
}
```

### 5.2 Service 接口 (业务逻辑层)

#### AuthService
```go
// internal/auth/auth.go
type AuthService interface {
    Register(ctx context.Context, username, password string) (*User, error)
    Login(ctx context.Context, username, password string) (*Session, error)
    Logout(ctx context.Context, sessionToken string) error
    ValidateSession(ctx context.Context, sessionToken string) (*User, error)
    HashPassword(password string) (string, error)
    VerifyPassword(password, hash string) bool
}

type Session struct {
    Token     string
    UserID    int
    Username  string
    Role      Role
    ExpiresAt time.Time
}
```

#### OrderService
```go
// internal/order/service.go
type OrderService interface {
    CreateOrder(ctx context.Context, userID int, req *CreateOrderRequest) (*Order, error)
    CompleteOrder(ctx context.Context, orderID int, operatorID int) error
    CancelOrder(ctx context.Context, orderID int, operatorID int) error
    GetOrder(ctx context.Context, orderID int) (*Order, []*OrderItem, error)
    ListOrders(ctx context.Context, filter OrderFilter) ([]*Order, error)
}
```

#### FlowerService
```go
// internal/flower/flower.go
type FlowerService interface {
    CreateFlower(ctx context.Context, req *CreateFlowerRequest) error
    GetFlower(ctx context.Context, sku string) (*Flower, error)
    ListFlowers(ctx context.Context, filter FlowerFilter) ([]*Flower, error)
    UpdateFlower(ctx context.Context, sku string, req *UpdateFlowerRequest) error
    DeleteFlower(ctx context.Context, sku string) error
    AddStock(ctx context.Context, sku string, quantity int) error
}
```

#### UserService
```go
// internal/user/user.go
type UserService interface {
    ListUsers(ctx context.Context, page, pageSize int) ([]*User, error)
    DeleteUser(ctx context.Context, id int) error
    ResetPassword(ctx context.Context, id int, newPassword string) error
}
```

### 5.3 Handler 接口 (HTTP处理层)

#### 路由注册
```go
// internal/handler/handler.go
type Handler struct {
    authService   auth.AuthService
    flowerService flower.FlowerService
    orderService  order.OrderService
    userService   user.UserService
    // ...
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    // 认证相关
    mux.HandleFunc("POST /api/register", h.HandleRegister)
    mux.HandleFunc("POST /api/login", h.HandleLogin)
    mux.HandleFunc("POST /api/logout", h.HandleLogout)

    // 鲜花相关
    mux.HandleFunc("GET /api/flowers", h.HandleListFlowers)
    mux.HandleFunc("GET /api/flowers/", h.HandleGetFlower)
    mux.HandleFunc("POST /api/flowers", h.AuthMiddleware(h.RoleMiddleware(RoleClerk, RoleAdmin), h.HandleCreateFlower))
    // ... 其他路由
}
```

---

## 6. 关键实现细节

### 6.1 库存扣减与回退 (事务保证)

```go
// internal/order/service.go
func (s *OrderService) CreateOrder(ctx context.Context, userID int, req *CreateOrderRequest) (*Order, error) {
    // 1. 开启事务
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback() // 如果未Commit，自动Rollback

    // 2. 检查库存并锁定
    for _, item := range req.Items {
        var stock int
        err := tx.QueryRowContext(ctx,
            "SELECT stock FROM flowers WHERE sku = ? FOR UPDATE", item.FlowerSKU).Scan(&stock)
        if err != nil {
            return nil, fmt.Errorf("check stock for %s: %w", item.FlowerSKU, err)
        }
        if stock < item.Quantity {
            return nil, fmt.Errorf("insufficient stock for %s: have %d, need %d",
                item.FlowerSKU, stock, item.Quantity)
        }
    }

    // 3. 扣减库存
    for _, item := range req.Items {
        _, err := tx.ExecContext(ctx,
            "UPDATE flowers SET stock = stock - ? WHERE sku = ?", item.Quantity, item.FlowerSKU)
        if err != nil {
            return nil, fmt.Errorf("deduct stock for %s: %w", item.FlowerSKU, err)
        }
    }

    // 4. 创建订单和订单项
    orderNo := generateOrderNo()
    order := &Order{
        OrderNo:   orderNo,
        UserID:    userID,
        AddressID: req.AddressID,
        Status:    StatusPending,
    }
    // ... 插入订单和订单项

    // 5. 提交事务
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit transaction: %w", err)
    }

    return order, nil
}
```

### 6.2 订单编号生成

```go
// internal/order/order.go
func generateOrderNo() string {
    // 格式: ORD + YYYYMMDD + 6位随机数
    date := time.Now().Format("20060102")
    random := rand.Intn(999999)
    return fmt.Sprintf("ORD%s%06d", date, random)
}
```

### 6.3 认证中间件

```go
// pkg/middleware/auth.go
func AuthMiddleware(authService auth.AuthService, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_token")
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        user, err := authService.ValidateSession(r.Context(), cookie.Value)
        if err != nil {
            http.Error(w, "invalid session", http.StatusUnauthorized)
            return
        }

        // 将用户信息存入请求上下文
        ctx := context.WithValue(r.Context(), "user", user)
        next(w, r.WithContext(ctx))
    }
}
```

### 6.4 配置加载

```go
// internal/config/config.go
type Config struct {
    // Database
    DBHost     string
    DBPort     int
    DBName     string
    DBUser     string
    DBPassword string

    // JWT/Session
    SessionSecret string
    SessionExpiry int // hours

    // Server
    ServerPort int
    LogLevel   string

    // Business
    StockWarningThreshold int
}

func Load() *Config {
    return &Config{
        DBHost:               getEnv("DB_HOST", "mysql-service"),
        DBPort:               getEnvInt("DB_PORT", 3306),
        DBName:               getEnv("DB_NAME", "flower_sales"),
        DBUser:               getEnv("DB_USER", "flower_user"),
        DBPassword:           getEnv("DB_PASSWORD", ""),
        SessionSecret:        getEnv("SESSION_SECRET", ""),
        SessionExpiry:        getEnvInt("SESSION_EXPIRY", 24),
        ServerPort:           getEnvInt("SERVER_PORT", 8080),
        LogLevel:             getEnv("LOG_LEVEL", "info"),
        StockWarningThreshold: getEnvInt("STOCK_WARNING_THRESHOLD", 10),
    }
}
```

### 6.5 静态文件嵌入

使用 Go 1.16+ 的 `embed` 指令将静态文件嵌入二进制：

```go
// cmd/server/main.go
package main

import (
    "embed"
    "io/fs"
    "net/http"

    "github.com/biqiangwu/flower-sales-system/internal/config"
    "github.com/biqiangwu/flower-sales-system/internal/handler"
    "github.com/biqiangwu/flower-sales-system/pkg/middleware"
)

//go:embed web/static/*
var staticFiles embed.FS

func main() {
    cfg := config.Load()

    mux := http.NewServeMux()

    // 注册 API 路由
    h := handler.New(cfg)
    h.RegisterRoutes(mux)

    // 静态文件服务 (处理 /api/* 之外的所有请求)
    staticFS, _ := fs.Sub(staticFiles, "web/static")
    mux.Handle("/", http.FileServer(http.FS(staticFS)))

    // 应用中间件
    handler := middleware.LoggingMiddleware(
        middleware.RecoveryMiddleware(mux),
    )

    addr := fmt.Sprintf(":%d", cfg.ServerPort)
    http.ListenAndServe(addr, handler)
}
```

**路由规则**：
- `/api/*` → API Handler（返回 JSON）
- `/*` → 静态文件服务器（返回 HTML/CSS/JS）
- SPA 应用：未匹配的路由返回 `index.html`

---

## 7. 测试策略

### 7.1 表格驱动测试示例

```go
// internal/auth/auth_test.go
func TestHashPassword(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid password", "password123", false},
        {"empty password", "", true},
        {"short password", "123", true}, // 最小长度校验
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            hash, err := HashPassword(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && !VerifyPassword(tt.input, hash) {
                t.Error("VerifyPassword() failed")
            }
        })
    }
}
```

### 7.2 集成测试 (使用真实数据库)

```go
// internal/order/service_integration_test.go
func TestOrderService_CreateOrder_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    db := setupTestDB(t)
    defer db.Close()

    // 插入测试数据
    // ...

    svc := NewOrderService(db)

    // 执行测试
    // ...
}
```

---

## 8. 部署相关配置

### 8.1 Dockerfile

```dockerfile
# cmd/server/Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# 构建时将静态文件嵌入二进制
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/server .

# 单一端口同时处理 API 和静态文件
EXPOSE 8080
CMD ["./server"]
```

**说明**：
- 构建时 `go build` 会通过 `//go:embed` 将静态文件嵌入二进制
- 运行时无需额外的静态文件目录
- 单一容器、单一端口，简化部署

### 8.2 Makefile 目标

```makefile
.PHONY: test build run docker-build k8s-setup k8s-deploy k8s-clean

test:
    go test -v ./...

build:
    go build -o bin/server ./cmd/server

run:
    go run ./cmd/server

docker-build:
    docker build -t flower-sales-system:latest -f cmd/server/Dockerfile .

k8s-setup:
    kind create cluster --config=k8s/kind-cluster.yaml
    kubectl apply -f k8s/metallb-native.yaml
    kubectl apply -f k8s/ingress-nginx.yaml

k8s-deploy: docker-build
    kind load docker-image flower-sales-system:latest --name flower-sales-system
    helm install flower-sales-system helm/flower-sales-system --namespace flower-sales --create-namespace
    kubectl -n flower-sales rollout status deployment

k8s-clean:
    helm uninstall flower-sales-system -n flower-sales || true
    kind delete cluster --name flower-sales-system
```

---

## 9. 开发里程碑

| 阶段 | 任务 | 交付物 |
|------|------|--------|
| **Phase 1** | 基础框架 | 项目结构、数据库连接、配置加载 |
| **Phase 2** | 用户认证 | 注册、登录、Session、中间件 |
| **Phase 3** | 鲜花管理 | CRUD、列表搜索筛选分页 |
| **Phase 4** | 地址管理 | CRUD（顾客） |
| **Phase 5** | 订单核心 | 创建订单、库存扣减（事务） |
| **Phase 6** | 订单流转 | 完成、取消、库存回退 |
| **Phase 7** | 库存管理 | 进货入库、盘点调整、预警 |
| **Phase 8** | 用户管理 | 用户列表、删除、密码重置 |
| **Phase 9** | 订单日志 | 操作记录、查询 |
| **Phase 10** | 部署配置 | K8s、Helm、CI/CD |

---

## 附录

### A. 依赖清单 (go.mod)

```
module github.com/biqiangwu/flower-sales-system

go 1.25

require (
    github.com/go-sql-driver/mysql v1.8.1
    golang.org/x/crypto v0.32.0
)
```

**说明**：
- `embed` 是 Go 标准库（Go 1.16+），无需额外依赖
- 静态文件嵌入不增加任何外部依赖

### B. 数据库迁移脚本

```sql
-- internal/database/schema.sql
-- 见 spec.md 第4节
```

### C. API 响应码规范

| HTTP状态码 | 含义 | 场景 |
|-----------|------|------|
| 200 | OK | 查询成功 |
| 201 | Created | 创建成功 |
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | 未登录/Session无效 |
| 403 | Forbidden | 权限不足 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 资源冲突（如用户名已存在） |
| 500 | Internal Server Error | 服务器错误 |
