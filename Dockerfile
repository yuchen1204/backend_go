# 多阶段构建，基于Ubuntu 22.04
FROM ubuntu:22.04

# 设置环境变量
ENV DEBIAN_FRONTEND=noninteractive
ENV GO_VERSION=1.24.5
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# 创建工作目录
WORKDIR /app

# 安装基础工具和依赖
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    supervisor \
    openssh-server \
    postgresql \
    postgresql-contrib \
    redis-server \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# 安装Go语言
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

# 配置PostgreSQL
RUN service postgresql start && \
    sudo -u postgres psql -c "CREATE USER postgres WITH PASSWORD 'postgres';" && \
    sudo -u postgres psql -c "ALTER USER postgres CREATEDB;" && \
    sudo -u postgres createdb backend && \
    service postgresql stop

# 配置Redis
RUN sed -i 's/bind 127.0.0.1 ::1/bind 0.0.0.0/' /etc/redis/redis.conf && \
    sed -i 's/protected-mode yes/protected-mode no/' /etc/redis/redis.conf

# 配置SSH
RUN mkdir /var/run/sshd && \
    echo 'root:password' | chpasswd && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#Port 22/Port 1000/' /etc/ssh/sshd_config && \
    ssh-keygen -A

# 复制项目文件
COPY . /app/

# 设置Go模块代理（加速下载）
ENV GOPROXY=https://goproxy.cn,direct

# 构建Go应用
RUN go mod tidy && \
    go mod download && \
    go build -o backend cmd/main.go

# 创建上传目录
RUN mkdir -p uploads/default uploads/avatar uploads/document

# 创建环境配置文件
RUN echo "# 服务器配置" > .env && \
    echo "PORT=1001" >> .env && \
    echo "" >> .env && \
    echo "# 数据库配置" >> .env && \
    echo "DB_HOST=localhost" >> .env && \
    echo "DB_PORT=5432" >> .env && \
    echo "DB_USER=postgres" >> .env && \
    echo "DB_PASSWORD=postgres" >> .env && \
    echo "DB_NAME=backend" >> .env && \
    echo "DB_SSLMODE=disable" >> .env && \
    echo "" >> .env && \
    echo "# Redis 配置" >> .env && \
    echo "REDIS_HOST=localhost" >> .env && \
    echo "REDIS_PORT=6379" >> .env && \
    echo "REDIS_PASSWORD=" >> .env && \
    echo "REDIS_DB=0" >> .env && \
    echo "" >> .env && \
    echo "# SMTP 邮件服务配置（可选）" >> .env && \
    echo "SMTP_HOST=smtp.example.com" >> .env && \
    echo "SMTP_PORT=587" >> .env && \
    echo "SMTP_USERNAME=your-email@example.com" >> .env && \
    echo "SMTP_PASSWORD=your-email-password" >> .env && \
    echo "SMTP_FROM=your-email@example.com" >> .env && \
    echo "" >> .env && \
    echo "# 安全配置" >> .env && \
    echo "MAX_IP_REQUESTS_PER_DAY=1000" >> .env && \
    echo "" >> .env && \
    echo "# JWT 配置" >> .env && \
    echo "JWT_SECRET=your-very-secret-jwt-key-change-this-in-production" >> .env && \
    echo "JWT_ACCESS_TOKEN_EXPIRES_IN_MINUTES=30" >> .env && \
    echo "JWT_REFRESH_TOKEN_EXPIRES_IN_DAYS=7" >> .env && \
    echo "" >> .env && \
    echo "# 文件存储配置" >> .env && \
    echo "FILE_STORAGE_DEFAULT=local_default" >> .env && \
    echo "" >> .env && \
    echo "# 本地存储配置" >> .env && \
    echo "FILE_STORAGE_LOCAL_NAMES=default,avatar,document" >> .env && \
    echo "FILE_STORAGE_LOCAL_DEFAULT_PATH=./uploads/default" >> .env && \
    echo "FILE_STORAGE_LOCAL_DEFAULT_URL=http://localhost:1001/uploads/default" >> .env && \
    echo "FILE_STORAGE_LOCAL_AVATAR_PATH=./uploads/avatar" >> .env && \
    echo "FILE_STORAGE_LOCAL_AVATAR_URL=http://localhost:1001/uploads/avatar" >> .env && \
    echo "FILE_STORAGE_LOCAL_DOCUMENT_PATH=./uploads/document" >> .env && \
    echo "FILE_STORAGE_LOCAL_DOCUMENT_URL=http://localhost:1001/uploads/document" >> .env

