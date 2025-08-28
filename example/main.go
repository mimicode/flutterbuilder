package main

import (
	"fmt"
	"log"

	"github.com/mimicode/flutterbuilder/api"
)

// 示例：如何使用 flutterbuilder 作为 Go 库

func main() {
	fmt.Println("=== Flutter Builder API 使用示例 ===")

	// 示例 1: 快速构建 APK
	fmt.Println("\n1. 快速构建示例:")
	quickBuildExample()

	// 示例 2: 自定义参数构建
	fmt.Println("\n2. 自定义参数构建示例:")
	customArgsExample()

	// 示例 3: 自定义日志库
	fmt.Println("\n3. 自定义日志库示例:")
	customLoggerExample()

	// 示例 4: iOS 构建示例
	fmt.Println("\n4. iOS 构建示例:")
	iosExample()
}

// quickBuildExample 快速构建示例
func quickBuildExample() {
	fmt.Println("快速构建 APK...")

	// 注意：这里使用的是示例路径，实际使用时请替换为真实的 Flutter 项目路径
	result, err := api.QuickBuildAPK("/path/to/your/flutter/project")
	if err != nil {
		fmt.Printf("构建失败（这是预期的，因为路径不存在）: %v\n", err)
		return
	}

	fmt.Printf("构建成功！输出路径: %s, 耗时: %v\n", result.OutputPath, result.BuildTime)
}

// customArgsExample 自定义参数构建示例
func customArgsExample() {
	fmt.Println("使用自定义参数构建...")

	builder := api.NewFlutterBuilder()

	config := &api.BuildConfig{
		Platform:   api.PlatformAPK,
		SourcePath: "/path/to/your/flutter/project",
		CustomArgs: map[string]interface{}{
			// 自定义 Flutter 构建参数
			"flutter_build_args": []string{
				"--no-shrink",            // 禁用代码压缩
				"--flavor", "production", // 使用 production flavor
			},
			// 自定义 Dart 定义
			"dart_defines": []string{
				"ENV=production",
				"API_URL=https://api.prod.com",
				"DEBUG_MODE=false",
			},
			// 自定义目标平台（支持多架构）
			"target_platform": "android-arm,android-arm64",
			// 禁用默认参数（如果需要完全自定义）
			"disable_default_args": false,
		},
		Verbose: true, // 启用详细日志
	}

	result, err := builder.Build(config)
	if err != nil {
		fmt.Printf("构建失败（这是预期的，因为路径不存在）: %v\n", err)
		return
	}

	fmt.Printf("构建成功！平台: %s, 耗时: %v\n", result.Platform, result.BuildTime)
}

// customLoggerExample 自定义日志库示例
func customLoggerExample() {
	fmt.Println("使用自定义日志库...")

	// 创建自定义日志器
	logger := &CustomLogger{prefix: "[CUSTOM] "}

	builder := api.NewFlutterBuilder()
	builder.SetLogger(logger)

	config := &api.BuildConfig{
		Platform:   api.PlatformAPK,
		SourcePath: "/path/to/your/flutter/project",
		Logger:     logger, // 也可以直接在配置中设置
	}

	// 验证配置（这会触发日志输出）
	err := builder.Validate(config)
	if err != nil {
		fmt.Printf("验证失败（预期的）: %v\n", err)
	}
}

// iosExample iOS 构建示例
func iosExample() {
	fmt.Println("iOS 构建示例...")

	// iOS 构建配置
	iosConfig := &api.IOSConfig{
		P12Cert:             "/path/to/your/cert.p12",
		CertPassword:        "your_cert_password",
		ProvisioningProfile: "/path/to/your/profile.mobileprovision",
		TeamID:              "YOUR_TEAM_ID",
		BundleID:            "com.yourcompany.yourapp",
	}

	result, err := api.QuickBuildIOS("/path/to/your/flutter/project", iosConfig)
	if err != nil {
		fmt.Printf("iOS 构建失败（这是预期的，因为路径不存在）: %v\n", err)
		return
	}

	fmt.Printf("iOS 构建成功！输出路径: %s\n", result.OutputPath)
}

// CustomLogger 自定义日志器实现
type CustomLogger struct {
	prefix string
}

func (l *CustomLogger) Debug(format string, args ...interface{}) {
	log.Printf(l.prefix+"[DEBUG] "+format, args...)
}

func (l *CustomLogger) Info(format string, args ...interface{}) {
	log.Printf(l.prefix+"[INFO] "+format, args...)
}

func (l *CustomLogger) Warning(format string, args ...interface{}) {
	log.Printf(l.prefix+"[WARNING] "+format, args...)
}

func (l *CustomLogger) Error(format string, args ...interface{}) {
	log.Printf(l.prefix+"[ERROR] "+format, args...)
}

func (l *CustomLogger) Success(format string, args ...interface{}) {
	log.Printf(l.prefix+"[SUCCESS] "+format, args...)
}

func (l *CustomLogger) Header(title string) {
	log.Printf(l.prefix+"=== %s ===", title)
}

func (l *CustomLogger) Println(args ...interface{}) {
	log.Println(append([]interface{}{l.prefix}, args...)...)
}

func (l *CustomLogger) Printf(format string, args ...interface{}) {
	log.Printf(l.prefix+format, args...)
}
