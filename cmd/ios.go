package cmd

import (
	"fmt"

	"github.com/mimicode/flutterbuilder/pkg/builder"
	"github.com/mimicode/flutterbuilder/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	// iOS证书相关参数
	p12Cert             string
	certPassword        string
	provisioningProfile string
	teamID              string
	bundleID            string
)

var iosCmd = &cobra.Command{
	Use:   "ios",
	Short: "构建iOS应用发布版本",
	Long: `构建iOS应用发布版本

包含以下功能:
- 代码混淆和优化
- 调试信息分离
- Tree Shaking优化
- 动态证书管理
- 安全配置检查

支持动态证书配置:
- P12证书文件
- 证书密码
- 描述文件
- 团队ID
- Bundle ID`,
	RunE: runIOSBuild,
}

func NewIOSCommand() *cobra.Command {
	// 添加iOS证书相关标志
	iosCmd.Flags().StringVar(&p12Cert, "p12-cert", "", "P12证书文件路径")
	iosCmd.Flags().StringVar(&certPassword, "cert-password", "", "证书密码")
	iosCmd.Flags().StringVar(&provisioningProfile, "provisioning-profile", "", "描述文件(.mobileprovision)路径")
	iosCmd.Flags().StringVar(&teamID, "team-id", "", "开发者团队ID")
	iosCmd.Flags().StringVar(&bundleID, "bundle-id", "", "应用Bundle ID (如果与项目中的不同)")

	return iosCmd
}

func runIOSBuild(cmd *cobra.Command, args []string) error {
	logger.Header("FFXApp iOS Build")

	// 获取源代码路径
	sourcePath, _ := cmd.Flags().GetString("source-path")

	// 创建iOS配置
	iosConfig := &builder.IOSConfig{
		P12Cert:             p12Cert,
		CertPassword:        certPassword,
		ProvisioningProfile: provisioningProfile,
		TeamID:              teamID,
		BundleID:            bundleID,
	}

	// 创建构建器
	builder := builder.NewFlutterBuilder("ios", iosConfig, sourcePath)

	// 执行构建流程
	if err := builder.Run(); err != nil {
		return fmt.Errorf("iOS构建失败: %w", err)
	}

	return nil
}
