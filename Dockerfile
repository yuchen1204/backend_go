FROM golang:1.21-bullseye

# 设置工作目录
WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    openssh-server \
    postgresql \
    postgresql-contrib \
    redis-server \
    supervisor \
    vim \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 配置SSH
RUN mkdir /var/run/sshd
RUN echo 'root:root' | chpasswd
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config

# 配置PostgreSQL
USER postgres
RUN /etc/init.d/postgresql start && \
    psql --command "CREATE USER postgres WITH SUPERUSER PASSWORD 'postgres';" && \
    createdb -O postgres backend

USER root

# 配置Redis
RUN sed -i 's/^daemonize no/daemonize yes/' /etc/redis/redis.conf
RUN sed -i 's/^# requirepass foobared/requirepass/' /etc/redis/redis.conf

# 复制Go模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制项目文件
COPY . .

# 构建Go应用
RUN go build -o backend cmd/main.go

# 创建上传目录
RUN mkdir -p uploads/{default,avatar,document}

# 复制supervisor配置
COPY docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# 复制启动脚本
COPY docker/start.sh /start.sh
RUN chmod +x /start.sh

# 暴露端口
EXPOSE 22 5432 6379 8080

# 启动supervisor管理所有服务
CMD ["/start.sh"] 