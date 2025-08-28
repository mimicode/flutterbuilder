#!/bin/bash

# FFXApp Build Release - Go版本构建脚本

set -e

echo "🚀 开始构建 FFXApp Build Release Go版本..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go环境，请先安装Go 1.20或更高版本"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.20"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ 错误: Go版本过低，需要1.21或更高版本，当前版本: $GO_VERSION"
    exit 1
fi

echo "✅ Go环境检查通过，版本: $GO_VERSION"

# 安装依赖
echo "📦 安装项目依赖..."
go mod tidy

# 创建构建目录
BUILD_DIR="build"
mkdir -p $BUILD_DIR

# 构建当前平台版本
echo "🔨 构建当前平台版本..."
go build -o $BUILD_DIR/flutter-builder

# 构建跨平台版本
echo "🌍 构建跨平台版本..."

# Windows版本
echo "  - 构建Windows版本..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/flutter-builder-windows.exe

# macOS版本
echo "  - 构建macOS版本..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/flutter-builder-darwin

# Linux版本
echo "  - 构建Linux版本..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/flutter-builder-linux

# ARM64版本
echo "  - 构建ARM64版本..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $BUILD_DIR/flutter-builder-linux-arm64

echo "✅ 构建完成！"
echo ""
echo "📁 构建产物位置: $BUILD_DIR/"
echo "📋 文件列表:"
ls -la $BUILD_DIR/

echo ""
echo "🎯 使用方法:"
echo "  # 构建Android APK"
echo "  ./flutter-builder apk"
echo ""
echo "  # 构建iOS应用"
echo "  ./flutter-builder ios"
echo ""
echo "  # 启用详细日志"
echo "  ./flutter-builder apk --verbose"
