# 跨平台Flutter自动化构建工具

## 功能特性

- 🚀 **跨平台支持**: 支持 Windows、macOS 和 Linux
- 📱 **多平台构建**: 支持 Android APK 和 iOS 应用构建
- 🔐 **动态证书管理**: iOS 构建支持动态证书配置
- 🛡️ **安全配置检查**: 自动检查 ProGuard、签名配置等
- 🎨 **彩色输出**: 支持彩色终端输出，提升用户体验
- 📊 **详细日志**: 提供详细的构建过程日志，支持外部日志库集成
- ⚡ **高性能**: Go 语言提供更好的性能和并发支持
- 📚 **库引用支持**: 可作为Go模块被其他项目引用
- 🔧 **自定义构建参数**: 支持传入自定义Flutter构建参数

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

### 命令行工具使用

#### 基本用法

```bash
# 构建 Android APK
./flutter-builder apk --source-path /path/to/flutter/project

# 构建 iOS 应用（使用系统证书）
./flutter-builder ios --source-path /path/to/flutter/project

# 启用详细日志
./flutter-builder apk --source-path /path/to/flutter/project --verbose
```

#### iOS 动态证书构建

```
# 使用动态证书构建 IPA 文件
./flutter-builder ios \
  --source-path /path/to/flutter/project \
  --p12-cert /path/to/cert.p12 \
  --cert-password "your_password" \
  --provisioning-profile /path/to/profile.mobileprovision \
  --team-id "TEAM123456" \
  --bundle-id "com.company.app"

# 仅构建 iOS 项目（不生成 IPA）
./flutter-builder ios \
  --source-path /path/to/flutter/project
```

**iOS 构建逻辑说明：**
- **提供证书配置**：自动构建 IPA 文件，输出具体的 IPA 文件路径（如 `build/ios/ipa/Runner.ipa`）
- **未提供证书配置**：仅构建 iOS 项目，输出 Runner.app 文件，路径为 `build/ios/iphoneos/Runner.app`

### 作为Go模块引用

项目现在支持作为 Go 模块被其他项目引用：

#### 安装

```bash
go get github.com/mimicode/flutterbuilder
```

#### 快速使用

```go
package main

import (
    "fmt"
    "log"
    "github.com/mimicode/flutterbuilder/api"
)

func main() {
    // 快速构建 APK
    result, err := api.QuickBuildAPK("/path/to/flutter/project")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("构建成功！输出路径: %s", result.OutputPath)
}
```

#### 高级使用（自定义参数）

```go
package main

import (
    "fmt"
    "log"
    "github.com/mimicode/flutterbuilder/api"
)

func main() {
    builder := api.NewFlutterBuilder()
    
    config := &api.BuildConfig{
        Platform:   api.PlatformAPK,
        SourcePath: "/path/to/flutter/project",
        CustomArgs: map[string]interface{}{
            "flutter_build_args": []string{"--no-shrink", "--flavor", "production"},
            "dart_defines": []string{"ENV=production", "API_URL=https://api.prod.com"},
            "target_platform": "android-arm,android-arm64",
        },
        Verbose: true,
    }
    
    result, err := builder.Build(config)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("构建成功！耗时: %v", result.BuildTime)
}
```

#### 自定义日志库

``go
package main

import (
    "fmt"
    "github.com/mimicode/flutterbuilder/api"
)

// 实现自定义日志接口
type MyLogger struct{}

func (l *MyLogger) Debug(format string, args ...interface{}) {
    fmt.Printf("[DEBUG] "+format+"\n", args...)
}

func (l *MyLogger) Info(format string, args ...interface{}) {
    fmt.Printf("[INFO] "+format+"\n", args...)
}

// 实现其他日志方法...

func main() {
    builder := api.NewFlutterBuilder()
    builder.SetLogger(&MyLogger{})
    
    config := &api.BuildConfig{
        Platform:   api.PlatformAPK,
        SourcePath: "/path/to/flutter/project",
    }
    
    result, err := builder.Build(config)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 自定义构建参数说明

项目支持以下自定义参数：

| 参数名 | 类型 | 说明 |
|---------|------|------|
| `disable_default_args` | bool | 禁用所有默认构建参数 |
| `remove_default_args` | []string | 移除指定的默认参数（新增） |
| `flutter_build_args` | []string | 自定义Flutter构建参数 |
| `dart_defines` | []string | 自定义Dart定义参数 |
| `target_platform` | string | 自定义目标平台（仅Android） |

#### 参数优先级说明

1. **全部禁用** (`disable_default_args: true`): 不使用任何默认参数
2. **选择性移除** (`remove_default_args`): 从默认参数中移除指定参数
3. **添加自定义** (`flutter_build_args`): 添加新的构建参数

#### 默认参数列表

**Android APK 默认参数:**
- `--obfuscate` - 代码混淆
- `--split-debug-info=build/debug-info` - 调试信息分离
- `--tree-shake-icons` - 图标优化
- `--target-platform android-arm64` - 目标平台
- `--dart-define=FLUTTER_WEB_USE_SKIA=true` - Web配置
- `--dart-define=FLUTTER_WEB_AUTO_DETECT=true` - Web自动检测

**iOS 默认参数:**
- `--obfuscate` - 代码混淆
- `--split-debug-info=build/debug-info` - 调试信息分离
- `--tree-shake-icons` - 图标优化
- `--dart-define=FLUTTER_WEB_USE_SKIA=true` - Web配置
- `--dart-define=FLUTTER_WEB_AUTO_DETECT=true` - Web自动检测

## 项目结构

```
flutterbuilder/
├── main.go                    # 主程序入口
├── go.mod                     # Go 模块文件
├── go.sum                     # 依赖校验文件
├── api/                       # 公开API接口
│   └── api.go                 # 库引用接口
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
│   ├── logger/               # 日志系统
│   │   ├── logger.go         # 日志实现（支持外部日志库）
│   │   └── logger_test.go    # 日志测试
│   └── types/                # 公共类型定义
│       └── ios_config.go     # iOS 配置类型
├── Makefile                   # 构建、测试、部署脚本
├── build.sh                   # 跨平台构建脚本
├── build.bat                  # Windows 构建脚本
└── README.md                 # 项目说明
```

## 核心组件

### 1. FlutterBuilder
主要的构建逻辑实现，负责协调整个构建流程，支持自定义构建参数。

### 2. CommandExecutor
命令执行器，负责运行系统命令，支持跨平台。

### 3. SecurityChecker
安全配置检查器，检查 ProGuard、签名配置等。

### 4. CertificateManager
iOS 证书管理器，处理动态证书配置。

### 5. Logger
日志系统，提供彩色输出和不同级别的日志记录，支持外部日志库集成。

### 6. API 接口
公开的 API 接口，使其他 Go 项目可以直接引用本库进行 Flutter 构建。

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

本项目采用 MIT 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 更新日志

### v2.0.0
- 重构为 Go 语言实现，提升性能和跨平台兼容性
- 增加库引用支持，可作为 Go 模块被其他项目引用
- 优化日志系统，支持外部日志库集成
- 支持自定义构建参数传入
- 完善 API 接口设计
