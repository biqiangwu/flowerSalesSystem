# 鲜花销售系统开发任务列表

## 文档信息
- **版本**: 1.0
- **创建日期**: 2026-01-14
- **项目名称**: flowerSalesSystem
- **对应规格**: `001-core-functionality/spec.md`
- **对应方案**: `001-core-functionality/plan.md`

---

## 任务说明

### 符号约定
- `[P]` - 可并行执行的任务（无依赖关系）
- `→` - 任务依赖关系（左侧任务依赖于右侧任务）

### TDD 铁律
根据 `constitution.md` 第二条"测试先行铁律"：
- **所有实现任务前必须先有对应的测试任务**
- **严格遵循 Red-Green-Refactor 循环**
- **单元测试优先采用表格驱动测试风格**

---

## Phase 1: 基础框架

### 1.1 项目初始化
- [ ] **任务 1.1**: 创建项目目录结构
  - 创建 `internal/` 及其子目录：`auth/`, `flower/`, `address/`, `order/`, `user/`, `config/`, `database/`, `handler/`
  - 创建 `pkg/middleware/` 目录
  - 创建 `cmd/server/` 目录
  - 创建 `web/static/` 目录

- [ ] **任务 1.2**: 初始化 go.mod 文件
  - 创建 `go.mod`
  - 设置模块路径 `github.com/biqiangwu/flowerSalesSystem`
  - 设置 Go 版本 `go 1.25`
  - 添加依赖：`go-sql-driver/mysql`、`golang.org/x/crypto`

### 1.2 配置管理模块
- [ ] **任务 1.3** [P]: 创建配置管理模块测试 `internal/config/config_test.go`
  - 表格驱动测试：测试环境变量读取、默认值、类型转换
  - Mock 环境变量场景

- [ ] **任务 1.4**: 依赖任务 1.3 - 创建配置管理模块 `internal/config/config.go`
  - 定义 `Config` 结构体
  - 实现 `Load() *Config` 函数
  - 实现 `getEnv()`, `getEnvInt()` 辅助函数

### 1.3 数据库模块
- [ ] **任务 1.5** [P]: 创建数据库连接测试 `internal/database/database_test.go`
  - 测试连接建立、连接池配置、错误处理

- [ ] **任务 1.6**: 依赖任务 1.5 - 创建数据库连接模块 `internal/database/database.go`
  - 定义 `DBConfig` 结构体
  - 实现 `Open(cfg *config.Config) (*sql.DB, error)` 函数
  - 实现连接池配置

- [ ] **任务 1.7**: 创建数据库迁移脚本 `internal/database/schema.sql`
  - 创建 users 表
  - 创建 addresses 表
  - 创建 flowers 表
  - 创建 orders 表
  - 创建 order_items 表
  - 创建 order_logs 表

- [ ] **任务 1.8**: 依赖任务 1.6 - 创建数据库迁移模块 `internal/database/migrate.go`
  - 实现 `Migrate(db *sql.DB) error` 函数
  - 读取并执行 schema.sql

### 1.4 程序入口
- [ ] **任务 1.9**: 创建程序入口文件 `cmd/server/main.go`
  - 加载配置
  - 建立数据库连接
  - 执行数据库迁移
  - 创建 HTTP ServeMux
  - 注册静态文件服务（使用 embed）
  - 启动 HTTP 服务器

---

## Phase 2: 用户认证模块

### 2.1 用户模型和 Repository
- [ ] **任务 2.1** [P]: 创建用户模型测试 `internal/user/user_test.go`
  - 测试用户结构体创建、字段验证

- [ ] **任务 2.2**: 依赖任务 2.1 - 创建用户模型 `internal/user/user.go`
  - 定义 `Role` 类型及其常量
  - 定义 `User` 结构体

- [ ] **任务 2.3** [P]: 创建用户 Repository 测试 `internal/user/repository_test.go`
  - 表格驱动测试：Create, GetByID, GetByUsername, List, Delete, UpdatePassword
  - 集成测试使用真实数据库

