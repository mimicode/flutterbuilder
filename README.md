# 自动化flutter构建工具

## 功能特性

- 🚀 **跨平台支持**: 支持 Windows、macOS 和 Linux
- 📱 **多平台构建**: 支持 Android APK 和 iOS 应用构建
- 🔐 **动态证书管理**: iOS 构建支持动态证书配置
- 🛡️ **安全配置检查**: 自动检查 ProGuard、签名配置等
- 🎨 **彩色输出**: 支持彩色终端输出，提升用户体验
- 📊 **详细日志**: 提供详细的构建过程日志
- ⚡ **高性能**: Go 语言提供更好的性能和并发支持

## 系统要求

- Go 1.20 或更高版本
- Flutter SDK
- Android SDK (用于 APK 构建)
- Xcode (用于 iOS 构建，仅 macOS)

## 安装和构建

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 构建可执行文件

```bash
# 构建当前平台版本
go build -o flutter-builder

# 构建跨平台版本
go build -o flutter-builder-windows.exe -ldflags="-s -w" -tags="windows"
go build -o flutter-builder-darwin -ldflags="-s -w" -tags="darwin"
go build -o flutter-builder-linux -ldflags="-s -w" -tags="linux"
```

## 使用方法

### 基本用法

```bash
# 构建 Android APK
./flutter-builder apk

# 构建 iOS 应用（使用系统证书）
./flutter-builder ios

# 启用详细日志
./flutter-builder apk --verbose
```

### iOS 动态证书构建

```bash
./flutter-builder ios \
  --p12-cert /path/to/cert.p12 \
  --cert-password "your_password" \
  --provisioning-profile /path/to/profile.mobileprovision \
  --team-id "TEAM123456" \
  --bundle-id "com.company.app"
```

## 项目结构

```
scripts/ioscert/
├── main.go                    # 主程序入口
├── go.mod                     # Go 模块文件
├── go.sum                     # 依赖校验文件
├── cmd/                       # 命令行命令
│   ├── apk.go                # APK 构建命令
│   └── ios.go                # iOS 构建命令
├── pkg/                       # 核心包
│   ├── builder/              # 构建器
│   │   ├── types.go          # 类型定义
│   │   └── flutter_builder.go # Flutter 构建器实现
│   ├── executor/             # 命令执行器
│   │   └── executor.go       # 命令执行实现
│   ├── security/             # 安全配置检查
│   │   └── security.go       # 安全检查实现
│   ├── certificates/         # iOS 证书管理
│   │   └── certificates.go   # 证书管理实现
│   └── logger/               # 日志系统
│       └── logger.go         # 日志实现
└── README.md                 # 项目说明
```

## 核心组件

### 1. FlutterBuilder
主要的构建逻辑实现，负责协调整个构建流程。

### 2. CommandExecutor
命令执行器，负责运行系统命令，支持跨平台。

### 3. SecurityChecker
安全配置检查器，检查 ProGuard、签名配置等。

### 4. CertificateManager
iOS 证书管理器，处理动态证书配置。

### 5. Logger
日志系统，提供彩色输出和不同级别的日志记录。

## 构建流程

1. **环境验证**: 检查 Flutter 环境和平台参数
2. **项目清理**: 清理构建缓存和旧文件
3. **依赖获取**: 获取项目依赖
4. **代码生成**: 运行代码生成工具
5. **安全检查**: 检查安全配置
6. **构建执行**: 执行实际的构建过程
7. **后处理**: 创建构建信息和安全提醒

## 开发说明

### 添加新功能

1. 在相应的包中添加接口定义
2. 实现具体的功能逻辑
3. 在构建器中集成新功能
4. 添加相应的测试

### 错误处理

所有错误都应该使用 `fmt.Errorf` 包装，提供有意义的错误信息。

### 日志记录

使用 `logger` 包记录不同级别的日志，避免使用 `fmt.Print`。

## 许可证

本项目遵循与原 Python 版本相同的许可证。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。
