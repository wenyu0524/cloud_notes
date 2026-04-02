太棒了！这是一个非常标准且结构清晰的 **Cloud Notes** 项目架构。为了让你的项目在 GitHub 或其他平台展示时显得专业且易于上手，我为你整理了一份格式化的 `README.md`。

这份文档不仅包含你提供的所有信息，还采用了 Markdown 最佳实践（如徽章、代码块、任务列表等）。

-----

# 🚀 Cloud Notes - 云端笔记管理系统

[](https://golang.org)
[](https://gin-gonic.com)
[](https://gorm.io)
[](https://www.google.com/search?q=LICENSE)

Cloud Notes 是一个基于 Go 语言开发的云端笔记管理系统，提供高性能的 RESTful API 接口。本项目采用 **Clean Architecture** 分层架构，支持多用户隔离、笔记本分类及灵活的标签系统，适用于个人笔记存储或协作场景。

-----

## 🌟 核心功能

  * **用户系统**: 支持注册、登录，基于 JWT 的状态保持（24小时有效），密码采用 `bcrypt` 强哈希加密。
  * **笔记管理**: 创建、编辑、删除笔记，支持富文本内容，提供自动分配默认笔记本功能。
  * **组织架构**: 笔记本层级分类，支持在笔记本内保证标题唯一性。
  * **标签系统**: 灵活的多对多标签关联，支持去重及按标签快速过滤。
  * **安全保障**: 数据实现物理层级的用户隔离，使用数据库事务确保联级操作的数据一致性。

-----

## 🛠️ 技术栈

| 维度 | 技术选型 | 说明 |
| :--- | :--- | :--- |
| **语言** | Go 1.25.5 | 高效并发编程语言 |
| **Web 框架** | Gin v1.11.0 | 轻量级、高性能 HTTP 框架 |
| **数据库** | MySQL | 关系型数据库，存储持久化数据 |
| **ORM** | GORM | 自动迁移 (Auto-migrate) 与 复杂 SQL 映射 |
| **鉴权** | JWT (HS256) | 基于 Token 的无状态身份验证 |
| **配置** | godotenv | 环境变量管理 |

-----

## 🏗️ 系统架构

项目遵循 **Clean Architecture** 设计原则，确保了业务逻辑与底层实现的高度解耦：

  * `Handler 层`: 负责路由分发、参数绑定及 HTTP 响应。
  * `Service 层`: 核心业务逻辑实现，处理权限检查与复杂业务规约。
  * `Repository 层`: 数据访问层，封装 GORM 操作。
  * `Model 层`: 定义数据库 Schema 及 GORM 标签。
  * `Middleware 层`: 拦截器（如 JWT 鉴权中间件）。

-----

## 🗺️ API 概览

所有受保护接口需携带：`Authorization: Bearer <your_token>`

### 公共接口

  * `POST /api/register` - 用户注册
  * `POST /api/login` - 用户登录

### 笔记与组织

  * `GET /api/notes` - 列表查询（支持 `notebook_id`, `tag` 过滤）
  * `POST /api/notes` - 创建笔记
  * `PUT /api/notebooks/:id` - 更新笔记本信息
  * `DELETE /api/tags/:id` - 级联解绑并删除标签

-----

## 📦 快速开始

### 1\. 前置要求

  * Go `1.25.5+`
  * MySQL `5.7+`

### 2\. 配置环境

在项目根目录创建 `.env` 文件：

```ini
MYSQL_DSN=user:password@tcp(127.0.0.1:3306)/cloud_notes?charset=utf8mb4&parseTime=True&loc=Local
JWT_SECRET=your-secure-secret-key-here
```

### 3\. 安装运行

```bash
# 下载依赖
go mod tidy

# 启动服务
go run cmd/main.go
```

服务器默认运行在 `http://localhost:8080`。

-----

## 📂 项目结构

```text
cloud_notes/
├── cmd/                # 程序入口
├── internal/
│   ├── config/         # 配置初始化
│   ├── handler/        # 接口处理
│   ├── service/        # 业务逻辑
│   ├── repository/     # 数据库交互
│   ├── model/          # 数据实体
│   ├── middleware/     # 中间件 (JWT)
│   └── router/         # 路由注册
├── .env                # 环境变量
└── go.mod              # 依赖管理
```

-----

## 🛡️ 开发规范

1.  **代码风格**: 严格执行 `gofmt`。
2.  **依赖注入**: 避免使用 `init()` 函数和全局变量。
3.  **错误处理**: 区分业务错误（4xx）与系统错误（5xx）。
4.  **数据安全**: 任何 Repository 操作必须携带 `userID` 约束。