- [ ] **任务 2.4**: 依赖任务 2.3 - 创建用户 Repository `internal/user/repository.go`
  - 定义 `UserRepository` interface
  - 实现 `MySQLUserRepository` 结构体
  - 实现 Create, GetByID, GetByUsername, List, Delete, UpdatePassword 方法

### 2.2 认证服务
- [ ] **任务 2.5** [P]: 创建认证服务测试 `internal/auth/auth_test.go`
  - 测试密码哈希和验证
  - 测试注册、登录、登出、Session验证
  - 测试用户名已存在场景
  - 测试密码错误场景

- [ ] **任务 2.6**: 依赖任务 2.4, 2.5 - 创建认证服务 `internal/auth/auth.go`
  - 定义 `AuthService` interface
  - 定义 `Session` 结构体
  - 实现 `AuthService` 的实现
  - 实现密码哈希和验证（使用 bcrypt）
  - 实现 Register, Login, Logout, ValidateSession 方法

### 2.3 Session 管理
- [ ] **任务 2.7** [P]: 创建 Session 管理测试 `internal/auth/session_test.go`
  - 测试 Session 创建、验证、过期、删除

- [ ] **任务 2.8**: 依赖任务 2.7 - 创建 Session 管理 `internal/auth/session.go`
  - 实现 Session 存储结构（内存或数据库）
  - 实现 CreateSession, ValidateSession, DeleteSession 方法

### 2.4 认证 Handler
- [ ] **任务 2.9** [P]: 创建认证 Handler 测试 `internal/handler/auth_handler_test.go`
  - 测试注册接口
  - 测试登录接口
  - 测试登出接口
  - 测试错误场景

- [ ] **任务 2.10**: 依赖任务 2.6, 2.9 - 创建认证 Handler `internal/handler/auth_handler.go`
  - 创建 `RegisterRequest`, `LoginRequest` DTO
  - 实现 HandleRegister 函数
  - 实现 HandleLogin 函数
  - 实现 HandleLogout 函数

---

## Phase 3: 鲜花管理模块

### 3.1 鲜花模型
- [ ] **任务 3.1** [P]: 创建 Decimal 类型测试 `internal/flower/decimal_test.go`
  - 测试 Decimal 创建、加减乘除、比较

- [ ] **任务 3.2**: 依赖任务 3.1 - 创建 Decimal 类型 `internal/flower/decimal.go`
  - 定义 `Decimal` 结构体
  - 实现从 float64 转换、精度处理

- [ ] **任务 3.3** [P]: 创建鲜花模型测试 `internal/flower/flower_test.go`
  - 测试鲜花结构体创建、字段验证

- [ ] **任务 3.4**: 依赖任务 3.2, 3.3 - 创建鲜花模型 `internal/flower/flower.go`
  - 定义 `Flower` 结构体
  - 定义 `FlowerFilter` 结构体

### 3.2 鲜花 Repository
- [ ] **任务 3.5** [P]: 创建鲜花 Repository 测试 `internal/flower/repository_test.go`
  - 表格驱动测试：Create, GetBySKU, List, Update, Delete, UpdateStock
  - 测试搜索、筛选、排序、分页
  - 集成测试使用真实数据库

- [ ] **任务 3.6**: 依赖任务 3.5 - 创建鲜花 Repository `internal/flower/repository.go`
  - 定义 `FlowerRepository` interface
  - 实现 `MySQLFlowerRepository` 结构体
  - 实现 Create, GetBySKU, List, Update, Delete, UpdateStock 方法

### 3.3 鲜花服务
- [ ] **任务 3.7** [P]: 创建鲜花服务测试 `internal/flower/service_test.go`
  - 测试创建鲜花、获取鲜花、列表查询、更新鲜花、删除鲜花
  - 测试库存操作（进货入库、盘点调整）
  - 测试库存预警逻辑

- [ ] **任务 3.8**: 依赖任务 3.6, 3.7 - 创建鲜花服务 `internal/flower/service.go`
  - 定义 `FlowerService` interface
  - 实现 `FlowerService` 的实现
  - 实现 CreateFlower, GetFlower, ListFlowers, UpdateFlower, DeleteFlower, AddStock 方法

