package cmd

import (
	"fmt"

	"github.com/mimicode/flutterbuilder/pkg/builder"
	"github.com/mimicode/flutterbuilder/pkg/logger"

	"github.com/spf13/cobra"
)

var apkCmd = &cobra.Command{
	Use:   "apk",
	Short: "构建Android APK发布版本",
	Long: `构建Android APK发布版本

包含以下功能:
- 代码混淆和优化
- 调试信息分离
- Tree Shaking优化
- 仅ARM64架构支持
- 安全配置检查`,
	RunE: runAPKBuild,
}

func NewAPKCommand() *cobra.Command {
	return apkCmd
}

func runAPKBuild(cmd *cobra.Command, args []string) error {
	logger.Header("FFXApp Android APK Build")

	// 获取源代码路径
	sourcePath, _ := cmd.Flags().GetString("source-path")

	// 创建构建器
	builder := builder.NewFlutterBuilder("apk", nil, sourcePath)

	// 执行构建流程
	if err := builder.Run(); err != nil {
		return fmt.Errorf("APK构建失败: %w", err)
	}

	return nil
}
