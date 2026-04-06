# Phase 1 & Phase 2 测试指南

## 前置要求

### 1. 启动 Redis（必需）

**Windows 环境推荐选项：**

#### 方案 A: 使用 Docker Desktop（推荐）
```powershell
docker run -d --name redis-test -p 6379:6379 redis:7-alpine
```

#### 方案 B: Windows Subsystem for Linux (WSL2)
```bash
# 在 WSL2 中运行
sudo apt-get install redis-server
redis-server
```

#### 方案 C: 下载 Redis 客户端工具
从 https://github.com/microsoftarchive/redis/releases 下载 Windows 版本

#### 验证 Redis 连接
```bash
redis-cli ping
# 预期输出: PONG
```

### 2. 环境变量配置

`.env` 文件已配置：
```env
MYSQL_DSN=root:yu199805245435@tcp(127.0.0.1:3306)/cloud_note?charset=utf8mb4&parseTime=True&loc=Local
JWT_SECRET=cloud-note-secret-2026-01-09
REDIS_ADDR=127.0.0.1:6379
REDIS_DB=0
REDIS_PASSWORD=
```

## 运行测试

### 编译
```bash
cd c:\Users\Lee\Desktop\cloud_notes
go build -o cloud_notes.exe cmd/main.go
```

### 运行测试
```bash
go test -v test_redis_blacklist.go
```

## 测试场景说明

### Phase 1: Redis 配置初始化 ✓
- [x] Redis 连接初始化
- [x] 连接池配置
- [x] 健康检查 (Ping)
- [x] 优雅关闭

**验证方式**：应用启动时应输出 `"Redis 连接成功: 127.0.0.1:6379"`

---

### Phase 2: 会话黑名单机制 ✓
- [x] Token 加入黑名单：`AddTokenToBlacklist(token, expiredAt)`
- [x] 批量加入黑名单：`AddTokensToBlacklist(tokens, expiredAt)`
- [x] 检查黑名单：`IsTokenBlacklisted(token)`
- [x] Session 模型支持 Token 和 RevokedAt

**单元测试验证**：
1. `TestAddTokenToBlacklist` - 单个 token 加入黑名单
2. `TestNonBlacklistedToken` - 验证非黑名单 token
3. `TestAddTokensToBlacklist` - 批量加入黑名单
4. `TestBlacklistExpiry` - 验证黑名单过期机制
5. `TestSessionModel` - 验证 Session 模型字段

---

## 集成测试流程

### 1. 启动应用
```bash
./cloud_notes.exe
# 预期输出:
# 数据库连接成功
# JWT 初始化成功
# Redis 连接成功: 127.0.0.1:6379
# 服务运行在 :8080
```

### 2. 测试登录
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123","device_id":"device-001"}'
# 返回: {"token":"eyJ..."}
```

### 3. 测试黑名单 - 登出后 token 立即失效
```bash
# 保存上一步返回的 token
TOKEN="eyJ..."

# 调用受保护接口（应成功）
curl -X GET http://localhost:8080/api/notebooks \
  -H "Authorization: Bearer $TOKEN"
# 预期: 200 OK

# 登出
curl -X POST http://localhost:8080/api/logout \
  -H "Authorization: Bearer $TOKEN"
# 预期: {"msg":"登出成功"}

# 再次调用受保护接口（应失败）
curl -X GET http://localhost:8080/api/notebooks \
  -H "Authorization: Bearer $TOKEN"
# 预期: 401 {"msg":"token 已被撤销"}
```

### 4. 测试整体流程
```bash
# 清理旧 docker 容器（如有）
docker stop redis-test 2>/dev/null
docker rm redis-test 2>/dev/null

# 启动 Redis
docker run -d --name redis-test -p 6379:6379 redis:7-alpine

# 运行单元测试
go test -v test_redis_blacklist.go

# 启动应用
./cloud_notes.exe

# 在另一个终端测试已上面的 curl 命令
```

---

## 预期测试结果

✓ Phase 1: Redis 连接 → 成功
✓ Phase 2: Token 黑名单  → 立即失效
✓ 集成验证：登出后重复使用 token → 拒绝（401）

---

## 常见问题

**Q: Redis 连接失败**
A: 检查 Redis 是否运行在 127.0.0.1:6379，使用 `redis-cli ping` 验证

**Q: Token 黑名单不生效**
A: 确认 Redis 中 key `blacklist:token_value` 存在，使用 `redis-cli` 查看：
```bash
redis-cli
> KEYS blacklist:*
```

**Q: 应用编译失败**
A: 运行 `go mod tidy` 更新依赖

---

## 下一步

完成 Phase 1 和 Phase 2 测试后，可继续：
- Phase 3: 设备管理优化（Sorted Set）
- Phase 4: 登录限流（Rate Limiter）
- Phase 5: 查询缓存（Repository Layer）
