# Backend 用户注册系统

这是一个基于 Go 语言和 Gin 框架的用户注册系统后端项目。

## 功能特性

- 用户注册功能
- 密码加盐哈希存储
- 用户信息查询（按ID和用户名）
- **文件上传系统**（支持本地存储和S3）
- **多存储配置**（灵活配置多个存储桶）
- 文件管理功能（上传、下载、删除、更新）
- RESTful API 设计
- PostgreSQL 数据库支持
- 统一响应格式
- CORS 跨域支持
- **完整的API文档** (Swagger/OpenAPI)

## 技术栈

- **语言**: Go 1.24.5
- **框架**: Gin
- **数据库**: PostgreSQL
- **ORM**: GORM
- **UUID**: Google UUID
- **文件存储**: 本地存储 + AWS S3
- **文档**: Swagger/OpenAPI

## API 文档

### 在线文档
启动服务后，访问以下地址查看完整的交互式API文档：

🌐 **Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### 生成文档
```bash
# 生成API文档
chmod +x scripts/generate-docs.sh
./scripts/generate-docs.sh

# 或者手动生成
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/main.go -o ./docs
```

### 文档特性
- **交互式测试**: 可以直接在浏览器中测试API
- **认证支持**: 支持JWT Bearer Token认证
- **完整的请求/响应示例**: 包含所有字段的详细说明
- **错误代码说明**: 详细的错误响应文档

## 项目结构

```
backend/
├── cmd/                    # 应用程序入口
│   └── main.go
├── internal/              # 内部代码
│   ├── config/           # 配置相关
│   │   └── database.go
│   ├── handler/          # HTTP 处理器
│   │   ├── response.go
│   │   └── user_handler.go
│   ├── model/            # 数据模型
│   │   └── user.go
│   ├── repository/       # 数据访问层
│   │   └── user_repository.go
│   ├── router/           # 路由配置
│   │   └── router.go
│   └── service/          # 业务逻辑层
│       └── user_service.go
├── configs/              # 配置文件
│   └── env.example
├── go.mod
└── README.md
```

## 用户表结构

用户表包含以下字段：

- `id`: UUID 主键
- `username`: 用户名（唯一）
- `email`: 邮箱地址（唯一）
- `password_salt`: 密码盐和哈希（格式：salt:hash）
- `nickname`: 昵称
- `bio`: 个人简介
- `avatar`: 头像URL
- `created_at`: 创建时间
- `updated_at`: 更新时间
- `deleted_at`: 软删除时间

## API 接口

### 发送注册验证码
- **POST** `/api/v1/users/send-code`
- **描述**: 在发送验证码前，会预先检查用户名和邮箱是否都未被注册。都通过后，才会向指定邮箱发送一个用于注册的6位数验证码（5分钟内有效）。

**请求体示例:**
```json
{
    "username": "testuser",
    "email": "test@example.com"
}
```

**响应示例:**
```json
{
    "code": 200,
    "message": "验证码已发送至您的邮箱，请注意查收",
    "data": null,
    "timestamp": 1640995200
}
```

### 用户注册
- **POST** `/api/v1/users/register`
- **描述**: 使用邮箱验证码创建新用户账户

**请求体示例:**
```json
{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "verification_code": "123456",
    "nickname": "测试用户",
    "bio": "这是我的个人简介",
    "avatar": "https://example.com/avatar.jpg"
}
```

**响应示例:**
```json
{
    "code": 201,
    "message": "注册成功",
    "data": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "username": "testuser",
        "email": "test@example.com",
        "nickname": "测试用户",
        "bio": "这是我的个人简介",
        "avatar": "https://example.com/avatar.jpg",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    },
    "timestamp": 1640995200
}
```

### 用户登录
- **POST** `/api/v1/users/login`
- **描述**: 使用用户名和密码登录，成功后返回包含Access Token、Refresh Token和用户信息的对象。

**请求体示例:**
```json
{
    "username": "testuser",
    "password": "password123"
}
```

**响应示例:**
```json
{
    "code": 200,
    "message": "登录成功",
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "user": {
            "id": "123e4567-e89b-12d3-a456-426614174000",
            "username": "testuser",
            "email": "test@example.com",
            "nickname": "测试用户",
            "bio": "这是我的个人简介",
            "avatar": "https://example.com/avatar.jpg",
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        }
    },
    "timestamp": 1640995200
}
```

### 刷新访问Token
- **POST** `/api/v1/users/refresh`
- **描述**: 使用有效的Refresh Token获取新的Access Token。

