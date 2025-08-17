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

## 使用 Docker Compose 运行 (推荐)

本项目提供了 Docker Compose 配置，可以一键启动应用所需的所有服务（PostgreSQL, Redis, Go Backend）。这是最简单、最推荐的运行方式。

### 1. 准备工作

首先，请确保你已经安装了 `Docker` 和 `docker-compose`。

> 依赖使用 vendoring（`vendor/` 目录），构建前请先生成 vendor：
```bash
go mod tidy
go mod vendor
```

### 2. 选择配置文件

我们提供了两种开箱即用的配置：

- **`docker-compose.multi-local.yml`**: 使用本地文件系统作为存储。上传的文件会保存在项目根目录的 `uploads/` 文件夹下。
- **`docker-compose.multi-s3.yml`**: 使用 AWS S3 作为文件存储。

### 3. 启动服务

#### 选项 A: 使用本地存储 (适合快速测试)

```bash
# 使用 multi-local 配置文件启动所有服务
docker-compose -f docker-compose.multi-local.yml up --build -d
```

#### 选项 B: 使用 S3 存储 (适合生产或模拟生产环境)

**在启动前**，请务-必打开 `docker-compose.multi-s3.yml` 文件，并将其中所有 `YOUR_...` 占位符替换为你的真实 AWS S3 凭证。

```yaml
      # ...
      # S3 存储 'primary' 的配置
      FILE_STORAGE_S3_PRIMARY_REGION: "us-east-1"  # <- 修改这里
      FILE_STORAGE_S3_PRIMARY_BUCKET: "your-primary-bucket" # <- 修改这里
      FILE_STORAGE_S3_PRIMARY_ACCESS_KEY: "YOUR_PRIMARY_ACCESS_KEY" # <- 修改这里
      FILE_STORAGE_S3_PRIMARY_SECRET_KEY: "YOUR_PRIMARY_SECRET_KEY" # <- 修改这里
      # ...
```

然后运行以下命令启动：

```bash
# 使用 multi-s3 配置文件启动所有服务
docker-compose -f docker-compose.multi-s3.yml up --build -d
```

### 3.1 环境变量与 Compose 插值（重要）

- 运行 Compose 时，项目根目录的 `.env`（Compose 专用）会在“解析阶段”用于变量插值；而 `env_file`（如 `./configs/.env`）只在容器内生效。
- 本项目约定使用 `configs/.env` 提供应用所需环境变量，避免根 `.env` 干扰。
- 如果你的根目录存在 `.env`，请确保也包含 `REDIS_PASSWORD`，或临时重命名为 `.env.bak` 以避免编排期将其置空。
- Redis 在 Compose 中通过命令行参数设置密码，我们已使用 `$$REDIS_PASSWORD` 让变量在“容器内”展开，规避解析期替换。

快速校验与重建：
```bash
# 确保在文件 configs/.env 中设置了 REDIS_PASSWORD
# 例如：REDIS_PASSWORD=your-redis-password

docker-compose -f docker-compose.multi-local.yml down
docker-compose -f docker-compose.multi-local.yml up -d --force-recreate
docker-compose -f docker-compose.multi-local.yml logs -f redis
```

### 4. 访问应用

服务启动后：

- **应用**: `http://localhost:8080`
- **API 文档**: `http://localhost:8080/swagger/index.html`

### 5. 查看日志和停止服务

```bash
# 查看所有服务的实时日志 (使用对应的 -f 文件)
docker-compose -f docker-compose.multi-local.yml logs -f

# 停止并移除所有容器、网络和卷
docker-compose -f docker-compose.multi-local.yml down
```

### 6. 常见问题（FAQ）

- **看到警告 The "REDIS_PASSWORD" variable is not set**：
  - 说明 Compose 解析期没有拿到该变量。请确认根 `.env` 不干扰，且 `configs/.env` 中已设置 `REDIS_PASSWORD`。
  - 我们已在 Compose 中使用 `$$REDIS_PASSWORD`，变量会在容器内展开。只要 `configs/.env` 有值，Redis 会正确启用密码。
- **Redis 日志出现 requirepass wrong number of arguments**：
  - 通常是密码为空导致。按上面步骤“校验与重建”，确保 `REDIS_PASSWORD` 有值后 `--force-recreate` 重启。
- **Compose 提示 version 字段 obsolete**：
  - 该提示可忽略，不影响运行；也可自行移除 compose 文件中的 `version:` 以消除提示。

---

## 快速开始 (本地手动部署)

此方法适用于不使用 Docker，希望在本地手动配置和运行所有依赖的开发者。

1.  **克隆项目**
    ```bash
    git clone https://github.com/yuchen1204/backend_go
    cd backend
    ```

2.  **安装依赖**
    ```bash
    go mod tidy
    ```

3.  **生成API文档**
    ```bash
    # (在 Linux/macOS)
    chmod +x scripts/generate-docs.sh
    ./scripts/generate-docs.sh
    ```

4.  **设置环境变量**
    复制 `configs/env.example` 文件到项目根目录，并重命名为 `.env`。
    ```bash
    cp configs/env.example .env
    ```
    然后编辑 `.env` 文件，至少需要配置好数据库、Redis和SMTP服务的连接信息。

5.  **生成 vendor 依赖**（项目使用 `-mod=vendor`）
    ```bash
    go mod tidy
    go mod vendor
    ```

6.  **启动 PostgreSQL 数据库和 Redis**
    你需要在本地手动安装并启动 PostgreSQL 和 Redis 服务，并确保已创建好应用所需的数据库。
    ```bash
    # 示例: 在 Ubuntu 上安装
    # sudo apt-get install postgresql postgresql-contrib redis-server

    # 创建数据库
    createdb backend
    ```

