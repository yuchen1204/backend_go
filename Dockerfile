# ---- Base Go Builder ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git

# 构建Go应用
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/main.go -o ./docs
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# ---- Final Image ----
FROM alpine:latest

# 安装运行依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户和组
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# 复制Go后端和相关文件
COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/panel ./panel

# 创建上传目录并设置权限
RUN mkdir -p uploads/docs uploads/avatars && \
    chown -R appuser:appgroup /app

USER appuser

# 暴露后端端口
EXPOSE 8080

# 健康检查 (针对后端服务)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 直接启动Go应用
CMD ["./main"]