**请求体示例:**
```json
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应示例:**
```json
{
    "code": 200,
    "message": "刷新成功",
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    },
    "timestamp": 1640995200
}
```

### 用户登出
- **POST** `/api/v1/users/logout`
- **描述**: 登出用户并撤销所有Token（Access Token和Refresh Token）。Access Token将被加入黑名单立即失效，Refresh Token也将被删除。

**请求体示例:**
```json
{
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应示例:**
```json
{
    "code": 200,
    "message": "登出成功",
    "data": null,
    "timestamp": 1640995200
}
```

### 发送重置密码验证码
- **POST** `/api/v1/users/send-reset-code`
- **描述**: 向指定邮箱发送用于重置密码的6位数验证码（5分钟内有效）。为了安全考虑，即使邮箱未注册也会返回成功，避免邮箱枚举攻击。

**请求体示例:**
```json
{
    "email": "test@example.com"
}
```

**响应示例:**
```json
{
    "code": 200,
    "message": "验证码已发送至您的邮箱，请注意查收",
    "data": null,
    "timestamp": 1640995200
}
```

### 重置密码
- **POST** `/api/v1/users/reset-password`
- **描述**: 使用邮箱验证码重置用户密码。重置成功后，该用户的所有refresh token将被撤销，需要重新登录。

**请求体示例:**
```json
{
    "email": "test@example.com",
    "verification_code": "123456",
    "new_password": "newpassword123"
}
```

**响应示例:**
```json
{
    "code": 200,
    "message": "密码重置成功，请使用新密码登录",
    "data": null,
    "timestamp": 1640995200
}
```

### 获取当前用户信息
- **GET** `/api/v1/users/me`
- **描述**: 需要在请求头中提供有效的Access Token来获取当前登录用户的详细信息。
- **认证**: `Bearer Token` (仅接受Access Token)

**请求头示例:**
```
Authorization: Bearer <your-access-token>
```

**响应示例:**
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "username": "testuser",
        // ... a UserResponse object
    },
    "timestamp": 1640995200
}
```

### 更新当前用户信息
- **PUT** `/api/v1/users/me`
- **描述**: 更新当前登录用户的基本信息（昵称、简介、头像）。
- **认证**: `Bearer Token` (仅接受Access Token)

**请求头示例:**
```
Authorization: Bearer <your-access-token>
```

**请求体示例:**
```json
{
    "nickname": "新昵称",
    "bio": "我的新个人简介",
    "avatar": "https://example.com/new-avatar.jpg"
}
```

**注意事项:**
- 所有字段都是可选的，只更新提供的字段
- 如果某个字段为空字符串或未提供，该字段不会被更新
- `avatar` 字段如果提供，必须是有效的URL格式

**响应示例:**
```json
{
    "code": 200,
    "message": "更新成功",
    "data": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "username": "testuser",
        "email": "test@example.com",
        "nickname": "新昵称",
        "bio": "我的新个人简介",
        "avatar": "https://example.com/new-avatar.jpg",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T12:30:00Z"
    },
    "timestamp": 1640995200
}
```

### 根据ID获取用户
- **GET** `/api/v1/users/{id}`
- **描述**: 通过用户ID获取用户详细信息

### 根据用户名获取用户
- **GET** `/api/v1/users/username/{username}`
- **描述**: 通过用户名获取用户详细信息

### 健康检查
- **GET** `/health`
- **描述**: 服务健康状态检查

## 环境配置

复制 `configs/env.example` 文件并根据需要修改配置：

```bash
# 服务器配置
PORT=8080

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=backend
DB_SSLMODE=disable

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# SMTP 邮件服务配置
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-email@example.com
SMTP_PASSWORD=your-email-password
SMTP_FROM=your-email@example.com

# 安全配置
MAX_IP_REQUESTS_PER_DAY=10
```

## 快速开始

1. **克隆项目**
```bash
git clone <repository-url>
cd backend
```

2. **安装依赖**
```bash
go mod tidy
```

3. **生成API文档**
```bash
./scripts/generate-docs.sh
```

4. **设置环境变量**
```bash
cp configs/env.example .env
# 编辑 .env 文件设置数据库连接信息
```

5. **启动 PostgreSQL 数据库和 Redis**
```bash
# 手动安装并启动 PostgreSQL 和 Redis 服务
# PostgreSQL 安装: sudo apt-get install postgresql postgresql-contrib
# Redis 安装: sudo apt-get install redis-server

# 创建数据库
createdb backend
```

6. **运行应用**
```bash
go run cmd/main.go
```

7. **访问API文档**
```bash
# 在浏览器中访问
http://localhost:8080/swagger/index.html
```

服务器将在 `http://localhost:8080` 启动。

## API 接口概览

### 🔓 公开接口（无需认证）
- **POST** `/api/v1/users/send-code` - 发送注册验证码
- **POST** `/api/v1/users/register` - 用户注册
- **POST** `/api/v1/users/login` - 用户登录
- **POST** `/api/v1/users/refresh` - 刷新访问Token
- **POST** `/api/v1/users/logout` - 用户登出
- **POST** `/api/v1/users/send-reset-code` - 发送重置密码验证码
- **POST** `/api/v1/users/reset-password` - 重置密码
- **GET** `/api/v1/users/{id}` - 根据ID获取用户信息
- **GET** `/api/v1/users/username/{username}` - 根据用户名获取用户信息
- **GET** `/health` - 健康检查

