# Backend Go API 文档

## 项目概述

这是一个基于 Go 语言和 Gin 框架构建的企业级用户认证与文件管理系统后端 API。

### 技术栈
- **语言**: Go 1.24+
- **框架**: Gin 1.10.0
- **数据库**: PostgreSQL + GORM
- **缓存**: Redis
- **认证**: JWT (Access Token + Refresh Token)
- **文档**: Swagger/OpenAPI 3.0
- **存储**: 本地存储 + AWS S3

### 基础信息
- **Base URL**: `http://localhost:8080/api/v1`
- **Swagger 文档**: `http://localhost:8080/swagger/index.html`
- **健康检查**: `GET /health`

## 认证机制

### JWT 双 Token 机制
- **Access Token**: 30分钟有效期，用于API访问
- **Refresh Token**: 7天有效期，用于刷新Access Token
- **Token 黑名单**: 登出时立即失效所有Token

### 认证头格式
```
Authorization: Bearer <access_token>
```

### 设备验证
- 陌生设备登录需要邮箱验证码
- 基于设备指纹识别（IP + User-Agent）

## API 接口详情

## 1. 用户管理 API

### 1.1 发送注册验证码
**POST** `/users/send-code`

发送用于注册的6位数验证码到指定邮箱（5分钟有效）。

**请求体**:
```json
{
  "username": "testuser",
  "email": "test@example.com"
}
```

**响应**:
- `200`: 验证码发送成功
- `400`: 请求参数错误
- `409`: 用户名或邮箱已被注册
- `429`: 请求过于频繁
- `500`: 服务器内部错误

### 1.2 用户注册
**POST** `/users/register`

使用邮箱验证码创建新用户账户。

**请求体**:
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123",
  "verification_code": "123456"
}
```

**响应**:
- `201`: 注册成功，返回用户信息
- `400`: 请求参数错误或验证码错误
- `409`: 用户名或邮箱已存在
- `500`: 服务器内部错误

### 1.3 用户登录
**POST** `/users/login`

用户登录，支持陌生设备邮箱验证。

**请求体**:
```json
{
  "username": "testuser",
  "password": "password123",
  "device_verification_code": "123456"  // 陌生设备需要
}
```

**响应**:
- `200`: 登录成功，返回Token和用户信息
- `400`: 请求参数错误或设备验证码相关错误
- `401`: 用户名或密码错误
- `403`: 账户已被封禁或未激活
- `500`: 服务器内部错误

**成功响应示例**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "uuid",
      "username": "testuser",
      "email": "test@example.com",
      "nickname": "Test User",
      "avatar": "http://example.com/avatar.jpg",
      "status": "active"
    }
  }
}
```

### 1.4 刷新Token
**POST** `/users/refresh`

使用Refresh Token获取新的Access Token。

**请求体**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应**:
- `200`: 刷新成功，返回新的Access Token
- `400`: 请求参数错误
- `401`: Refresh Token无效或已过期
- `403`: 账户已被封禁或未激活
- `500`: 服务器内部错误

### 1.5 用户登出
**POST** `/users/logout`

登出用户并撤销所有Token。

**请求体**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应**:
- `200`: 登出成功
- `400`: 请求参数错误
- `401`: Token无效
- `500`: 服务器内部错误

### 1.6 获取当前用户信息
**GET** `/users/me` 🔒

获取当前登录用户的详细信息。

**Headers**: `Authorization: Bearer <access_token>`

**响应**:
- `200`: 获取成功，返回用户信息
- `401`: 未授权或Token无效
- `500`: 服务器内部错误

### 1.7 更新用户信息
**PUT** `/users/me` 🔒

更新当前登录用户的基本信息。

**Headers**: `Authorization: Bearer <access_token>`

**请求体**:
```json
{
  "nickname": "新昵称",
  "bio": "个人简介",
  "avatar": "http://example.com/new-avatar.jpg"
}
```

**响应**:
- `200`: 更新成功，返回更新后的用户信息
- `400`: 请求参数错误
- `401`: 未授权或Token无效
- `404`: 用户不存在
- `500`: 服务器内部错误

### 1.8 根据ID获取用户信息
**GET** `/users/{id}`

通过用户ID获取用户详细信息。

**路径参数**:
- `id`: 用户UUID

**响应**:
- `200`: 获取成功，返回用户信息
- `400`: 请求参数错误
- `404`: 用户不存在
- `500`: 服务器内部错误

### 1.9 根据用户名获取用户信息
**GET** `/users/username/{username}`

通过用户名获取用户详细信息。

**路径参数**:
- `username`: 用户名

**响应**:
- `200`: 获取成功，返回用户信息
- `404`: 用户不存在
- `500`: 服务器内部错误

## 2. 密码管理 API

### 2.1 发送重置密码验证码
**POST** `/users/send-reset-code`

向指定邮箱发送用于重置密码的验证码。

**请求体**:
```json
{
  "email": "test@example.com"
}
```

