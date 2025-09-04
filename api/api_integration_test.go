// +build integration

package api

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIntegration_APIBuildWithValidation 测试API级别的构建和验证集成
func TestIntegration_APIBuildWithValidation(t *testing.T) {
	// 跳过此测试，除非设置了集成测试环境变量
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("跳过集成测试，设置 INTEGRATION_TEST=1 来启用")
	}

	// 这需要一个真实的Flutter项目
	projectPath := os.Getenv("FLUTTER_TEST_PROJECT")
	if projectPath == "" {
		t.Skip("跳过集成测试，设置 FLUTTER_TEST_PROJECT 环境变量")
	}

	t.Run("API构建Android APK带默认验证", func(t *testing.T) {
		builder := NewFlutterBuilder()

		config := &BuildConfig{
			Platform:   PlatformAPK,
			SourcePath: projectPath,
			Verbose:    true,
			// 使用默认验证配置
			ValidationConfig: GetDefaultValidationConfig(),
		}

		result, err := builder.Build(config)
		if err != nil {
			t.Fatalf("构建失败: %v", err)
		}

		// 检查构建结果
		if !result.Success {
			t.Errorf("构建应该成功")
		}

		if result.Platform != PlatformAPK {
			t.Errorf("平台应该是APK，实际: %s", result.Platform)
		}

		if result.OutputPath == "" {
			t.Errorf("输出路径不应该为空")
		}

		// 验证相关检查
		if !result.Verified {
			t.Errorf("验证应该通过")
		}

		if result.ArtifactSize == 0 {
			t.Errorf("产物大小应该大于0")
		}

		if result.ValidationResult == nil {
			t.Errorf("验证结果不应该为nil")
		}

		if !result.ValidationResult.Success {
			t.Errorf("验证结果应该成功")
		}

		// 验证文件实际存在
		if _, err := os.Stat(result.OutputPath); os.IsNotExist(err) {
			t.Errorf("输出文件不存在: %s", result.OutputPath)
		}

		t.Logf("构建成功，输出文件: %s, 大小: %.2f MB, 耗时: %v", 
			result.OutputPath, 
			float64(result.ArtifactSize)/(1024*1024), 
			result.BuildTime)
	})

	t.Run("API构建Android APK带自定义验证", func(t *testing.T) {
		builder := NewFlutterBuilder()

		// 自定义验证配置
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: false, // 禁用完整性检查以加快测试
			CustomMinSize:        1024 * 1024, // 1MB
			CustomMaxSize:        0, // 使用默认最大值
		}

		config := &BuildConfig{
			Platform:         PlatformAPK,
			SourcePath:       projectPath,
			Verbose:          true,
			ValidationConfig: validationConfig,
		}

		result, err := builder.Build(config)
		if err != nil {
			t.Fatalf("构建失败: %v", err)
		}

		// 检查自定义验证配置是否生效
		if !result.Success {
			t.Errorf("构建应该成功")
		}

		if !result.Verified {
			t.Errorf("验证应该通过")
		}

		if result.ArtifactSize < 1024*1024 {
			t.Errorf("产物大小 %d 应该大于等于自定义最小值 1MB", result.ArtifactSize)
		}

		// 检查验证详情
		if result.ValidationResult != nil {
			var hasIntegrityCheck bool
			for _, detail := range result.ValidationResult.ValidationDetails {
				if detail.Check == "APK完整性检查" {
					hasIntegrityCheck = true
					break
				}
			}
			if hasIntegrityCheck {
				t.Errorf("应该禁用完整性检查")
			}
		}
	})

	t.Run("API构建禁用验证", func(t *testing.T) {
		builder := NewFlutterBuilder()

		// 禁用验证
		validationConfig := DisableValidationConfig()

		config := &BuildConfig{
			Platform:         PlatformAPK,
			SourcePath:       projectPath,
			Verbose:          false,
			ValidationConfig: validationConfig,
		}

		result, err := builder.Build(config)
		if err != nil {
			t.Fatalf("构建失败: %v", err)
		}

		// 检查构建结果
		if !result.Success {
			t.Errorf("构建应该成功")
		}

		// 当验证被禁用时，Verified字段应该是false，但构建仍应成功
		if result.Verified {
			t.Errorf("禁用验证时Verified应该为false")
		}

		// ArtifactSize可能为0，因为没有执行验证
		if result.ValidationResult != nil && result.ValidationResult.Success {
			// 如果有验证结果，检查是否正确反映了禁用状态
			hasSkipDetail := false
			for _, detail := range result.ValidationResult.ValidationDetails {
				if detail.Check == "验证跳过" {
					hasSkipDetail = true
					break
				}
			}
			if !hasSkipDetail {
				t.Errorf("应该有验证跳过的详情")
			}
		}
	})
}