### 🔒 需要认证的接口
- **GET** `/api/v1/users/me` - 获取当前用户信息
- **PUT** `/api/v1/users/me` - 更新当前用户信息

### 📁 文件管理接口

#### 🔓 公开接口
- **GET** `/api/v1/files/public` - 获取公开文件列表
- **GET** `/api/v1/files/storages` - 获取存储信息
- **GET** `/api/v1/files/{id}` - 获取文件详情（支持公开和私有）

#### 🔒 需要认证的接口
- **POST** `/api/v1/files/upload` - 上传单个文件
- **POST** `/api/v1/files/upload-multiple` - 上传多个文件
- **GET** `/api/v1/files/my` - 获取当前用户文件列表
- **PUT** `/api/v1/files/{id}` - 更新文件信息
- **DELETE** `/api/v1/files/{id}` - 删除文件

## 测试 API

### 方法 1: 使用 Swagger UI (推荐)
1. 启动服务器
2. 访问 [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
3. 在页面右上角点击"Authorize"按钮
4. 输入Bearer Token: `Bearer your-access-token`
5. 直接在页面中测试各个API

### 方法 2: 使用 curl
详细的curl命令请参考下面的"详细API文档"部分。

## 开发工具

### API文档生成
```bash
# 安装swag工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g cmd/main.go -o ./docs

# 重新生成文档（开发时）
./scripts/generate-docs.sh
```

### 代码格式化
```bash
# 格式化代码
go fmt ./...

# 代码检查
go vet ./...
```

## 文件上传功能

### 特性
- **多存储支持**: 同时支持本地存储和AWS S3存储
- **灵活配置**: 可配置多个存储桶，每个存储桶独立配置
- **文件分类**: 支持按类别组织文件（avatar、document、image等）
- **权限控制**: 支持公开和私有文件访问控制
- **文件管理**: 完整的CRUD操作（创建、读取、更新、删除）

### 配置示例

#### 本地存储配置
```bash
# 支持多个本地存储
FILE_STORAGE_LOCAL_NAMES=default,avatar,document
FILE_STORAGE_LOCAL_DEFAULT_PATH=./uploads/default
FILE_STORAGE_LOCAL_DEFAULT_URL=http://localhost:8080/uploads/default
```

#### S3存储配置
```bash
# 支持多个S3存储桶
FILE_STORAGE_S3_NAMES=main,backup
FILE_STORAGE_S3_MAIN_REGION=us-east-1
FILE_STORAGE_S3_MAIN_BUCKET=my-app-files
FILE_STORAGE_S3_MAIN_ACCESS_KEY=your-access-key
FILE_STORAGE_S3_MAIN_SECRET_KEY=your-secret-key
```

### 使用示例

#### 上传文件
```bash
curl -X POST "http://localhost:8080/api/v1/files/upload" \
  -H "Authorization: Bearer your-access-token" \
  -F "file=@example.jpg" \
  -F "storage_name=avatar" \
  -F "category=profile" \
  -F "is_public=true"
```

#### 获取文件列表
```bash
curl -X GET "http://localhost:8080/api/v1/files/my?category=profile&page=1&page_size=10" \
  -H "Authorization: Bearer your-access-token"
```

## 部署

### 本地部署
```bash
# 编译应用
go build -o backend cmd/main.go

# 运行应用
./backend
```

### 环境变量
生产环境需要设置的关键环境变量：
- `JWT_SECRET`: JWT签名密钥（必须修改）
- `DB_PASSWORD`: 数据库密码
- `REDIS_PASSWORD`: Redis密码
- `SMTP_*`: 邮件服务配置
- `FILE_STORAGE_*`: 文件存储配置

## 安全特性

- 密码使用加盐哈希存储
- **双Token机制**：
  - **Access Token**: 短期有效（默认30分钟），用于API访问
  - **Refresh Token**: 长期有效（默认7天），仅用于刷新Access Token
  - 提升安全性的同时保持良好的用户体验
- **Token黑名单机制**：
  - **Access Token黑名单**: 登出时Access Token立即加入黑名单失效
  - **Refresh Token管理**: 登出时删除Refresh Token，防止再次使用
  - 确保用户登出后所有Token立即失效，消除安全隐患
- **JWT会话管理**：用户登录后使用JWT进行无状态认证。
- **IP请求频率限制**：限制每个IP每天请求验证码的次数，防止接口被恶意攻击。
- 响应中不包含敏感信息（密码）
- 输入验证和参数绑定
- 统一错误处理

## 开发说明

- 使用分层架构设计（Handler -> Service -> Repository）
- 依赖注入模式
- 接口驱动开发
- GORM 自动数据库迁移
- 统一响应格式 