@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo 🚀 开始构建 FFXApp Build Release Go版本...

REM 检查Go环境
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ 错误: 未找到Go环境，请先安装Go 1.20或更高版本
    pause
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
set GO_VERSION=!GO_VERSION:go=!

echo ✅ Go环境检查通过，版本: !GO_VERSION!

REM 安装依赖
echo 📦 安装项目依赖...
go mod tidy

REM 创建构建目录
set BUILD_DIR=build
if not exist %BUILD_DIR% mkdir %BUILD_DIR%

REM 构建当前平台版本
echo 🔨 构建当前平台版本...
go build -o %BUILD_DIR%\flutter-builder.exe

REM 构建跨平台版本
echo 🌍 构建跨平台版本...

REM Windows版本
echo   - 构建Windows版本...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%\flutter-builder-windows.exe

REM macOS版本
echo   - 构建macOS版本...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%\flutter-builder-darwin

REM Linux版本
echo   - 构建Linux版本...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%\flutter-builder-linux

REM ARM64版本
echo   - 构建ARM64版本...
set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -o %BUILD_DIR%\flutter-builder-linux-arm64

echo ✅ 构建完成！
echo.
echo 📁 构建产物位置: %BUILD_DIR%\
echo 📋 文件列表:
dir %BUILD_DIR%

echo.
echo 🎯 使用方法:
echo   # 构建Android APK
echo   flutter-builder.exe apk
echo.
echo   # 构建iOS应用
echo   flutter-builder.exe ios
echo.
echo   # 启用详细日志
echo   flutter-builder.exe apk --verbose

pause
