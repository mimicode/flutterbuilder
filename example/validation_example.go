package main

import (
	"fmt"
	"log"

	"github.com/mimicode/flutterbuilder/api"
)

func main() {
	fmt.Println("Flutter构建产物验证功能示例")
	fmt.Println("==============================")

	// 示例1：使用默认验证配置构建APK
	fmt.Println("\n1. 使用默认验证配置构建APK")
	result1, err := api.QuickBuildAPK("/path/to/flutter/project")
	if err != nil {
		fmt.Printf("构建失败: %v\n", err)
	} else {
		fmt.Printf("构建成功！\n")
		fmt.Printf("  输出文件: %s\n", result1.OutputPath)
		fmt.Printf("  文件大小: %.2f MB\n", float64(result1.ArtifactSize)/(1024*1024))
		fmt.Printf("  验证通过: %t\n", result1.Verified)
		fmt.Printf("  构建耗时: %v\n", result1.BuildTime)
	}

	// 示例2：使用自定义验证配置
	fmt.Println("\n2. 使用自定义验证配置构建APK")
	customValidation := &api.ArtifactValidationConfig{
		EnableValidation:     true,
		EnableIntegrityCheck: false, // 禁用完整性检查以提高速度
		CustomMinSize:        1024 * 1024, // 1MB最小要求
		CustomMaxSize:        0, // 使用默认最大值
	}

	result2, err := api.QuickBuildAPKWithValidation("/path/to/flutter/project", customValidation)
	if err != nil {
		fmt.Printf("构建失败: %v\n", err)
	} else {
		fmt.Printf("构建成功！\n")
		fmt.Printf("  输出文件: %s\n", result2.OutputPath)
		fmt.Printf("  文件大小: %.2f MB\n", float64(result2.ArtifactSize)/(1024*1024))
		fmt.Printf("  验证通过: %t\n", result2.Verified)
		
		// 显示验证详情
		if result2.ValidationResult != nil {
			fmt.Println("  验证详情:")
			for _, detail := range result2.ValidationResult.ValidationDetails {
				status := "✓"
				if detail.Status != "success" {
					if detail.Critical {
						status = "✗"
					} else {
						status = "⚠"
					}
				}
				fmt.Printf("    %s %s\n", status, detail.Check)
			}
		}
	}

	// 示例3：完全禁用验证
	fmt.Println("\n3. 完全禁用验证构建APK")
	disabledValidation := api.DisableValidationConfig()

	result3, err := api.QuickBuildAPKWithValidation("/path/to/flutter/project", disabledValidation)
	if err != nil {
		fmt.Printf("构建失败: %v\n", err)
	} else {
		fmt.Printf("构建成功！\n")
		fmt.Printf("  输出文件: %s\n", result3.OutputPath)
		fmt.Printf("  验证状态: 已禁用\n")
		fmt.Printf("  构建耗时: %v\n", result3.BuildTime)
	}

	// 示例4：构建iOS应用（需要macOS环境）
	fmt.Println("\n4. 构建iOS应用示例")
	iosConfig := &api.IOSConfig{
		TeamID:   "YOUR_TEAM_ID",
		BundleID: "com.example.app",
		// 如果有证书文件，可以设置：
		// P12Cert: "/path/to/cert.p12",
		// CertPassword: "password",
		// ProvisioningProfile: "/path/to/profile.mobileprovision",
	}

	validationConfig := api.GetDefaultValidationConfig()
	result4, err := api.QuickBuildIOSWithValidation("/path/to/flutter/project", iosConfig, validationConfig)
	if err != nil {
		fmt.Printf("iOS构建失败: %v (可能需要macOS环境)\n", err)
	} else {
		fmt.Printf("iOS构建成功！\n")
		fmt.Printf("  输出文件: %s\n", result4.OutputPath)
		fmt.Printf("  文件大小: %.2f MB\n", float64(result4.ArtifactSize)/(1024*1024))
		fmt.Printf("  验证通过: %t\n", result4.Verified)
	}

	// 示例5：使用完整的BuildConfig进行高级配置
	fmt.Println("\n5. 使用完整配置进行构建")
	builder := api.NewFlutterBuilder()

	buildConfig := &api.BuildConfig{
		Platform:   api.PlatformAPK,
		SourcePath: "/path/to/flutter/project",
		Verbose:    true,
		CustomArgs: map[string]interface{}{
			"flutter_build_args": []string{"--target=lib/main_prod.dart"},
			"dart_defines":       []string{"ENVIRONMENT=production", "API_URL=https://api.prod.com"},
		},
		ValidationConfig: &api.ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: true,
			CustomMinSize:        5 * 1024 * 1024, // 5MB
		},
	}

	result5, err := builder.Build(buildConfig)
	if err != nil {
		fmt.Printf("高级构建失败: %v\n", err)
	} else {
		fmt.Printf("高级构建成功！\n")
		fmt.Printf("  平台: %s\n", result5.Platform)
		fmt.Printf("  输出文件: %s\n", result5.OutputPath)
		fmt.Printf("  文件大小: %.2f MB\n", float64(result5.ArtifactSize)/(1024*1024))
		fmt.Printf("  验证通过: %t\n", result5.Verified)
		fmt.Printf("  构建耗时: %v\n", result5.BuildTime)
	}

	fmt.Println("\n注意：以上示例需要真实的Flutter项目路径才能正常工作。")
	fmt.Println("产物验证功能现在已集成到所有构建流程中！")
}

// 演示如何在现有代码中升级到支持验证
func upgradeExistingCode() {
	// 旧代码（仍然可以工作）
	result, err := api.QuickBuildAPK("/path/to/project")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("构建完成，输出: %s\n", result.OutputPath)

	// 新代码（增加验证功能）
	validationConfig := api.GetDefaultValidationConfig()
	resultWithValidation, err := api.QuickBuildAPKWithValidation("/path/to/project", validationConfig)
	if err != nil {
		log.Fatal(err)
	}
	
	// 现在可以访问验证相关信息
	fmt.Printf("构建完成，输出: %s\n", resultWithValidation.OutputPath)
	fmt.Printf("验证通过: %t\n", resultWithValidation.Verified)
	fmt.Printf("文件大小: %.2f MB\n", float64(resultWithValidation.ArtifactSize)/(1024*1024))

	// 检查验证详情
	if resultWithValidation.ValidationResult != nil {
		for _, detail := range resultWithValidation.ValidationResult.ValidationDetails {
			if detail.Status == "failed" && detail.Critical {
				fmt.Printf("关键验证失败: %s - %s\n", detail.Check, detail.Message)
			}
		}
	}
}