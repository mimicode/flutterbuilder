# flutter自动化构建工具 Makefile

.PHONY: help build clean test install deps lint format

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_NAME=flutter-builder
BUILD_DIR=build
MAIN_FILE=main.go

# 帮助信息
help: ## 显示帮助信息
	@echo "flutter自动化构建工具"
	@echo ""
	@echo "可用命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# 安装依赖
deps: ## 安装项目依赖
	go mod tidy
	go mod download

# 构建当前平台版本
build: deps ## 构建当前平台版本
	@echo "🔨 构建当前平台版本..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 构建所有平台版本
build-all: deps ## 构建所有平台版本
	@echo "🌍 构建所有平台版本..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "  - Windows (AMD64)..."
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe $(MAIN_FILE)
	
	@echo "  - macOS (AMD64)..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin $(MAIN_FILE)
	
	@echo "  - Linux (AMD64)..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_FILE)
	
	@echo "  - Linux (ARM64)..."
	@GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)
	
	@echo "✅ 所有平台构建完成！"

# 运行测试
test: ## 运行测试
	@echo "🧪 运行测试..."
	go test -v ./...

# 运行测试并生成覆盖率报告
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "🧪 运行测试并生成覆盖率报告..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 覆盖率报告已生成: coverage.html"

# 代码格式化
format: ## 格式化代码
	@echo "🎨 格式化代码..."
	go fmt ./...
	goimports -w .

# 代码检查
lint: ## 运行代码检查
	@echo "🔍 运行代码检查..."
	golangci-lint run

# 安装到系统
install: build ## 安装到系统
	@echo "📦 安装到系统..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ 安装完成！可以使用 '$(BINARY_NAME)' 命令"

# 清理构建文件
clean: ## 清理构建文件
	@echo "🧹 清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ 清理完成！"

# 开发模式（监听文件变化并自动构建）
dev: ## 开发模式（需要安装air）
	@echo "🚀 启动开发模式..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "❌ 未找到air工具，请先安装: go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

# 快速构建（仅当前平台）
quick: ## 快速构建（跳过依赖安装）
	@echo "⚡ 快速构建..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ 快速构建完成！"

# 显示版本信息
version: ## 显示版本信息
	@echo "FFXApp Build Release - Go版本 v2.0.0 (Go 1.20+)"
	@echo "Go版本: $(shell go version)"
	@echo "构建时间: $(shell date)"

# 检查环境
check-env: ## 检查构建环境
	@echo "🔍 检查构建环境..."
	@echo "Go版本: $(shell go version)"
	@echo "Go模块: $(shell go env GOMOD)"
	@echo "Go工作目录: $(shell go env GOPWD)"
	@echo "操作系统: $(shell go env GOOS)"
	@echo "架构: $(shell go env GOARCH)"