# 创建supervisor配置文件
RUN echo "[supervisord]" > /etc/supervisor/conf.d/services.conf && \
    echo "nodaemon=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "user=root" >> /etc/supervisor/conf.d/services.conf && \
    echo "" >> /etc/supervisor/conf.d/services.conf && \
    echo "[program:postgresql]" >> /etc/supervisor/conf.d/services.conf && \
    echo "command=/usr/lib/postgresql/14/bin/postgres -D /var/lib/postgresql/14/main -c config_file=/etc/postgresql/14/main/postgresql.conf" >> /etc/supervisor/conf.d/services.conf && \
    echo "user=postgres" >> /etc/supervisor/conf.d/services.conf && \
    echo "autostart=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "autorestart=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "redirect_stderr=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "stdout_logfile=/var/log/postgresql.log" >> /etc/supervisor/conf.d/services.conf && \
    echo "" >> /etc/supervisor/conf.d/services.conf && \
    echo "[program:redis]" >> /etc/supervisor/conf.d/services.conf && \
    echo "command=redis-server /etc/redis/redis.conf" >> /etc/supervisor/conf.d/services.conf && \
    echo "autostart=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "autorestart=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "redirect_stderr=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "stdout_logfile=/var/log/redis.log" >> /etc/supervisor/conf.d/services.conf && \
    echo "" >> /etc/supervisor/conf.d/services.conf && \
    echo "[program:sshd]" >> /etc/supervisor/conf.d/services.conf && \
    echo "command=/usr/sbin/sshd -D -p 1000" >> /etc/supervisor/conf.d/services.conf && \
    echo "autostart=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "autorestart=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "redirect_stderr=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "stdout_logfile=/var/log/sshd.log" >> /etc/supervisor/conf.d/services.conf && \
    echo "" >> /etc/supervisor/conf.d/services.conf && \
    echo "[program:backend]" >> /etc/supervisor/conf.d/services.conf && \
    echo "command=/app/backend" >> /etc/supervisor/conf.d/services.conf && \
    echo "directory=/app" >> /etc/supervisor/conf.d/services.conf && \
    echo "autostart=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "autorestart=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "redirect_stderr=true" >> /etc/supervisor/conf.d/services.conf && \
    echo "stdout_logfile=/var/log/backend.log" >> /etc/supervisor/conf.d/services.conf && \
    echo "environment=PATH=\"/usr/local/go/bin:%(ENV_PATH)s\"" >> /etc/supervisor/conf.d/services.conf

# 创建启动脚本
RUN echo "#!/bin/bash" > /start.sh && \
    echo "" >> /start.sh && \
    echo "# 等待PostgreSQL启动" >> /start.sh && \
    echo "echo \"启动PostgreSQL...\"" >> /start.sh && \
    echo "service postgresql start" >> /start.sh && \
    echo "sleep 5" >> /start.sh && \
    echo "" >> /start.sh && \
    echo "# 确保数据库和用户存在" >> /start.sh && \
    echo "sudo -u postgres psql -c \"SELECT 1;\" 2>/dev/null || {" >> /start.sh && \
    echo "    echo \"初始化PostgreSQL...\"" >> /start.sh && \
    echo "    sudo -u postgres psql -c \"CREATE USER postgres WITH PASSWORD 'postgres';\" 2>/dev/null || true" >> /start.sh && \
    echo "    sudo -u postgres psql -c \"ALTER USER postgres CREATEDB;\" 2>/dev/null || true" >> /start.sh && \
    echo "    sudo -u postgres createdb backend 2>/dev/null || true" >> /start.sh && \
    echo "}" >> /start.sh && \
    echo "" >> /start.sh && \
    echo "# 停止PostgreSQL（由supervisor管理）" >> /start.sh && \
    echo "service postgresql stop" >> /start.sh && \
    echo "sleep 2" >> /start.sh && \
    echo "" >> /start.sh && \
    echo "# 启动supervisor管理所有服务" >> /start.sh && \
    echo "echo \"启动所有服务...\"" >> /start.sh && \
    echo "exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf" >> /start.sh

RUN chmod +x /start.sh

# 暴露端口
EXPOSE 1000 1001

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:1001/health || exit 1

# 设置数据卷
VOLUME ["/var/lib/postgresql", "/var/lib/redis", "/app/uploads"]

# 启动脚本
CMD ["/start.sh"] 