# Cloud Notes - 云端笔记管理系统

Cloud Notes 是一个基于 Go 语言开发的云端笔记管理系统，提供高性能的 RESTful API 接口。本项目采用 Clean Architecture 分层架构，支持多用户隔离、笔记本分类、标签系统和会话安全策略，适用于个人笔记存储和协作场景。

## 核心功能

- **用户系统**: 注册与登录，基于 JWT 的认证，密码使用 `bcrypt` 哈希。
- **笔记管理**: 创建、编辑、删除笔记，支持富文本与默认笔记本机制。
- **笔记本组织**: 管理笔记本，支持级联删除及标题唯一性。
- **标签系统**: 支持多对多标签关联、按标签过滤和标签删除时解绑。
- **全文搜索**: 支持 `q` 关键词搜索，覆盖标题、内容与标签。
- **安全与会话管理**: JWT 黑名单即时失效、活跃设备限制、登录限流和 Redis 缓存加速。

## 技术栈

| 维度 | 技术选型 | 说明 |
| :--- | :--- | :--- |
| **语言** | Go 1.25.5 | 高效并发编程语言 |
| **Web 框架** | Gin | 轻量级、高性能 HTTP 框架 |
| **数据库** | MySQL | 持久化存储 |
| **ORM** | GORM | 结构化数据库访问 |
| **鉴权** | JWT (HS256) | 无状态认证 |
| **缓存/会话** | Redis | Token 黑名单、设备集合、登录限流、结果缓存 |
| **配置** | godotenv | 环境变量加载 |

## 系统架构

项目遵循 Clean Architecture 设计，确保业务逻辑、HTTP 处理和数据访问相互独立：

- `handler/` - HTTP 请求处理与响应。
- `service/` - 业务逻辑和会话管理。
- `repository/` - 数据访问和 Redis 支持。
- `model/` - 数据实体定义。
- `middleware/` - JWT 认证与登录限流。
- `router/` - 路由注册与分组。

## API 概览

所有受保护接口需携带：`Authorization: Bearer <token>`。

### 公共接口

- `POST /api/register` - 用户注册
- `POST /api/login` - 用户登录

### 认证后接口

- `POST /api/logout` - 单设备登出
- `POST /api/logout-all` - 退出所有设备

### 笔记管理

- `GET /api/notes` - 查询笔记（支持 `notebook_id`, `tag`, `q`）
- `POST /api/notes` - 创建笔记
- `PUT /api/notes/:id` - 更新笔记
- `DELETE /api/notes/:id` - 删除笔记

### 笔记本管理

- `GET /api/notebooks` - 列出笔记本
- `POST /api/notebooks` - 创建笔记本
- `PUT /api/notebooks/:id` - 更新笔记本
- `DELETE /api/notebooks/:id` - 删除笔记本（级联删除笔记）

### 标签管理

- `GET /api/tags` - 列出标签
- `POST /api/tags` - 创建标签
- `POST /api/notes/:id/tags` - 绑定标签到笔记
- `GET /api/tags/:id/notes` - 获取标签下笔记
- `DELETE /api/tags/:id` - 删除标签并解绑关联

## 快速开始

### 前置要求

- Go `1.25.5+`
- MySQL `5.7+`
- Redis `6+`

### 环境变量配置

在项目根目录创建 `.env`：

```ini
MYSQL_DSN=user:password@tcp(127.0.0.1:3306)/cloud_notes?charset=utf8mb4&parseTime=True&loc=Local
JWT_SECRET=your-secure-secret-key-here
REDIS_ADDR=127.0.0.1:6379
REDIS_DB=0
REDIS_PASSWORD=
```

### 启动项目

```bash
go mod tidy
go run cmd/main.go
```

默认监听 `http://localhost:8080`。

## Redis 关键实现

项目使用 Redis 提升安全性与性能：

- **JWT 黑名单**: 用户登出或会话撤销时，将 token 写入 Redis 黑名单，确保立即失效。
- **活跃设备管理**: 使用 Redis Sorted Set 维护每个用户的活跃设备，最多保留 2 个设备，并在超限时自动撤销最旧会话。
- **登录限流**: `POST /api/login` 按 IP + username 限制请求频率，1 分钟内最多 10 次。
- **结果缓存**: 对笔记、笔记本、标签列表结果进行 Redis 缓存，加速高频读取。
- **缓存失效**: 数据变更时清理相关缓存，保持读取一致性。

## 会话与安全策略

- 登录成功后生成 JWT，包含 `user_id` 和 `device_id`。
- JWT 中间件校验签名、Token 黑名单、数据库会话有效性与设备限制。
- 单设备登出时撤销当前会话并加入黑名单。
- 全局登出时撤销用户所有会话并清空活跃设备集合。

## 项目结构

```text
cloud_notes/
├── cmd/
│   └── main.go            # 程序入口
├── internal/
│   ├── config/            # DB、Redis、JWT 初始化
│   ├── handler/           # HTTP 接口处理
│   ├── middleware/        # 认证与限流
│   ├── model/             # 数据结构
│   ├── repository/        # 数据访问与 Redis 操作
│   ├── router/            # 路由注册
│   └── service/           # 业务逻辑
├── .env                   # 环境变量示例
└── go.mod                 # 依赖管理
```

## 开发规范

1. 执行 `gofmt` 保持格式一致。
2. 使用 `go mod tidy` 管理依赖。
3. 区分业务错误（4xx）与系统错误（5xx）。
4. Repository 操作需遵循 `user_id` 访问约束。
5. 机密信息通过环境变量管理。