### 3.4 鲜花 Handler
- [ ] **任务 3.9** [P]: 创建鲜花 Handler 测试 `internal/handler/flower_handler_test.go`
  - 测试获取鲜花列表
  - 测试获取鲜花详情
  - 测试创建鲜花
  - 测试更新鲜花
  - 测试删除鲜花
  - 测试进货入库

- [ ] **任务 3.10**: 依赖任务 3.8, 3.9 - 创建鲜花 Handler `internal/handler/flower_handler.go`
  - 创建 `CreateFlowerRequest`, `UpdateFlowerRequest` DTO
  - 实现 HandleListFlowers 函数
  - 实现 HandleGetFlower 函数
  - 实现 HandleCreateFlower 函数
  - 实现 HandleUpdateFlower 函数
  - 实现 HandleDeleteFlower 函数
  - 实现 HandleAddStock 函数

---

## Phase 4: 地址管理模块

### 4.1 地址模型和 Repository
- [ ] **任务 4.1** [P]: 创建地址模型测试 `internal/address/address_test.go`
  - 测试地址结构体创建、字段验证

- [ ] **任务 4.2**: 依赖任务 4.1 - 创建地址模型 `internal/address/address.go`
  - 定义 `Address` 结构体

- [ ] **任务 4.3** [P]: 创建地址 Repository 测试 `internal/address/repository_test.go`
  - 表格驱动测试：Create, GetByID, ListByUserID, Update, Delete
  - 集成测试使用真实数据库

- [ ] **任务 4.4**: 依赖任务 4.3 - 创建地址 Repository `internal/address/repository.go`
  - 定义 `AddressRepository` interface
  - 实现 `MySQLAddressRepository` 结构体
  - 实现 Create, GetByID, ListByUserID, Update, Delete 方法

### 4.2 地址服务
- [ ] **任务 4.5** [P]: 创建地址服务测试 `internal/address/service_test.go`
  - 测试创建地址、获取地址、列表查询、更新地址、删除地址
  - 测试用户只能操作自己的地址

- [ ] **任务 4.6**: 依赖任务 4.4, 4.5 - 创建地址服务 `internal/address/service.go`
  - 定义 `AddressService` interface
  - 实现 `AddressService` 的实现
  - 实现 CreateAddress, GetAddress, ListAddresses, UpdateAddress, DeleteAddress 方法

### 4.3 地址 Handler
- [ ] **任务 4.7** [P]: 创建地址 Handler 测试 `internal/handler/address_handler_test.go`
  - 测试获取地址列表
  - 测试创建地址
  - 测试更新地址
  - 测试删除地址

- [ ] **任务 4.8**: 依赖任务 4.6, 4.7 - 创建地址 Handler `internal/handler/address_handler.go`
  - 创建 `CreateAddressRequest`, `UpdateAddressRequest` DTO
  - 实现 HandleListAddresses 函数
  - 实现 HandleCreateAddress 函数
  - 实现 HandleUpdateAddress 函数
  - 实现 HandleDeleteAddress 函数

---

## Phase 5: 订单核心模块

### 5.1 订单模型
- [ ] **任务 5.1** [P]: 创建订单模型测试 `internal/order/order_test.go`
  - 测试订单结构体创建、字段验证
  - 测试订单项结构体

- [ ] **任务 5.2**: 依赖任务 5.1 - 创建订单模型 `internal/order/order.go`
  - 定义 `OrderStatus` 类型及其常量
  - 定义 `Order` 结构体
  - 定义 `OrderItem` 结构体
  - 定义 `OrderFilter` 结构体
  - 实现 `generateOrderNo()` 函数

### 5.2 订单日志模型和 Repository
- [ ] **任务 5.3** [P]: 创建订单日志模型测试 `internal/order/log_test.go`
  - 测试订单日志结构体

- [ ] **任务 5.4**: 依赖任务 5.3 - 创建订单日志模型 `internal/order/log.go`
  - 定义 `OrderLog` 结构体

- [ ] **任务 5.5** [P]: 创建订单日志 Repository 测试 `internal/order/log_repository_test.go`
  - 测试 CreateLog, GetLogs