// TestIntegration_QuickBuildMethods 测试快捷构建方法
func TestIntegration_QuickBuildMethods(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("跳过集成测试，设置 INTEGRATION_TEST=1 来启用")
	}

	projectPath := os.Getenv("FLUTTER_TEST_PROJECT")
	if projectPath == "" {
		t.Skip("跳过集成测试，设置 FLUTTER_TEST_PROJECT 环境变量")
	}

	t.Run("QuickBuildAPK默认配置", func(t *testing.T) {
		result, err := QuickBuildAPK(projectPath)
		if err != nil {
			t.Fatalf("快速构建APK失败: %v", err)
		}

		if !result.Success {
			t.Errorf("构建应该成功")
		}

		// 快速构建使用默认验证配置
		if !result.Verified {
			t.Errorf("默认应该启用验证")
		}

		// 验证输出文件
		expectedPath := filepath.Join(projectPath, "build", "app", "outputs", "flutter-apk", "app-release.apk")
		if result.OutputPath != expectedPath {
			t.Errorf("输出路径不匹配，预期: %s, 实际: %s", expectedPath, result.OutputPath)
		}
	})

	t.Run("QuickBuildAPKWithValidation自定义配置", func(t *testing.T) {
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: true,
			CustomMinSize:        512 * 1024, // 512KB
		}

		result, err := QuickBuildAPKWithValidation(projectPath, validationConfig)
		if err != nil {
			t.Fatalf("带验证配置的快速构建APK失败: %v", err)
		}

		if !result.Success {
			t.Errorf("构建应该成功")
		}

		if !result.Verified {
			t.Errorf("验证应该通过")
		}

		if result.ArtifactSize < 512*1024 {
			t.Errorf("产物大小 %d 应该大于等于512KB", result.ArtifactSize)
		}
	})

	t.Run("QuickBuildWithValidation通用方法", func(t *testing.T) {
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: false,
		}

		result, err := QuickBuildWithValidation(PlatformAPK, projectPath, nil, validationConfig)
		if err != nil {
			t.Fatalf("通用快速构建失败: %v", err)
		}

		if !result.Success {
			t.Errorf("构建应该成功")
		}

		if !result.Verified {
			t.Errorf("验证应该通过")
		}
	})
}

// TestIntegration_IOSAPIBuild 测试iOS API构建
func TestIntegration_IOSAPIBuild(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("跳过集成测试，设置 INTEGRATION_TEST=1 来启用")
	}

	// 检查是否在macOS上运行
	if os.Getenv("GOOS") != "darwin" {
		t.Skip("iOS构建需要macOS环境")
	}

	projectPath := os.Getenv("FLUTTER_TEST_PROJECT")
	if projectPath == "" {
		t.Skip("跳过集成测试，设置 FLUTTER_TEST_PROJECT 环境变量")
	}

	t.Run("API构建iOS App", func(t *testing.T) {
		builder := NewFlutterBuilder()

		config := &BuildConfig{
			Platform:   PlatformIOS,
			SourcePath: projectPath,
			Verbose:    true,
			// iOS App构建使用默认验证配置
			ValidationConfig: GetDefaultValidationConfig(),
		}

		result, err := builder.Build(config)
		if err != nil {
			t.Fatalf("iOS构建失败: %v", err)
		}

		// 检查构建结果
		if !result.Success {
			t.Errorf("构建应该成功")
		}

		if result.Platform != PlatformIOS {
			t.Errorf("平台应该是iOS，实际: %s", result.Platform)
		}

		// 验证输出路径
		expectedPath := filepath.Join(projectPath, "build", "ios", "iphoneos", "Runner.app")
		if result.OutputPath != expectedPath {
			t.Errorf("输出路径不匹配，预期: %s, 实际: %s", expectedPath, result.OutputPath)
		}

		// 验证相关检查
		if !result.Verified {
			t.Errorf("验证应该通过")
		}

		// 验证文件实际存在
		if _, err := os.Stat(result.OutputPath); os.IsNotExist(err) {
			t.Errorf("输出文件不存在: %s", result.OutputPath)
		}

		t.Logf("iOS构建成功，输出: %s, 大小: %.2f MB, 耗时: %v", 
			result.OutputPath, 
			float64(result.ArtifactSize)/(1024*1024), 
			result.BuildTime)
	})

	t.Run("QuickBuildIOS", func(t *testing.T) {
		result, err := QuickBuildIOS(projectPath, nil)
		if err != nil {
			t.Fatalf("快速构建iOS失败: %v", err)
		}

		if !result.Success {
			t.Errorf("构建应该成功")
		}

		// 快速构建使用默认验证配置
		if !result.Verified {
			t.Errorf("默认应该启用验证")
		}
	})
}

// TestIntegration_ErrorScenarios 测试错误场景
func TestIntegration_ErrorScenarios(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("跳过集成测试，设置 INTEGRATION_TEST=1 来启用")
	}

	t.Run("无效项目路径", func(t *testing.T) {
		builder := NewFlutterBuilder()

		config := &BuildConfig{
			Platform:   PlatformAPK,
			SourcePath: "/nonexistent/project",
			ValidationConfig: GetDefaultValidationConfig(),
		}

		result, err := builder.Build(config)
		
		// 应该失败
		if err == nil {
			t.Errorf("预期应该失败")
		}

		if result.Success {
			t.Errorf("构建结果应该标记为失败")
		}

		if result.Error == nil {
			t.Errorf("结果应该包含错误信息")
		}
	})

	t.Run("空配置", func(t *testing.T) {
		builder := NewFlutterBuilder()

		result, err := builder.Build(nil)
		
		// 应该失败
		if err == nil {
			t.Errorf("预期应该失败")
		}

		if result == nil || result.Success {
			t.Errorf("应该返回失败结果")
		}
	})
}