# 开发环境配置

本文档详细说明了backend_go项目的开发环境搭建、配置和调试方法。

## 📋 开发环境要求

### 必需软件
- **Go**: >= 1.24.5
- **PostgreSQL**: >= 15
- **Redis**: >= 7
- **Git**: 最新版本
- **Docker**: >= 20.0 (可选，用于容器化部署)
- **Docker Compose**: >= 2.0 (可选)

### 推荐工具
- **VS Code** 或 **GoLand**: Go开发IDE
- **Postman** 或 **Insomnia**: API测试工具
- **pgAdmin** 或 **DBeaver**: 数据库管理工具
- **Redis Desktop Manager**: Redis可视化工具

## 🔧 本地开发环境搭建

### 1. 克隆项目
```bash
git clone https://github.com/yuchen1204/backend_go.git
cd backend_go
```

### 2. 安装Go依赖
```bash
go mod tidy
go mod vendor  # 生成vendor目录（Docker构建需要）
```

### 3. 配置环境变量
```bash
# 复制环境变量模板
cp env.example .env

# 编辑.env文件，配置以下关键参数：
# - SERVICE_PORT: API服务端口 (默认1234)
# - DB_*: PostgreSQL数据库配置
# - REDIS_*: Redis配置
# - SMTP_*: 邮件服务配置
# - JWT_SECRET: JWT密钥（生产环境必须修改）
```

### 4. 启动数据库服务

#### PostgreSQL
```bash
# Ubuntu/Debian
sudo systemctl start postgresql
sudo -u postgres createdb backend

# macOS (使用Homebrew)
brew services start postgresql
createdb backend

# Windows (使用Docker)
docker run --name postgres-dev -e POSTGRES_DB=backend -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres:15
```

#### Redis
```bash
# Ubuntu/Debian
sudo systemctl start redis

# macOS (使用Homebrew)
brew services start redis

# Windows (使用Docker)
docker run --name redis-dev -p 6379:6379 -d redis:7-alpine
```

### 5. 生成API文档
```bash
# 安装swag工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成Swagger文档
swag init -g cmd/main.go -o ./docs
```

### 6. 启动应用
```bash
go run cmd/main.go
```

## 🌐 端口配置

### 默认端口分配
- **API服务**: 1234 (可通过SERVICE_PORT环境变量修改)
- **管理面板**: 8081 (可通过PANEL_PORT环境变量修改)
- **PostgreSQL**: 5432
- **Redis**: 6379

### 可用端口范围
- 开发环境推荐端口范围: 1100-1200
- 避免使用系统保留端口 (1-1023)

## 🐳 Docker开发环境

### 网络配置
- 开发网络名称: `dev-net`
- 容器间通信通过网络名称解析

### 容器命名规范
- **数据库容器**: `postgresql-dev`
- **Redis容器**: `redis-dev`
- **应用容器**: `backend-app`

### 快速启动
```bash
# 使用Docker Compose启动完整环境
docker-compose -f docker-compose.multi-local.yml up -d

# 查看容器状态
docker-compose -f docker-compose.multi-local.yml ps

# 查看日志
docker-compose -f docker-compose.multi-local.yml logs -f
```

## 🔍 调试和测试

### API测试
1. **Swagger UI**: http://localhost:1234/swagger/index.html
2. **健康检查**: http://localhost:1234/health
3. **管理面板**: http://localhost:8081

### 数据库连接测试
```bash
# PostgreSQL连接测试
psql -h localhost -p 5432 -U postgres -d backend

# Redis连接测试
redis-cli -h localhost -p 6379 ping
```

### 日志调试
```bash
# 查看应用日志
go run cmd/main.go 2>&1 | tee app.log

# 查看Docker容器日志
docker logs backend-app -f
```

## 📁 项目结构说明

```
backend_go/
├── cmd/main.go              # 应用入口点
├── internal/                # 内部代码
│   ├── config/             # 配置管理
│   ├── handler/            # HTTP处理器
│   ├── middleware/         # 中间件
│   ├── model/              # 数据模型
│   ├── repository/         # 数据访问层
│   ├── service/            # 业务逻辑层
│   └── router/             # 路由配置
├── docs/                   # API文档
├── uploads/                # 文件上传目录
└── panel/                  # 管理面板静态文件
```

## 🚀 开发工作流

### 1. 功能开发流程
1. 在`internal/model/`中定义数据模型
2. 在`internal/repository/`中实现数据访问
3. 在`internal/service/`中编写业务逻辑
4. 在`internal/handler/`中实现HTTP处理器
5. 在`internal/router/`中配置路由
6. 添加Swagger注释并重新生成文档

### 2. 代码规范
- 使用`gofmt`格式化代码
- 遵循Go命名约定
- 为公开函数添加注释
- 使用依赖注入模式
- 统一错误处理

### 3. 测试建议
- 单元测试覆盖核心业务逻辑
- 集成测试验证API接口
- 使用Postman或Swagger UI进行手动测试

## 🔧 常见问题解决

### 端口占用
```bash
# 查看端口占用
netstat -tulpn | grep :1234
lsof -i :1234

# 杀死占用进程
kill -9 <PID>
```

### 数据库连接失败
1. 检查PostgreSQL服务是否启动
2. 验证数据库连接参数
3. 确认数据库用户权限
4. 检查防火墙设置

### Redis连接失败
1. 检查Redis服务状态
2. 验证Redis密码配置
3. 确认Redis配置文件

### 文件上传失败
1. 检查`uploads/`目录权限
2. 验证文件存储配置
3. 确认磁盘空间充足

## 🛠️ 开发工具配置

### VS Code推荐插件
- Go (官方Go插件)
- REST Client (API测试)
- Docker (容器管理)
- PostgreSQL (数据库连接)

### GoLand配置
- 启用Go Modules支持
- 配置代码格式化
- 设置数据库连接
- 配置Docker集成

## 📚 相关文档
- [API文档](./API_DOCUMENTATION.md)
- [项目README](./README.md)
- [Swagger文档](http://localhost:1234/swagger/index.html)
