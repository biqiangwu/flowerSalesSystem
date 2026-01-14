# 鲜花销售系统规格说明 (Specification)

## 文档信息
- **版本**: 1.0
- **创建日期**: 2026-01-14
- **项目名称**: flowerSalesSystem

---

## 1. 系统概述

### 1.1 项目目标
构建一个基于 Web 的鲜花销售管理系统，支持顾客在线下单、管理员管理商品和库存、店员处理订单。

### 1.2 技术栈
- **语言**: Go 1.25+
- **Web 框架**: `net/http` (标准库)
- **数据库**: MySQL
- **驱动**: `go-sql-driver/mysql`

---

## 2. 用户角色与权限

| 角色 | 描述 | 权限 |
|------|------|------|
| **顾客** | 终端消费者 | 浏览鲜花、下单购买、查看个人订单、管理收货地址 |
| **店员** | 店内销售人员 | 浏览鲜花、查看所有订单、处理订单、管理鲜花信息、管理库存 |
| **管理员** | 系统管理员 | 店员所有权限 + 客户管理（用户账号管理） |

---

## 3. 功能模块

### 3.1 用户认证模块

#### 3.1.1 注册
- **字段**: 用户名、密码
- **规则**:
  - 用户名唯一
  - 密码需要加密存储（使用 bcrypt）
- **是否需要审核**: 否

#### 3.1.2 登录
- **方式**: 用户名 + 密码
- **会话管理**: 使用 Session/Token 记住登录状态

#### 3.1.3 用户管理（管理员专属）
- **操作**: 查看用户列表、删除用户账号、**重置用户密码**

#### 3.1.4 重置用户密码（管理员专属）
- **字段**: 新密码
- **规则**:
  - 管理员可以为任何用户重置密码
  - 新密码需要使用 bcrypt 加密存储
  - 重置后用户可以使用新密码登录
  - 可选：强制用户下次登录时修改密码

---

### 3.2 鲜花管理模块

#### 3.2.1 鲜花信息字段
| 字段 | 类型 | 说明 | 必填 |
|------|------|------|------|
| sku | string | 商品编号/SKU，唯一标识 | 是 |
| name | string | 鲜花名称 | 是 |
| origin | string | 产地 | 是 |
| shelf_life | string | 花期 | 是 |
| preservation | string | 保鲜要求 | 是 |
| purchase_price | decimal | 进货价格 | 是 |
| sale_price | decimal | 销售价格 | 是 |
| stock | int | 库存数量（单位：支） | 是 |

#### 3.2.2 鲜花操作（店员/管理员）
- **CRUD**: 创建、读取、更新、删除鲜花信息
- **库存操作**: 进货入库、库存盘点调整

#### 3.2.3 鲜花浏览（所有角色）
- **列表展示**: 支持分页
- **搜索**: 按 sku、名称、产地搜索
- **筛选**: 按产地、价格区间筛选
- **排序**: 按价格、库存数量排序

#### 3.2.4 库存预警
- 当鲜花库存 < 10 支时，系统在列表页面展示预警提示

#### 3.2.5 删除规则
- 删除鲜花时，如果存在关联订单，标记为"已下架"而非物理删除

---

### 3.3 地址管理模块（顾客）

#### 3.3.1 地址字段
| 字段 | 类型 | 说明 | 必填 |
|------|------|------|------|
| label | string | 地址标签（如：家、公司） | 否 |
| address | string | 收货详细地址 | 是 |
| contact | string | 联系方式（电话/微信） | 是 |

#### 3.3.2 地址操作
- **CRUD**: 创建、读取、更新、删除地址
- **数量限制**: 一个用户可以有多个地址

---

### 3.4 订单模块

#### 3.4.1 订单信息字段
| 字段 | 类型 | 说明 | 必填 |
|------|------|------|------|
| order_no | string | 订单编号，唯一标识 | 是 |
| user_id | int | 下单用户ID | 是 |
| address_id | int | 收货地址ID | 是 |
| total_amount | decimal | 订单总金额 | 是 |
| status | string | 订单状态 | 是 |
| created_at | datetime | 创建时间 | 是 |
| updated_at | datetime | 更新时间 | 是 |

#### 3.4.2 订单状态
| 状态 | 说明 |
|------|------|
| pending | 待处理 |
| completed | 已完成 |
| cancelled | 已取消 |