**响应**:
- `200`: 验证码发送成功
- `400`: 请求参数错误
- `429`: 请求过于频繁
- `500`: 服务器内部错误

### 2.2 重置密码
**POST** `/users/reset-password`

使用验证码重置密码。

**请求体**:
```json
{
  "email": "test@example.com",
  "verification_code": "123456",
  "new_password": "newpassword123"
}
```

**响应**:
- `200`: 密码重置成功
- `400`: 请求参数错误或验证码错误
- `500`: 服务器内部错误

## 3. 账户激活 API

### 3.1 发送激活验证码
**POST** `/users/send-activation-code`

为非活跃账户发送激活验证码。

**请求体**:
```json
{
  "email": "test@example.com"
}
```

**响应**:
- `200`: 验证码发送成功
- `400`: 请求参数错误
- `429`: 请求过于频繁
- `500`: 服务器内部错误

### 3.2 激活账户
**POST** `/users/activate`

使用验证码激活非活跃账户。

**请求体**:
```json
{
  "email": "test@example.com",
  "verification_code": "123456"
}
```

**响应**:
- `200`: 账户激活成功
- `400`: 请求参数错误或验证码错误
- `500`: 服务器内部错误

## 4. 文件管理 API

### 4.1 上传单个文件
**POST** `/files/upload` 🔒

上传单个文件到指定的存储位置。

**Headers**: `Authorization: Bearer <access_token>`

**请求体** (multipart/form-data):
- `file`: 要上传的文件
- `storage_name`: 存储名称（可选）
- `category`: 文件分类（可选）
- `description`: 文件描述（可选）
- `is_public`: 是否公开访问（可选）

**响应**:
- `201`: 上传成功，返回文件信息
- `400`: 请求参数错误
- `401`: 未授权
- `413`: 文件过大
- `500`: 服务器内部错误

### 4.2 批量上传文件
**POST** `/files/upload-multiple` 🔒

批量上传多个文件。

**Headers**: `Authorization: Bearer <access_token>`

**请求体** (multipart/form-data):
- `files`: 要上传的文件列表
- `storage_name`: 存储名称（可选）
- `category`: 文件分类（可选）
- `description`: 文件描述（可选）
- `is_public`: 是否公开访问（可选）

**响应**:
- `201`: 上传成功，返回文件列表
- `400`: 请求参数错误
- `401`: 未授权
- `413`: 文件过大
- `500`: 服务器内部错误

### 4.3 获取文件详情
**GET** `/files/{id}`

根据文件ID获取文件详细信息。

**路径参数**:
- `id`: 文件UUID

**响应**:
- `200`: 获取成功，返回文件信息
- `400`: 请求参数错误
- `403`: 访问被拒绝
- `404`: 文件不存在

### 4.4 获取用户文件列表
**GET** `/files/my` 🔒

获取当前登录用户的文件列表。

**Headers**: `Authorization: Bearer <access_token>`

**查询参数**:
- `category`: 文件分类筛选
- `storage_type`: 存储类型筛选
- `storage_name`: 存储名称筛选
- `is_public`: 是否公开筛选
- `page`: 页码（默认1）
- `page_size`: 每页大小（默认20）

**响应**:
- `200`: 获取成功，返回文件列表
- `401`: 未授权
- `500`: 服务器内部错误

### 4.5 获取公开文件列表
**GET** `/files/public`

获取所有公开访问的文件列表。

**查询参数**:
- `category`: 文件分类筛选
- `storage_type`: 存储类型筛选
- `storage_name`: 存储名称筛选
- `page`: 页码（默认1）
- `page_size`: 每页大小（默认20）

**响应**:
- `200`: 获取成功，返回文件列表
- `500`: 服务器内部错误

### 4.6 更新文件信息
**PUT** `/files/{id}` 🔒

更新文件的分类、描述等信息（仅文件所有者可操作）。

**Headers**: `Authorization: Bearer <access_token>`

**路径参数**:
- `id`: 文件UUID

**请求体**:
```json
{
  "category": "新分类",
  "description": "新描述",
  "is_public": true
}
```

**响应**:
- `200`: 更新成功，返回更新后的文件信息
- `400`: 请求参数错误
- `401`: 未授权
- `403`: 访问被拒绝
- `404`: 文件不存在

### 4.7 删除文件
**DELETE** `/files/{id}` 🔒

删除指定的文件（仅文件所有者可操作）。

**Headers**: `Authorization: Bearer <access_token>`

**路径参数**:
- `id`: 文件UUID

**响应**:
- `200`: 删除成功
- `400`: 请求参数错误
- `401`: 未授权
- `403`: 访问被拒绝
- `404`: 文件不存在

### 4.8 获取存储信息
**GET** `/files/storages`

获取系统可用的存储配置信息。

**响应**:
- `200`: 获取成功，返回存储配置信息
- `500`: 服务器内部错误

## 5. 管理员 API

### 5.1 管理员登录
**POST** `/admin/login`

管理员登录获取管理员Token。

**请求体**:
```json
{
  "username": "admin",
  "password": "admin"
}
```