- [ ] **任务 5.6**: 依赖任务 5.5 - 创建订单日志 Repository `internal/order/log_repository.go`
  - 定义 `OrderLogRepository` interface
  - 实现 `MySQLOrderLogRepository` 结构体
  - 实现 CreateLog, GetLogs 方法

### 5.3 订单 Repository
- [ ] **任务 5.7** [P]: 创建订单 Repository 测试 `internal/order/repository_test.go`
  - 表格驱动测试：Create, GetByID, GetByOrderNo, List, UpdateStatus
  - 集成测试使用真实数据库

- [ ] **任务 5.8**: 依赖任务 5.6, 5.7 - 创建订单 Repository `internal/order/repository.go`
  - 定义 `OrderRepository` interface
  - 实现 `MySQLOrderRepository` 结构体
  - 实现 Create, GetByID, GetByOrderNo, List, UpdateStatus 方法

### 5.4 订单服务
- [ ] **任务 5.9** [P]: 创建订单服务测试 `internal/order/service_test.go`
  - 测试创建订单（含库存扣减事务）
  - 测试库存不足场景
  - 测试获取订单、列表查询
  - 集成测试验证事务完整性

- [ ] **任务 5.10**: 依赖任务 5.8, 5.9 - 创建订单服务 `internal/order/service.go`
  - 定义 `OrderService` interface
  - 定义 `CreateOrderRequest`, `CreateOrderItemRequest` DTO
  - 实现 `OrderService` 的实现
  - 实现 CreateOrder 方法（含库存扣减事务）
  - 实现 GetOrder, ListOrders 方法

### 5.5 订单 Handler
- [ ] **任务 5.11** [P]: 创建订单 Handler 测试 `internal/handler/order_handler_test.go`
  - 测试获取订单列表
  - 测试获取订单详情
  - 测试创建订单

- [ ] **任务 5.12**: 依赖任务 5.10, 5.11 - 创建订单 Handler `internal/handler/order_handler.go`
  - 实现 HandleListOrders 函数
  - 实现 HandleGetOrder 函数
  - 实现 HandleCreateOrder 函数

---

## Phase 6: 订单流转模块

### 6.1 订单状态流转
- [ ] **任务 6.1** [P]: 创建订单流转服务测试 `internal/order/service_transition_test.go`
  - 测试完成订单
  - 测试取消订单（含库存回退）
  - 测试状态流转权限
  - 集成测试验证库存回退事务

- [ ] **任务 6.2**: 依赖任务 6.1 - 扩展订单服务 `internal/order/service.go`
  - 实现 CompleteOrder 方法
  - 实现 CancelOrder 方法（含库存回退事务）

### 6.2 订单流转 Handler
- [ ] **任务 6.3** [P]: 创建订单流转 Handler 测试 `internal/handler/order_handler_transition_test.go`
  - 测试完成订单
  - 测试取消订单

- [ ] **任务 6.4**: 依赖任务 6.2, 6.3 - 扩展订单 Handler `internal/handler/order_handler.go`
  - 实现 HandleCompleteOrder 函数
  - 实现 HandleCancelOrder 函数

---

## Phase 7: 用户管理模块

### 7.1 用户管理服务
- [ ] **任务 7.1** [P]: 创建用户管理服务测试 `internal/user/service_test.go`
  - 测试获取用户列表
  - 测试删除用户
  - 测试重置用户密码
  - 测试权限验证

- [ ] **任务 7.2**: 依赖任务 2.4, 7.1 - 创建用户管理服务 `internal/user/service.go`
  - 定义 `UserService` interface
  - 实现 `UserService` 的实现
  - 实现 ListUsers, DeleteUser, ResetPassword 方法

### 7.2 用户管理 Handler
- [ ] **任务 7.3** [P]: 创建用户管理 Handler 测试 `internal/handler/user_handler_test.go`
  - 测试获取用户列表
  - 测试删除用户
  - 测试重置用户密码
  - 测试权限验证

