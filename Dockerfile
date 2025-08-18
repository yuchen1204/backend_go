# ---- 第一阶段：构建 ----
# 使用一个支持你项目所需 Go 版本的官方镜像
FROM golang:1.24-alpine AS builder

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

# ---- 第二阶段：运行 ----
# 使用一个轻量的基础镜像
FROM debian:latest

# 设置工作目录
WORKDIR /root/

# 安装系统 CA 证书，确保 TLS 校验证书链正常
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# 从构建器阶段复制编译后的二进制文件
COPY --from=builder /app/main .

# 从构建器阶段复制生成的 docs 目录
COPY --from=builder /app/docs ./docs

# 复制配置文件
COPY configs/ /root/configs/

# 暴露应用运行的端口
EXPOSE 8080

# 容器启动时运行的命令
CMD ["./main"]