#### 3.4.3 订单项字段
| 字段 | 类型 | 说明 | 必填 |
|------|------|------|------|
| id | int | 订单项ID | 是 |
| order_id | int | 关联订单ID | 是 |
| flower_sku | string | 鲜花SKU | 是 |
| flower_name | string | 鲜花名称（快照） | 是 |
| quantity | int | 购买数量（支） | 是 |
| unit_price | decimal | 单价（销售价格快照） | 是 |
| subtotal | decimal | 小计 | 是 |

#### 3.4.4 订单操作

**顾客**:
- 创建订单（选择收货地址、添加多种鲜花）
- 查看个人订单列表（支持分页）
- 查看订单详情
- 取消订单（仅 pending 状态）

**店员/管理员**:
- 查看所有订单（支持分页、搜索、筛选）
- 查看订单详情
- 处理订单（pending → completed）
- 取消订单（任何状态均可取消）

#### 3.4.5 下单流程
1. 顾客选择收货地址
2. 顾客添加多种鲜花到购物车/订单
3. 系统校验库存是否充足
4. 如果库存充足，扣减库存，创建订单（状态：pending）
5. 如果库存不足，拒绝下单，提示用户

#### 3.4.6 取消订单流程
1. 顾客或店员/管理员发起取消
2. 系统将订单状态改为 cancelled
3. 系统回退库存（将已扣减的库存加回）

#### 3.4.7 订单操作日志
- 记录每次订单状态变更：操作人、操作时间、操作类型、变更前状态、变更后状态

---

### 3.5 库存管理模块（店员/管理员）

#### 3.5.1 进货入库
- 选择鲜花，输入入库数量
- 库存数量增加

#### 3.5.2 库存盘点调整
- 手动调整库存数量
- 记录调整原因（可选）

#### 3.5.3 库存预警
- 鲜花库存 < 10 支时，在列表页面展示预警标识

---

## 4. 数据库设计

### 4.1 用户表 (users)
```sql
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('customer', 'clerk', 'admin') NOT NULL DEFAULT 'customer',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 4.2 地址表 (addresses)
```sql
CREATE TABLE addresses (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    label VARCHAR(50),
    address VARCHAR(255) NOT NULL,
    contact VARCHAR(50) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### 4.3 鲜花表 (flowers)
```sql
CREATE TABLE flowers (
    sku VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    origin VARCHAR(100) NOT NULL,
    shelf_life VARCHAR(50) NOT NULL,
    preservation VARCHAR(255) NOT NULL,
    purchase_price DECIMAL(10, 2) NOT NULL,
    sale_price DECIMAL(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 4.4 订单表 (orders)
```sql
CREATE TABLE orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    order_no VARCHAR(50) UNIQUE NOT NULL,
    user_id INT NOT NULL,
    address_id INT NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status ENUM('pending', 'completed', 'cancelled') NOT NULL DEFAULT 'pending',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (address_id) REFERENCES addresses(id)
);
```

### 4.5 订单项表 (order_items)
```sql
CREATE TABLE order_items (
    id INT PRIMARY KEY AUTO_INCREMENT,
    order_id INT NOT NULL,
    flower_sku VARCHAR(50) NOT NULL,
    flower_name VARCHAR(100) NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);
