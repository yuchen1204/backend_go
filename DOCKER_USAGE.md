# Docker 部署使用说明

这个Dockerfile创建了一个包含所有服务的单容器解决方案，包括PostgreSQL、Redis、SSH服务和Go后端应用。

## 包含的服务

- **PostgreSQL**: 数据库服务 (内部端口5432，无外部映射)
- **Redis**: 缓存服务 (内部端口6379，无外部映射)  
- **SSH服务**: 远程访问 (映射到宿主机端口1000)
- **Go后端应用**: API服务 (映射到宿主机端口1001)

## 构建和运行

### 方式一：使用Docker命令

```bash
# 构建镜像
docker build -t backend-app .

# 运行容器
docker run -d \
  --name backend-all-services \
  -p 1000:1000 \
  -p 1001:1001 \
  -v postgres_data:/var/lib/postgresql \
  -v redis_data:/var/lib/redis \
  -v uploads_data:/app/uploads \
  backend-app
```

### 方式二：使用Docker Compose（推荐）

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重新构建
docker-compose up --build -d
```

## 访问服务

- **API服务**: http://localhost:1001
- **API文档**: http://localhost:1001/swagger/index.html
- **SSH访问**: ssh root@localhost -p 1000 (密码: password)

## 环境配置

默认配置已包含在Dockerfile中，主要配置项：

- 后端端口: 1001
- PostgreSQL: localhost:5432 (postgres/postgres)
- Redis: localhost:6379
- SSH端口: 1000
- JWT密钥: 建议在生产环境中修改

## 数据持久化

使用了Docker卷来持久化数据：
- `postgres_data`: PostgreSQL数据
- `redis_data`: Redis数据  
- `uploads_data`: 文件上传数据

## 日志查看

```bash
# 查看所有服务日志
docker exec -it backend-all-services supervisorctl status

# 查看特定服务日志
docker exec -it backend-all-services tail -f /var/log/backend.log
docker exec -it backend-all-services tail -f /var/log/postgresql.log
docker exec -it backend-all-services tail -f /var/log/redis.log
docker exec -it backend-all-services tail -f /var/log/sshd.log
```

## 进入容器

```bash
# 进入容器Shell
docker exec -it backend-all-services bash

# 或通过SSH
ssh root@localhost -p 1000
```

## 健康检查

容器包含健康检查功能，自动检测API服务状态：

```bash
# 查看健康状态
docker ps
```

## 生产环境建议

1. **修改SSH密码**: 
   ```bash
   docker exec -it backend-all-services passwd root
   ```

2. **修改JWT密钥**: 更新环境变量或.env文件中的JWT_SECRET

3. **配置SMTP**: 如需邮件功能，配置SMTP相关环境变量

4. **备份数据**: 定期备份Docker卷数据

## 故障排查

1. 如果服务启动失败，检查日志：
   ```bash
   docker logs backend-all-services
   ```

2. 如果数据库连接失败，确认PostgreSQL已启动：
   ```bash
   docker exec -it backend-all-services supervisorctl status postgresql
   ```

3. 如果API无法访问，检查端口映射和防火墙设置

## 停止和清理

```bash
# 停止容器
docker-compose down

# 删除所有数据（谨慎使用）
docker-compose down -v

# 删除镜像
docker rmi backend-app
``` 