- [ ] **任务 7.4**: 依赖任务 7.2, 7.3 - 创建用户管理 Handler `internal/handler/user_handler.go`
  - 创建 `ResetPasswordRequest` DTO
  - 实现 HandleListUsers 函数
  - 实现 HandleDeleteUser 函数
  - 实现 HandleResetPassword 函数

---

## Phase 8: 订单日志模块

### 8.1 订单日志服务
- [ ] **任务 8.1** [P]: 创建订单日志服务测试 `internal/order/log_service_test.go`
  - 测试记录订单状态变更
  - 测试查询订单日志

- [ ] **任务 8.2**: 依赖任务 5.6, 8.1 - 创建订单日志服务 `internal/order/log_service.go`
  - 定义 `OrderLogService` interface
  - 实现 `OrderLogService` 的实现
  - 实现 LogOrderAction, GetOrderLogs 方法

### 8.2 订单日志 Handler
- [ ] **任务 8.3** [P]: 创建订单日志 Handler 测试 `internal/handler/order_log_handler_test.go`
  - 测试获取订单操作日志

- [ ] **任务 8.4**: 依赖任务 8.2, 8.3 - 创建订单日志 Handler `internal/handler/order_log_handler.go`
  - 实现 HandleGetOrderLogs 函数

### 8.5 集成订单日志到流转服务
- [ ] **任务 8.5**: 扩展订单服务 `internal/order/service.go`
  - 在 CompleteOrder 方法中添加日志记录
  - 在 CancelOrder 方法中添加日志记录

---

## Phase 9: 中间件模块

### 9.1 认证中间件
- [ ] **任务 9.1** [P]: 创建认证中间件测试 `pkg/middleware/auth_test.go`
  - 测试有效 Session 通过
  - 测试无效/过期 Session 拒绝
  - 测试用户信息注入上下文

- [ ] **任务 9.2**: 依赖任务 2.6, 9.1 - 创建认证中间件 `pkg/middleware/auth.go`
  - 实现 AuthMiddleware 函数
  - 从 Cookie 读取 session_token
  - 验证 Session 并注入用户信息到上下文

### 9.2 角色中间件
- [ ] **任务 9.3** [P]: 创建角色中间件测试 `pkg/middleware/role_test.go`
  - 测试角色权限验证
  - 测试无权限拒绝访问

- [ ] **任务 9.4**: 依赖任务 9.3 - 创建角色中间件 `pkg/middleware/role.go`
  - 实现 RoleMiddleware 函数
  - 验证用户角色是否有权限访问

### 9.3 日志中间件
- [ ] **任务 9.5** [P]: 创建日志中间件测试 `pkg/middleware/logging_test.go`
  - 测试请求日志记录

- [ ] **任务 9.6**: 依赖任务 9.5 - 创建日志中间件 `pkg/middleware/logging.go`
  - 实现 LoggingMiddleware 函数
  - 记录请求方法、路径、状态码、耗时

### 9.4 恢复中间件
- [ ] **任务 9.7** [P]: 创建恢复中间件测试 `pkg/middleware/recovery_test.go`
  - 测试 Panic 捕获和恢复

- [ ] **任务 9.8**: 依赖任务 9.7 - 创建恢复中间件 `pkg/middleware/recovery.go`
  - 实现 RecoveryMiddleware 函数
  - 捕获 Panic 并返回 500 错误

---

## Phase 10: 路由注册和完善

### 10.1 Handler 基础结构
- [ ] **任务 10.1** [P]: 创建 Handler 基础测试 `internal/handler/handler_test.go`
  - 测试 Handler 结构体创建

- [ ] **任务 10.2**: 依赖任务 10.1 - 创建 Handler 基础结构 `internal/handler/handler.go`
  - 定义 `Handler` 结构体（包含所有 Service）
  - 实现 `New()` 构造函数
  - 定义公共响应函数（JSON 响应、错误响应）

### 10.2 DTO 定义
- [ ] **任务 10.3**: 创建 DTO 定义文件 `internal/handler/dto.go`
  - 定义所有请求 DTO
  - 定义所有响应 DTO
  - 定义 `ErrorResponse` 结构体

