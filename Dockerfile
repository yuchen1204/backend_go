# ---- 第一阶段：构建 Go 后端 ----
# 使用一个支持你项目所需 Go 版本的官方镜像
FROM golang:1.24-alpine AS go-builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 复制 vendored 依赖
COPY vendor ./vendor

# 复制所有源代码
COPY . .

# 从 vendored 依赖中安装 swag 工具
RUN go install -mod=vendor github.com/swaggo/swag/cmd/swag

# 运行 swag init 来生成 docs 目录
RUN /go/bin/swag init --parseInternal -g ./cmd/main.go

# 使用 vendored 依赖进行编译
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/main ./cmd/main.go

# ---- 第二阶段：构建前端 Admin Panel ----
FROM node:20-alpine AS frontend-builder

# 设置工作目录
WORKDIR /app/frontend

# 复制 package.json 和 package-lock.json
COPY backend-panel/package*.json ./

# 清理并重新安装依赖
RUN npm ci

# 复制前端源代码
COPY backend-panel/ ./

# 构建前端项目
RUN npm run build

# ---- 第三阶段：运行 ----
# 使用一个轻量的基础镜像
FROM debian:latest

# 设置工作目录
WORKDIR /root/

# 安装系统 CA 证书和 Python3（用于简易 HTTP 服务器），确保 TLS 校验证书链正常
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates python3 \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# 从 Go 构建器阶段复制编译后的二进制文件
COPY --from=go-builder /app/main .

# 从 Go 构建器阶段复制生成的 docs 目录
COPY --from=go-builder /app/docs ./docs

# 从前端构建器阶段复制构建后的前端文件
COPY --from=frontend-builder /app/frontend/dist ./admin-panel

# 复制配置文件
COPY configs/ /root/configs/

# 创建启动脚本
RUN echo '#!/bin/bash\n\
# 启动后端服务\n\
./main &\n\
\n\
# 启动前端静态文件服务器在1234端口\n\
cd /root/admin-panel\n\
python3 -m http.server 1234 &\n\
\n\
# 等待所有后台进程\n\
wait' > start.sh && chmod +x start.sh

# 暴露应用运行的端口
EXPOSE 8080 1234

# 容器启动时运行启动脚本
CMD ["./start.sh"]