```

### 4.6 订单操作日志表 (order_logs)
```sql
CREATE TABLE order_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    order_id INT NOT NULL,
    operator_id INT NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (operator_id) REFERENCES users(id)
);
```

---

## 5. API 设计（RESTful）

### 5.1 认证相关
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | /api/register | 用户注册 | 公开 |
| POST | /api/login | 用户登录 | 公开 |
| POST | /api/logout | 用户登出 | 登录用户 |

### 5.2 鲜花相关
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/flowers | 获取鲜花列表（支持搜索、筛选、分页） | 所有用户 |
| GET | /api/flowers/:sku | 获取鲜花详情 | 所有用户 |
| POST | /api/flowers | 创建鲜花 | 店员、管理员 |
| PUT | /api/flowers/:sku | 更新鲜花信息 | 店员、管理员 |
| DELETE | /api/flowers/:sku | 删除/下架鲜花 | 店员、管理员 |
| POST | /api/flowers/:sku/stock | 进货入库 | 店员、管理员 |

### 5.3 地址相关
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/addresses | 获取当前用户地址列表 | 顾客 |
| POST | /api/addresses | 创建地址 | 顾客 |
| PUT | /api/addresses/:id | 更新地址 | 顾客 |
| DELETE | /api/addresses/:id | 删除地址 | 顾客 |

### 5.4 订单相关
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/orders | 获取订单列表（顾客：自己的，店员/管理员：所有） | 登录用户 |
| GET | /api/orders/:id | 获取订单详情 | 登录用户 |
| POST | /api/orders | 创建订单 | 顾客 |
| PUT | /api/orders/:id/complete | 完成订单 | 店员、管理员 |
| PUT | /api/orders/:id/cancel | 取消订单 | 登录用户 |
| GET | /api/orders/:id/logs | 获取订单操作日志 | 登录用户 |

### 5.5 用户管理（管理员）
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/users | 获取用户列表 | 管理员 |
| DELETE | /api/users/:id | 删除用户 | 管理员 |
| PUT | /api/users/:id/password | 重置用户密码 | 管理员 |

---

## 6. 非功能性需求

### 6.1 安全性
- 密码使用 bcrypt 加密存储
- 实现 CSRF 防护
- SQL 注入防护（使用参数化查询）

### 6.2 性能
- 列表接口支持分页，默认每页 20 条
- 数据库查询使用索引优化

### 6.3 可维护性
- 代码遵循 Go 语言规范
- 关键业务逻辑有单元测试覆盖
- 使用清晰的错误消息

---

## 7. 边缘场景处理

| 场景 | 处理方式 |
|------|----------|
| 下单时库存不足 | 拒绝下单，返回具体哪些鲜花库存不足 |
| 取消已完成的订单 | 允许取消，库存回退 |
| 删除有订单关联的鲜花 | 软删除（标记 is_active = false） |
| 删除用户账号 | 级联删除其地址，订单保留（user_id 置空或标记） |
| 并发下单同一商品 | 使用数据库事务保证库存扣减的原子性 |
| 订单金额精度 | 使用 DECIMAL 类型，保证金额计算精确 |

---

## 8. 开发里程碑

### Phase 1: 基础框架
- 项目结构搭建
- 数据库连接与表结构创建
- 用户认证（注册、登录、Session）

### Phase 2: 鲜花管理
- 鲜花 CRUD
- 鲜花列表（搜索、筛选、分页）

### Phase 3: 地址与订单
- 地址管理
- 订单创建与流转
- 库存自动扣减与回退

### Phase 4: 库存管理
- 进货入库
- 库存盘点调整
- 库存预警

### Phase 5: 管理功能
- 订单操作日志
- 用户管理

---

## 9. 附录

### 9.1 订单编号生成规则
格式：`ORD` + `YYYYMMDD` + `6位随机数`
示例：`ORD20260114123456`

### 9.2 状态流转图
```
下单 → pending → completed
        ↓
      cancelled
```

---

## 10. 部署方式

### 10.1 部署架构概览

系统采用 Kubernetes 容器编排，使用微服务架构部署：

```
                    ┌─────────────────┐
                    │    Ingress      │
                    │   (nginx)       │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  MetalLB (LB)   │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼────────┐  ┌───────▼────────┐  ┌───────▼────────┐
│   Web Frontend │  │   API Server   │  │     MySQL      │
│   (Deployment) │  │  (Deployment)  │  │  (Deployment)  │
│   Replicas: 2  │  │  Replicas: 2   │  │  Replicas: 1   │
└────────────────┘  └────────────────┘  └───────┬────────┘
                                                  │
                                          ┌───────▼────────┐
                                          │      PVC       │
                                          │  (Storage)     │
                                          └────────────────┘
```

### 10.2 组件说明

| 组件 | 类型 | 副本数 | 存储需求 |
|------|------|--------|----------|
| Web Frontend | Deployment | 2 | 无 |
| API Server | Deployment | 2 | 无 |
| MySQL | Deployment | 1 | PVC (10Gi) |

### 10.3 运行时版本要求

| 组件 | 版本要求 |
|------|----------|
| **Go** | v1.25 |
| **Kubernetes** | v1.34 |
| **Kind (测试)** | v1.34 |

**Go 版本说明**:
- API Server 使用 Go v1.25 编译
- go.mod 中指定 `go 1.25`
- Docker 基础镜像使用 `golang:1.25-alpine`

### 10.4 本地测试环境 (Kind)

#### 10.4.1 Kind 集群配置
- **Master 节点**: 1 个
- **Worker 节点**: 2 个
- **K8s 版本**: v1.34

#### 10.4.2 Kind 集群创建
```bash
kind create cluster --config=k8s/kind-cluster.yaml
```

`k8s/kind-cluster.yaml`:
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: flower-sales-system
nodes:
  - role: control-plane
  - role: worker
  - role: worker
```

