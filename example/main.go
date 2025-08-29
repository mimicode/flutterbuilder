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

	// 示例 5: 移除特定参数示例（新功能）
	fmt.Println("\n5. 移除特定参数示例:")
	removeSpecificArgsExample()
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
			// 移除特定的默认参数（新功能）
			"remove_default_args": []string{
				"--obfuscate",        // 移除代码混淆（用于调试）
				"--tree-shake-icons", // 移除图标优化（保留所有图标）
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

	// 示例1: 使用证书配置构建IPA
	fmt.Println("\n  示例1: 构建IPA文件（提供证书配置）")
	iosConfigWithCert := &api.IOSConfig{
		P12Cert:             "/path/to/your/cert.p12",
		CertPassword:        "your_cert_password",
		ProvisioningProfile: "/path/to/your/profile.mobileprovision",
		TeamID:              "YOUR_TEAM_ID", // 关键：提供TeamID表示有证书配置
		BundleID:            "com.yourcompany.yourapp",
	}

	result1, err := api.QuickBuildIOS("/path/to/your/flutter/project", iosConfigWithCert)
	if err != nil {
		fmt.Printf("    IPA构建失败（这是预期的，因为路径不存在）: %v\n", err)
	} else {
		fmt.Printf("    IPA构建成功！输出路径: %s\n", result1.OutputPath)
	}

	// 示例2: 仅构建iOS项目（不提供证书配置）
	fmt.Println("\n  示例2: 仅构建iOS项目（不提供证书配置）")
	result2, err := api.QuickBuildIOS("/path/to/your/flutter/project", nil)
	if err != nil {
		fmt.Printf("    iOS构建失败（这是预期的，因为路径不存在）: %v\n", err)
	} else {
		fmt.Printf("    iOS构建成功！输出路径: %s\n", result2.OutputPath)
	}

	fmt.Println("\n  构建逻辑说明:")
	fmt.Println("  - 提供TeamID: 自动构建IPA文件，返回具体文件路径")
	fmt.Println("  - 未提供TeamID: 仅构建iOS项目，生成Runner.app")
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

// removeSpecificArgsExample 移除特定参数示例（新功能）
func removeSpecificArgsExample() {
	fmt.Println("移除特定默认参数...")

	builder := api.NewFlutterBuilder()

	config := &api.BuildConfig{
		Platform:   api.PlatformAPK,
		SourcePath: "/path/to/your/flutter/project",
		CustomArgs: map[string]interface{}{
			// 移除特定的默认参数
			"remove_default_args": []string{
				"--obfuscate",        // 移除代码混淆（便于调试）
				"--tree-shake-icons", // 移除图标优化（保留所有图标）
				"--dart-define=FLUTTER_WEB_USE_SKIA=true", // 移除Web配置
				"--target-platform",                       // 移除默认目标平台
			},
			// 添加自定义参数
			"flutter_build_args": []string{
				"--target-platform", "android-arm,android-arm64,android-x64", // 支持更多架构
				"--no-tree-shake-icons", // 不优化图标
			},
			// 添加自定义Dart定义
			"dart_defines": []string{
				"DEBUG_MODE=true",        // 开启调试模式
				"LOG_LEVEL=verbose",      // 设置日志级别
				"ENABLE_ANALYTICS=false", // 禁用分析
			},
		},
		Verbose: true, // 启用详细日志
	}

	fmt.Println("配置说明:")
	fmt.Println("- 移除了默认的代码混淆参数")
	fmt.Println("- 移除了默认的图标优化参数")
	fmt.Println("- 移除了默认的Web配置参数")
	fmt.Println("- 移除了默认的目标平台设置")
	fmt.Println("- 添加了支持更多架构的自定义目标平台")
	fmt.Println("- 添加了调试相关的Dart定义")

	result, err := builder.Build(config)
	if err != nil {
		fmt.Printf("构建失败（这是预期的，因为路径不存在）: %v\n", err)
		return
	}

	fmt.Printf("构建成功！平台: %s, 耗时: %v\n", result.Platform, result.BuildTime)
}