7.  **运行应用**
    ```bash
    go run cmd/main.go
    ```

8.  **访问API文档**
    浏览器访问 `http://localhost:8080/swagger/index.html`。

> 何时需要重新运行 `go mod vendor`？
> - 新增/升级/移除依赖后。
> - CI/CD 或 Docker 构建若在 `RUN go install -mod=vendor ...` 或 `go build -mod=vendor` 时报依赖缺失。

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

### 🔓 认证相关接口（公开访问）

#### 用户注册流程

##### 1. 发送注册验证码
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

##### 2. 用户注册
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

#### 用户登录流程

##### 3. 用户登录
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

#### Token管理

##### 4. 刷新访问Token
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

##### 5. 用户登出
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

#### 密码重置流程

##### 6. 发送重置密码验证码
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

##### 7. 重置密码
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

### 🔍 用户信息查询接口（公开访问）

#### 8. 根据ID获取用户
- **GET** `/api/v1/users/{id}`
- **描述**: 通过用户ID获取用户详细信息

#### 9. 根据用户名获取用户
- **GET** `/api/v1/users/username/{username}`
- **描述**: 通过用户名获取用户详细信息

### 🔒 用户个人信息管理接口（需要认证）

#### 10. 获取当前用户信息
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

#### 11. 更新当前用户信息
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

### 📁 文件管理接口（需要认证）

#### 13. 上传单个文件
- **POST** `/files/upload`
- **描述**: 上传单个文件到指定的存储位置（本地或S3）。支持自定义存储配置、文件分类和访问权限设置。
- **认证**: `Bearer Token` (仅接受Access Token)
- **内容类型**: `multipart/form-data`

**请求头示例:**
```
Authorization: Bearer <your-access-token>
```

**请求参数:**
- `file` (formData, file, 必填): 要上传的文件
- `storage_name` (formData, string, 可选): 存储名称（默认使用系统默认存储）
- `category` (formData, string, 可选): 文件分类
- `description` (formData, string, 可选): 文件描述
- `is_public` (formData, boolean, 可选): 是否公开访问（默认false）

**响应示例:**
```json
{
    "code": 201,
    "message": "文件上传成功",
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "filename": "example.jpg",
        "original_name": "照片.jpg",
        "file_size": 1024000,
        "mime_type": "image/jpeg",
        "url": "https://your-domain.com/uploads/2024/01/550e8400-e29b-41d4-a716-446655440000.jpg",
        "category": "avatar",
        "description": "用户头像",
        "is_public": true,
        "storage_type": "local",
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-01T12:00:00Z"
    },
    "timestamp": 1640995200
}
```

#### 14. 上传多个文件
- **POST** `/files/upload-multiple`
- **描述**: 批量上传多个文件到指定的存储位置，支持相同的配置参数。
- **认证**: `Bearer Token` (仅接受Access Token)
- **内容类型**: `multipart/form-data`

**请求参数:**
- `files` (formData, file[], 必填): 要上传的文件列表
- `storage_name` (formData, string, 可选): 存储名称
- `category` (formData, string, 可选): 文件分类
- `description` (formData, string, 可选): 文件描述
- `is_public` (formData, boolean, 可选): 是否公开访问

**响应示例:**
```json
{
    "code": 201,
    "message": "文件上传成功",
    "data": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440001",
            "filename": "doc1.pdf",
            "original_name": "文档1.pdf",
            "file_size": 2048000,
            "mime_type": "application/pdf",
            "url": "https://your-domain.com/uploads/2024/01/550e8400-e29b-41d4-a716-446655440001.pdf",
            "category": "document",
            "is_public": false,
            "storage_type": "s3",
            "created_at": "2024-01-01T12:00:00Z"
        },
        {
            "id": "550e8400-e29b-41d4-a716-446655440002",
            "filename": "image2.png",
            "original_name": "图片2.png",
            "file_size": 512000,
            "mime_type": "image/png",
            "url": "https://your-domain.com/uploads/2024/01/550e8400-e29b-41d4-a716-446655440002.png",
            "category": "gallery",
            "is_public": true,
            "storage_type": "s3",
            "created_at": "2024-01-01T12:00:00Z"
        }
    ],
    "timestamp": 1640995200
}
```

#### 15. 获取文件详情
- **GET** `/files/{id}`
- **描述**: 根据文件ID获取文件详细信息。支持公开文件无需认证访问，私有文件需要认证。
- **认证**: 可选（公开文件无需认证，私有文件需要Bearer Token）

**路径参数:**
- `id` (path, string, 必填): 文件UUID

**响应示例:**
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "filename": "example.jpg",
        "original_name": "照片.jpg",
        "file_size": 1024000,
        "mime_type": "image/jpeg",
        "url": "https://your-domain.com/uploads/2024/01/550e8400-e29b-41d4-a716-446655440000.jpg",
        "category": "avatar",
        "description": "用户头像",
        "is_public": true,
        "storage_type": "local",
        "owner": {
            "id": "123e4567-e89b-12d3-a456-426614174000",
            "username": "testuser",
            "nickname": "测试用户"
        },
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-01T12:00:00Z"
    },
    "timestamp": 1640995200
}
```

### ⚙️ 系统接口

#### 16. 健康检查
- **GET** `/health`
- **描述**: 服务健康状态检查，用于监控系统运行状态

**响应示例:**
```json
{
    "code": 200,
    "message": "服务运行正常",
    "data": {
        "status": "healthy",
        "timestamp": "2024-01-01T12:00:00Z",
        "version": "v1.0.0"
    },
    "timestamp": 1640995200
}
```

## 环境配置

复制 `configs/env.example` 文件并根据需要修改配置：

```bash
# 服务器配置
PORT=8080

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-postgres-password
DB_NAME=backend
DB_SSLMODE=disable

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password
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