### 10.5 网络与负载均衡

#### 10.5.1 MetalLB (LoadBalancer)
用于在 Kind/裸金属环境提供 LoadBalancer 类型的 Service。

**安装方式**: 使用 kubectl manifest
```bash
kubectl apply -f k8s/metallb-native.yaml
```

**IPAddressPool 配置**:
```yaml
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: example-pool
  namespace: metallb-system
spec:
  addresses:
    - 172.18.255.200-172.18.255.250
```

#### 10.5.2 Ingress
使用 NGINX Ingress Controller 对外暴露服务。

**安装方式**:
```bash
kubectl apply -f k8s/ingress-nginx.yaml
```

**Ingress 规则**:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: flower-sales-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
    - host: flower.local
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: api-server
                port:
                  number: 80
          - path: /
            pathType: Prefix
            backend:
              service:
                name: web-frontend
                port:
                  number: 80
```

### 10.6 存储 (PVC)

MySQL 数据持久化使用 PVC：

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard
```

**Kind 环境 StorageClass**: 使用 `local-path` 或 `standard` (默认)。

### 10.7 Helm Chart

#### 10.7.1 Chart 结构
```
helm/flower-sales-system/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── _helpers.tpl
│   ├── deployment-web.yaml
│   ├── deployment-api.yaml
│   ├── deployment-mysql.yaml
│   ├── service-web.yaml
│   ├── service-api.yaml
│   ├── service-mysql.yaml
│   ├── ingress.yaml
│   ├── pvc-mysql.yaml
│   └── configmap.yaml
```

#### 10.7.2 values.yaml 示例
```yaml
replicaCount: 2

image:
  repository: ghcr.io/biqiangwu/flower-sales-system
  pullPolicy: IfNotPresent
  tag: "v1.0.0"

web:
  containerPort: 80
  resources:
    limits:
      cpu: 200m
      memory: 128Mi
    requests:
      cpu: 100m
      memory: 64Mi

api:
  containerPort: 8080
  resources:
    limits:
      cpu: 500m
      memory: 256Mi
    requests:
      cpu: 250m
      memory: 128Mi

mysql:
  containerPort: 3306
  database: flower_sales
  username: flower_user
  # password: 使用 secret
  persistence:
    size: 10Gi
    storageClass: standard
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 250m
      memory: 256Mi

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: flower.local
      paths:
        - path: /api
          service: api-server
          port: 80
        - path: /
          service: web-frontend
          port: 80
```

#### 10.7.3 部署命令
```bash
# 构建镜像
make docker-build

# 加载到 Kind 集群
kind load docker-image flower-sales-system:v1.0.0 --name flower-sales-system

# 使用 Helm 部署
helm install flower-sales-system helm/flower-sales-system \
  --namespace flower-sales \
  --create-namespace
```

### 10.8 部署流程

#### 10.8.1 初始化顺序
1. 创建 Kind 集群
2. 安装 MetalLB
3. 安装 NGINX Ingress Controller
4. 部署 MySQL (等待 Ready)
5. 运行数据库迁移 (创建表结构)
6. 部署 API Server
7. 部署 Web Frontend
8. 配置 Ingress

#### 10.8.2 Makefile 目标
```makefile
.PHONY: k8s-setup k8s-deploy k8s-clean

k8s-setup:
	kind create cluster --config=k8s/kind-cluster.yaml
	kubectl apply -f k8s/metallb-native.yaml
	kubectl apply -f k8s/ingress-nginx.yaml

k8s-deploy: docker-build
	kind load docker-image flower-sales-system:v1.0.0 --name flower-sales-system
	helm install flower-sales-system helm/flower-sales-system --namespace flower-sales --create-namespace
	kubectl -n flower-sales rollout status deployment

k8s-clean:
	helm uninstall flower-sales-system -n flower-sales || true
	kind delete cluster --name flower-sales-system
```

### 10.9 环境变量配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_HOST` | MySQL 服务地址 | mysql-service |
| `DB_PORT` | MySQL 端口 | 3306 |
| `DB_NAME` | 数据库名称 | flower_sales |
| `DB_USER` | 数据库用户 | flower_user |
| `DB_PASSWORD` | 数据库密码 | - |
| `JWT_SECRET` | JWT 密钥 | - |
| `SERVER_PORT` | API 服务端口 | 8080 |
