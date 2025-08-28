package main

import (
	"fmt"
	"os"

	"github.com/mimicode/flutterbuilder/cmd"
	"github.com/mimicode/flutterbuilder/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	verbose    bool
	sourcePath string
)

func main() {
	// 创建根命令
	rootCmd := &cobra.Command{
		Use:   "flutter-builder",
		Short: "FFXApp Release Build Script v2.0 - 跨平台Flutter构建工具",
		Long: `FFXApp Release Build Script v2.0

跨平台Flutter项目构建脚本，支持iOS和Android release版本构建，
包含代码混淆、优化和安全配置。

支持平台:
  - apk: Android APK构建
  - ios: iOS应用构建

使用示例:
  flutter-builder apk --source-path /path/to/flutter/project
  flutter-builder ios --source-path /path/to/flutter/project
  flutter-builder apk --source-path . --verbose
  
  # iOS动态证书构建示例:
  flutter-builder ios --source-path /path/to/flutter/project \\
    --p12-cert /path/to/cert.p12 \\
    --cert-password "your_password" \\
    --provisioning-profile /path/to/profile.mobileprovision \\
    --team-id "TEAM123456" \\
    --bundle-id "com.company.app"`,
		Version: "2.0.0",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 设置日志级别
			if verbose {
				logger.SetLevel(logger.DebugLevel)
			}

			// 验证必需的源代码路径参数
			if sourcePath == "" {
				return fmt.Errorf("必须提供源代码路径参数: --source-path")
			}

			// 验证路径是否存在
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				return fmt.Errorf("源代码路径不存在: %s", sourcePath)
			}

			return nil
		},
	}

	// 添加全局标志
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "显示详细日志")
	rootCmd.PersistentFlags().StringVarP(&sourcePath, "source-path", "s", "", "Flutter项目源代码路径 (必需)")

	// 将源代码路径设为必需参数
	rootCmd.MarkPersistentFlagRequired("source-path")

	// 添加子命令
	rootCmd.AddCommand(cmd.NewAPKCommand())
	rootCmd.AddCommand(cmd.NewIOSCommand())

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