### 10.3 路由注册
- [ ] **任务 10.4**: 依赖任务 2.10, 3.10, 4.8, 5.12, 6.4, 7.4, 8.4, 9.2, 9.4, 9.6, 9.8 - 实现路由注册 `internal/handler/handler.go`
  - 实现 `RegisterRoutes()` 方法
  - 注册所有认证路由（/api/register, /api/login, /api/logout）
  - 注册所有鲜花路由（/api/flowers/*）
  - 注册所有地址路由（/api/addresses/*）
  - 注册所有订单路由（/api/orders/*）
  - 注册所有用户管理路由（/api/users/*）
  - 应用中间件（认证、角色、日志、恢复）

### 10.4 静态文件服务
- [ ] **任务 10.5**: 依赖任务 1.9 - 完善程序入口 `cmd/server/main.go`
  - 注册静态文件服务（使用 embed.FS）
  - 处理 SPA 路由（未匹配的路由返回 index.html）

---

## Phase 11: 部署配置

### 11.1 Docker 配置
- [ ] **任务 11.1**: 创建 Dockerfile `cmd/server/Dockerfile`
  - 使用 golang:1.25-alpine 作为构建镜像
  - 构建时嵌入静态文件
  - 使用 alpine:latest 作为运行镜像
  - 配置端口 8080

- [ ] **任务 11.2**: 创建 .dockerignore 文件
  - 排除不必要的文件

### 11.2 Kubernetes 配置
- [ ] **任务 11.3**: 创建 Kind 集群配置 `k8s/kind-cluster.yaml`
  - 配置 1 个 control-plane 节点
  - 配置 2 个 worker 节点

- [ ] **任务 11.4**: 创建 MetalLB 配置 `k8s/metallb-native.yaml`
  - IPAddressPool 配置

- [ ] **任务 11.5**: 创建 Ingress 配置 `k8s/ingress-nginx.yaml`

### 11.3 Helm Chart
- [ ] **任务 11.6**: 创建 Helm Chart 结构 `helm/flower-sales-system/Chart.yaml`
  - Chart.yaml 元数据

- [ ] **任务 11.7**: 创建 Helm values `helm/flower-sales-system/values.yaml`
  - 配置副本数、镜像、资源限制
  - 配置 MySQL 参数
  - 配置 Ingress

- [ ] **任务 11.8**: 创建 Helm Deployment 模板 `helm/flower-sales-system/templates/deployment-server.yaml`
  - Go Server Deployment
  - 环境变量引用 ConfigMap 和 Secret

- [ ] **任务 11.9**: 创建 Helm Deployment 模板 `helm/flower-sales-system/templates/deployment-mysql.yaml`
  - MySQL Deployment

- [ ] **任务 11.10**: 创建 Helm Service 模板 `helm/flower-sales-system/templates/service-server.yaml`
  - Go Server Service

- [ ] **任务 11.11**: 创建 Helm Service 模板 `helm/flower-sales-system/templates/service-mysql.yaml`
  - MySQL Service

- [ ] **任务 11.12**: 创建 Helm Ingress 模板 `helm/flower-sales-system/templates/ingress.yaml`
  - Ingress 配置

- [ ] **任务 11.13**: 创建 Helm PVC 模板 `helm/flower-sales-system/templates/pvc-mysql.yaml`
  - MySQL PVC

- [ ] **任务 11.14**: 创建 Helm ConfigMap 模板 `helm/flower-sales-system/templates/configmap.yaml`
  - Go Server 和 MySQL 配置

- [ ] **任务 11.15**: 创建 Helm Secret 模板 `helm/flower-sales-system/templates/secret.yaml`
  - 敏感信息（密码、密钥）

- [ ] **任务 11.16**: 创建 Helm Helper 模板 `helm/flower-sales-system/templates/_helpers.tpl`
  - 模板辅助函数

### 11.4 Makefile
- [ ] **任务 11.17**: 创建 Makefile `Makefile`
  - `make test` - 运行所有测试
  - `make build` - 构建二进制
  - `make run` - 运行服务器
  - `make docker-build` - 构建 Docker 镜像
  - `make k8s-setup` - 初始化 Kind 集群
  - `make k8s-deploy` - 部署到 K8s
  - `make k8s-clean` - 清理 K8s 资源

---

## Phase 12: 前端静态文件

### 12.1 基础页面
- [ ] **任务 12.1**: 创建首页 `web/static/index.html`
  - 单页应用框架
  - 引入 CSS 和 JS

- [ ] **任务 12.2**: 创建基础样式 `web/static/css/style.css`
  - 全局样式
  - 响应式布局

- [ ] **任务 12.3**: 创建基础脚本 `web/static/js/app.js`
  - API 客户端
  - 路由处理
  - 状态管理

### 12.2 页面组件
- [ ] **任务 12.4**: 创建登录注册页面
- [ ] **任务 12.5**: 创建鲜花列表页面
- [ ] **任务 12.6**: 创建订单管理页面
- [ ] **任务 12.7**: 创建用户管理页面（管理员）

---

## 任务执行建议

### 执行顺序
1. **Phase 1 (基础框架)** 必须首先完成
2. **Phase 2 (用户认证)** 是后续所有模块的基础
3. **Phase 3-8 (业务模块)** 可按需并行开发
4. **Phase 9-10 (中间件和路由)** 在业务模块完成后进行
5. **Phase 11-12 (部署和前端)** 最后完成

### 并行策略
- 同一 Phase 内标记 `[P]` 的测试任务可以并行编写
- 每个实现任务必须等待其对应的测试任务完成
- 不同业务模块（鲜花、地址、订单）可以并行开发

### 测试覆盖率目标
- 单元测试覆盖率 > 80%
- 关键业务逻辑（订单创建、库存扣减）必须有集成测试

---

## 附录：任务依赖图

```
Phase 1: 基础框架
├── 1.1 项目初始化
├── 1.2 go.mod
├── 1.3 → 1.4 (config)
├── 1.5 → 1.6 (database)
├── 1.7 → 1.8 (migrate)
└── 1.9 (main.go)

Phase 2: 用户认证
├── 2.1 → 2.2 (user model)
├── 2.3 → 2.4 (user repository)
├── 2.5 → 2.6 (auth service) → 2.7 → 2.8 (session)
└── 2.9 → 2.10 (auth handler)

Phase 3: 鲜花管理
├── 3.1 → 3.2 (decimal)
├── 3.3 → 3.4 (flower model)
├── 3.5 → 3.6 (flower repository)
├── 3.7 → 3.8 (flower service)
└── 3.9 → 3.10 (flower handler)

Phase 4: 地址管理
├── 4.1 → 4.2 (address model)
├── 4.3 → 4.4 (address repository)
├── 4.5 → 4.6 (address service)
└── 4.7 → 4.8 (address handler)

Phase 5: 订单核心
├── 5.1 → 5.2 (order model)
├── 5.3 → 5.4 (log model)
├── 5.5 → 5.6 (log repository)
├── 5.7 → 5.8 (order repository)
├── 5.9 → 5.10 (order service)
└── 5.11 → 5.12 (order handler)

Phase 6: 订单流转
├── 6.1 → 6.2 (extend service)
└── 6.3 → 6.4 (extend handler)

Phase 7: 用户管理
├── 7.1 → 7.2 (user service)
└── 7.3 → 7.4 (user handler)

Phase 8: 订单日志
├── 8.1 → 8.2 (log service)
├── 8.3 → 8.4 (log handler)
└── 8.5 (integrate logs)

Phase 9: 中间件
├── 9.1 → 9.2 (auth middleware)
├── 9.3 → 9.4 (role middleware)
├── 9.5 → 9.6 (logging middleware)
└── 9.7 → 9.8 (recovery middleware)

Phase 10: 路由注册
├── 10.1 → 10.2 (handler base)
├── 10.3 (dto)
├── 10.4 (register routes)
└── 10.5 (static files)

Phase 11: 部署配置
├── 11.1-11.2 (Docker)
├── 11.3-11.5 (K8s config)
├── 11.6-11.16 (Helm Chart)
└── 11.17 (Makefile)

Phase 12: 前端
├── 12.1-12.3 (base)
└── 12.4-12.7 (pages)
```

---

**文档结束**