**响应**:
- `200`: 登录成功，返回管理员Token
- `400`: 请求参数错误
- `401`: 用户名或密码错误
- `500`: 服务器内部错误

### 5.2 管理员面板
**GET** `/admin/dashboard` 🔒👑

访问管理员面板。

**Headers**: `Authorization: Bearer <admin_token>`

**响应**:
- `200`: 访问成功
- `401`: 未授权
- `403`: 权限不足

### 5.3 获取用户列表
**GET** `/admin/users` 🔒👑

获取系统中的用户列表。

**Headers**: `Authorization: Bearer <admin_token>`

**查询参数**:
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10）
- `search`: 搜索关键词

**响应**:
- `200`: 获取成功，返回用户列表和分页信息
- `401`: 未授权
- `403`: 权限不足
- `500`: 服务器内部错误

### 5.4 获取用户详情
**GET** `/admin/users/{id}` 🔒👑

获取指定用户的详细信息。

**Headers**: `Authorization: Bearer <admin_token>`

**路径参数**:
- `id`: 用户UUID

**响应**:
- `200`: 获取成功，返回用户详情
- `400`: 请求参数错误
- `401`: 未授权
- `403`: 权限不足
- `404`: 用户不存在

### 5.5 更新用户状态
**PUT** `/admin/users/{id}/status` 🔒👑

更新用户的状态（激活/禁用/封禁）。

**Headers**: `Authorization: Bearer <admin_token>`

**路径参数**:
- `id`: 用户UUID

**请求体**:
```json
{
  "status": "active"  // active, inactive, banned
}
```

**响应**:
- `200`: 更新成功
- `400`: 请求参数错误
- `401`: 未授权
- `403`: 权限不足
- `500`: 服务器内部错误

### 5.6 删除用户
**DELETE** `/admin/users/{id}` 🔒👑

删除指定用户。

**Headers**: `Authorization: Bearer <admin_token>`

**路径参数**:
- `id`: 用户UUID

**响应**:
- `200`: 删除成功
- `400`: 请求参数错误
- `401`: 未授权
- `403`: 权限不足
- `500`: 服务器内部错误

### 5.7 获取用户统计
**GET** `/admin/stats/users` 🔒👑

获取用户统计信息。

**Headers**: `Authorization: Bearer <admin_token>`

**响应**:
- `200`: 获取成功，返回统计信息
- `401`: 未授权
- `403`: 权限不足
- `500`: 服务器内部错误

## 6. 系统 API

### 6.1 健康检查
**GET** `/health`

检查服务健康状态。

**响应**:
```json
{
  "code": 200,
  "message": "服务正常",
  "data": {
    "status": "ok",
    "service": "backend"
  }
}
```

### 6.2 静态文件访问
**GET** `/uploads/{path}`

访问上传的静态文件。

**路径参数**:
- `path`: 文件路径

## 错误码说明

### HTTP 状态码
- `200`: 成功
- `201`: 创建成功
- `400`: 请求参数错误
- `401`: 未授权
- `403`: 权限不足
- `404`: 资源不存在
- `409`: 资源冲突
- `413`: 请求实体过大
- `429`: 请求过于频繁
- `500`: 服务器内部错误

### 响应格式
所有API响应都遵循统一格式：

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {}  // 可选，具体数据
}
```

错误响应格式：
```json
{
  "code": 400,
  "message": "请求参数错误",
  "error": "具体错误信息"  // 可选
}
```

## 安全特性

### 1. 密码安全
- 密码使用加盐哈希存储（SHA256）
- 支持密码强度验证

### 2. Token 安全
- JWT双Token机制
- Token黑名单机制
- 自动过期处理

### 3. 设备验证
- 基于设备指纹的陌生设备识别
- 邮箱二次验证

### 4. 频率限制
- IP请求频率限制
- 验证码发送频率限制

### 5. 跨域配置
- CORS中间件配置
- 安全的跨域访问控制

## 部署说明

### 环境变量配置
参考 `env.example` 文件配置以下环境变量：

- **数据库配置**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- **Redis配置**: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- **SMTP配置**: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`
- **JWT配置**: `JWT_SECRET`, `JWT_ACCESS_TOKEN_EXPIRES_IN_MINUTES`, `JWT_REFRESH_TOKEN_EXPIRES_IN_DAYS`
- **管理员配置**: `PANEL_USER`, `PANEL_PASSWORD`
- **文件存储配置**: `FILE_STORAGE_DEFAULT`, `FILE_STORAGE_LOCAL_NAMES`, `FILE_STORAGE_S3_NAMES`

### Docker 部署
```bash
# 构建并启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f backend
```

### 本地开发
```bash
# 安装依赖
go mod download

# 生成API文档
swag init -g cmd/main.go -o ./docs

# 运行服务
go run cmd/main.go
```

## 图标说明
- 🔒: 需要用户认证
- 👑: 需要管理员权限

---

**文档版本**: v1.0  
**最后更新**: 2025-08-21  
**联系方式**: 请通过项目仓库提交Issue
