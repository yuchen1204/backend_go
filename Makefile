.PHONY: help build run test docs clean fmt vet deps

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  build    - 构建应用程序"
	@echo "  run      - 运行应用程序"
	@echo "  test     - 运行测试"
	@echo "  docs     - 生成API文档"
	@echo "  clean    - 清理构建文件"
	@echo "  fmt      - 格式化代码"
	@echo "  vet      - 检查代码"
	@echo "  deps     - 安装依赖"

# 构建应用程序
build:
	@echo "构建应用程序..."
	/usr/bin/go/bin/go build -o backend cmd/main.go

# 运行应用程序
run:
	@echo "启动应用程序..."
	/usr/bin/go/bin/go run cmd/main.go

# 运行测试
test:
	@echo "运行测试..."
	/usr/bin/go/bin/go test ./...

# 生成API文档
docs:
	@echo "生成API文档..."
	@if ! command -v swag &> /dev/null; then \
		echo "安装swag工具..."; \
		/usr/bin/go/bin/go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@if [ -f ~/go/bin/swag ]; then \
		~/go/bin/swag init -g cmd/main.go -o ./docs; \
	elif [ -f /root/go/bin/swag ]; then \
		/root/go/bin/swag init -g cmd/main.go -o ./docs; \
	else \
		echo "查找swag工具..."; \
		find ~ -name "swag" -type f 2>/dev/null | head -1 | xargs -I {} {} init -g cmd/main.go -o ./docs; \
	fi
	@echo "✅ API文档生成完成! 访问: http://localhost:8080/swagger/index.html"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -f backend
	rm -rf docs/

# 格式化代码
fmt:
	@echo "格式化代码..."
	/usr/bin/go/bin/go fmt ./...

# 检查代码
vet:
	@echo "检查代码..."
	/usr/bin/go/bin/go vet ./...

# 安装依赖
deps:
	@echo "安装依赖..."
	/usr/bin/go/bin/go mod tidy
	/usr/bin/go/bin/go mod download

# 开发环境初始化
dev-setup: deps docs
	@echo "开发环境初始化完成!"

# 构建并运行
dev: docs run 