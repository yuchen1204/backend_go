# 🐳 Docker 多服务容器部署指南

本项目提供了完整的Docker解决方案，支持在单个容器中运行SSH、PostgreSQL、Redis和Go应用程序。

## 🏗️ 架构说明

### 服务组件
- **SSH服务**: 提供远程访问能力
- **PostgreSQL**: 主数据库
- **Redis**: 缓存和会话存储
- **Go应用**: Backend API服务
- **Supervisor**: 进程管理器，管理所有服务

### 端口映射
| 服务 | 容器端口 | 主机端口 | 说明 |
|------|----------|----------|------|
| Go应用 | 8080 | 8080 | Web API服务 |
| SSH | 22 | 2222 | SSH远程访问 |
| PostgreSQL | 5432 | 5432 | 数据库服务 |
| Redis | 6379 | 6379 | 缓存服务 |

## 🚀 快速开始

### 方法1: 使用脚本（推荐）

```bash
# 1. 构建Docker镜像
./scripts/docker-build.sh

# 2. 启动容器
./scripts/docker-run.sh

# 3. 查看日志
./scripts/docker-logs.sh
```

### 方法2: 使用Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 方法3: 手动Docker命令

```bash
# 构建镜像
docker build -t backend-app:latest .

# 运行容器
docker run -d \
  --name backend-container \
  -p 8080:8080 \
  -p 2222:22 \
  -p 5432:5432 \
  -p 6379:6379 \
  -v $(pwd)/uploads:/app/uploads \
  backend-app:latest
```

## 🔧 服务访问

### 🌐 Web应用
- **API服务**: http://localhost:8080
- **Swagger文档**: http://localhost:8080/swagger/index.html
- **健康检查**: http://localhost:8080/health

### 🔐 SSH访问
```bash
# SSH连接到容器
ssh root@localhost -p 2222
# 密码: root
```

### 🗄️ 数据库连接
```bash
# PostgreSQL连接
psql -h localhost -p 5432 -U postgres -d backend
# 密码: postgres

# 或在容器内
docker exec -it backend-container psql -U postgres -d backend
```

### 🚀 Redis连接
```bash
# Redis CLI连接
redis-cli -h localhost -p 6379

# 或在容器内
docker exec -it backend-container redis-cli
```

## 📊 监控和管理

### 查看服务状态
```bash
# 使用脚本
./scripts/docker-logs.sh

# 直接查看supervisor状态
docker exec backend-container supervisorctl status

# 查看容器日志
docker logs backend-container
```

### 重启服务
```bash
# 重启单个服务
docker exec backend-container supervisorctl restart backend
docker exec backend-container supervisorctl restart postgresql
docker exec backend-container supervisorctl restart redis

# 重启所有服务
docker exec backend-container supervisorctl restart all
```

### 进入容器调试
```bash
# 进入容器
docker exec -it backend-container bash

# 查看进程
docker exec backend-container ps aux

# 查看网络
docker exec backend-container netstat -tlnp
```

## 🔧 配置管理

### 环境变量配置
容器会自动从以下位置读取配置：
1. `configs/env.example` (复制为`.env`)
2. Docker环境变量
3. 容器内默认配置

### 关键配置项
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=backend

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=your-super-secret-key
```

## 📁 数据持久化

### 数据卷挂载
- `./uploads` → `/app/uploads` (文件上传目录)
- `./configs` → `/app/configs` (配置文件)
- `backend_data` → `/var/lib/postgresql/13/main` (数据库数据)

### 备份数据
```bash
# 备份数据库
docker exec backend-container pg_dump -U postgres backend > backup.sql

# 恢复数据库
docker exec -i backend-container psql -U postgres backend < backup.sql

# 备份Redis
docker exec backend-container redis-cli SAVE
docker cp backend-container:/var/lib/redis/dump.rdb ./redis-backup.rdb
```

## 🛠️ 开发模式

### 代码热重载
```bash
# 挂载源代码目录
docker run -d \
  --name backend-dev \
  -p 8080:8080 \
  -v $(pwd):/app \
  backend-app:latest
```

### 调试模式
```bash
# 以交互模式运行
docker run -it \
  --rm \
  -p 8080:8080 \
  backend-app:latest \
  bash
```

## 🚨 故障排除

### 常见问题

#### 1. 容器启动失败
```bash
# 查看构建日志
docker build -t backend-app:latest . --no-cache

# 查看启动日志
docker logs backend-container
```

#### 2. 服务无法连接
```bash
# 检查端口占用
netstat -tlnp | grep -E "(8080|2222|5432|6379)"

# 检查防火墙
sudo ufw status
```

#### 3. 数据库连接失败
```bash
# 检查PostgreSQL服务
docker exec backend-container supervisorctl status postgresql

# 手动启动PostgreSQL
docker exec backend-container supervisorctl start postgresql
```

#### 4. 权限问题
```bash
# 修复文件权限
sudo chown -R $(whoami):$(whoami) uploads/
chmod -R 755 uploads/
```

### 日志位置
- **应用日志**: `/var/log/backend.log`
- **PostgreSQL日志**: `/var/log/postgresql.log`
- **Redis日志**: `/var/log/redis.log`
- **SSH日志**: `/var/log/sshd.log`
- **Supervisor日志**: `/var/log/supervisor/supervisord.log`

## 🔒 安全配置

### 生产环境建议
1. **修改默认密码**
   ```bash
   # SSH root密码
   docker exec backend-container passwd root
   
   # PostgreSQL密码
   docker exec backend-container -u postgres psql -c "ALTER USER postgres PASSWORD 'new-password';"
   ```

2. **限制网络访问**
   ```bash
   # 只绑定本地接口
   docker run -p 127.0.0.1:8080:8080 ...
   ```

3. **使用非root用户**
   - 在Dockerfile中创建专用用户
   - 配置适当的文件权限

4. **更新JWT密钥**
   ```bash
   export JWT_SECRET="your-very-secure-random-key-here"
   ```

## 📈 性能优化

### 资源限制
```bash
# 限制内存和CPU
docker run -d \
  --memory=1g \
  --cpus=1.0 \
  --name backend-container \
  backend-app:latest
```

### 数据库优化
```bash
# 调整PostgreSQL配置
docker exec backend-container \
  sed -i 's/#shared_buffers = 128MB/shared_buffers = 256MB/' \
  /etc/postgresql/13/main/postgresql.conf
```

---

## 🆘 获取帮助

如果遇到问题，请：
1. 查看日志文件
2. 检查服务状态
3. 验证网络连接
4. 确认配置文件正确

更多信息请参考项目主README文档。 