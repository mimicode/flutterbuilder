// +build integration

package artifact

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mimicode/flutterbuilder/pkg/builder"
	"github.com/mimicode/flutterbuilder/pkg/types"
)

// TestIntegration_AndroidBuildWithValidation 测试Android构建与验证的集成
func TestIntegration_AndroidBuildWithValidation(t *testing.T) {
	// 跳过此测试，除非设置了集成测试环境变量
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("跳过集成测试，设置 INTEGRATION_TEST=1 来启用")
	}

	// 这需要一个真实的Flutter项目
	projectPath := os.Getenv("FLUTTER_TEST_PROJECT")
	if projectPath == "" {
		t.Skip("跳过集成测试，设置 FLUTTER_TEST_PROJECT 环境变量")
	}

	t.Run("构建Android APK并验证", func(t *testing.T) {
		// 创建构建器
		flutterBuilder := builder.NewFlutterBuilder("apk", nil, projectPath)

		// 设置自定义验证配置
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: true,
			CustomMinSize:        1024 * 1024, // 1MB
		}

		if validationBuilder, ok := flutterBuilder.(interface {
			SetValidationConfig(*ArtifactValidationConfig)
		}); ok {
			validationBuilder.SetValidationConfig(validationConfig)
		}

		// 执行构建
		err := flutterBuilder.Run()
		if err != nil {
			t.Fatalf("构建失败: %v", err)
		}

		// 验证APK文件存在
		apkPath := filepath.Join(projectPath, "build", "app", "outputs", "flutter-apk", "app-release.apk")
		if _, err := os.Stat(apkPath); os.IsNotExist(err) {
			t.Errorf("APK文件不存在: %s", apkPath)
		}

		// 手动执行验证以确认结果
		validator := NewArtifactValidator()
		artifactConfig := &ArtifactConfig{
			Platform:         PlatformAPK,
			SourcePath:       projectPath,
			ValidateIntegrity: true,
			ValidationConfig: validationConfig,
		}

		result, err := validator.ValidateArtifact(artifactConfig)
		if err != nil {
			t.Errorf("验证执行失败: %v", err)
		}

		if !result.Success {
			t.Errorf("验证失败: %v", result.Error)
		}

		if result.FileSize < 1024*1024 {
			t.Errorf("APK文件大小 %d 小于预期的最小值 1MB", result.FileSize)
		}

		// 检查验证详情
		var hasExistenceCheck, hasSizeCheck bool
		for _, detail := range result.ValidationDetails {
			switch detail.Check {
			case "APK文件存在性检查":
				hasExistenceCheck = true
				if detail.Status != "success" {
					t.Errorf("APK存在性检查失败: %s", detail.Message)
				}
			case "APK文件大小检查":
				hasSizeCheck = true
				if detail.Status != "success" {
					t.Errorf("APK大小检查失败: %s", detail.Message)
				}
			}
		}

		if !hasExistenceCheck {
			t.Error("缺少APK存在性检查")
		}
		if !hasSizeCheck {
			t.Error("缺少APK大小检查")
		}
	})
}

// TestIntegration_IOSBuildWithValidation 测试iOS构建与验证的集成
func TestIntegration_IOSBuildWithValidation(t *testing.T) {
	// 跳过此测试，除非设置了集成测试环境变量
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("跳过集成测试，设置 INTEGRATION_TEST=1 来启用")
	}

	// 检查是否在macOS上运行
	if os.Getenv("GOOS") != "darwin" {
		t.Skip("iOS构建需要macOS环境")
	}

	// 这需要一个真实的Flutter项目
	projectPath := os.Getenv("FLUTTER_TEST_PROJECT")
	if projectPath == "" {
		t.Skip("跳过集成测试，设置 FLUTTER_TEST_PROJECT 环境变量")
	}

	t.Run("构建iOS App并验证", func(t *testing.T) {
		// 创建构建器（无证书配置，仅构建.app）
		flutterBuilder := builder.NewFlutterBuilder("ios", nil, projectPath)

		// 设置自定义验证配置
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: false, // iOS App目录暂不支持完整性检查
			CustomMinSize:        10 * 1024 * 1024, // 10MB
		}

		if validationBuilder, ok := flutterBuilder.(interface {
			SetValidationConfig(*ArtifactValidationConfig)
		}); ok {
			validationBuilder.SetValidationConfig(validationConfig)
		}

		// 执行构建
		err := flutterBuilder.Run()
		if err != nil {
			t.Fatalf("构建失败: %v", err)
		}

		// 验证.app目录存在
		appPath := filepath.Join(projectPath, "build", "ios", "iphoneos", "Runner.app")
		if _, err := os.Stat(appPath); os.IsNotExist(err) {
			t.Errorf("iOS App目录不存在: %s", appPath)
		}

		// 手动执行验证以确认结果
		validator := NewArtifactValidator()
		artifactConfig := &ArtifactConfig{
			Platform:         PlatformIOS,
			SourcePath:       projectPath,
			ValidateIntegrity: false,
			ValidationConfig: validationConfig,
		}

		result, err := validator.ValidateArtifact(artifactConfig)
		if err != nil {
			t.Errorf("验证执行失败: %v", err)
		}

		if !result.Success {
			t.Errorf("验证失败: %v", result.Error)
		}

		// 检查验证详情
		var hasExistenceCheck, hasPlistCheck, hasExecutableCheck bool
		for _, detail := range result.ValidationDetails {
			switch detail.Check {
			case "iOS App目录存在性检查":
				hasExistenceCheck = true
				if detail.Status != "success" {
					t.Errorf("iOS App存在性检查失败: %s", detail.Message)
				}
			case "必要文件检查 (Info.plist)":
				hasPlistCheck = true
				if detail.Status != "success" {
					t.Errorf("Info.plist检查失败: %s", detail.Message)
				}
			case "必要文件检查 (Runner)":
				hasExecutableCheck = true
				if detail.Status != "success" {
					t.Errorf("Runner可执行文件检查失败: %s", detail.Message)
				}
			}
		}

		if !hasExistenceCheck {
			t.Error("缺少iOS App存在性检查")
		}
		if !hasPlistCheck {
			t.Error("缺少Info.plist检查")
		}
		if !hasExecutableCheck {
			t.Error("缺少Runner可执行文件检查")
		}
	})

	t.Run("构建IPA并验证", func(t *testing.T) {
		// 检查是否有iOS证书配置
		teamID := os.Getenv("IOS_TEAM_ID")
		bundleID := os.Getenv("IOS_BUNDLE_ID")
		if teamID == "" || bundleID == "" {
			t.Skip("跳过IPA测试，需要设置 IOS_TEAM_ID 和 IOS_BUNDLE_ID 环境变量")
		}

		iosConfig := &types.IOSConfig{
			TeamID:   teamID,
			BundleID: bundleID,
		}

		// 创建构建器
		flutterBuilder := builder.NewFlutterBuilder("ios", (*builder.IOSConfig)(iosConfig), projectPath)

		// 设置自定义验证配置
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: true,
			CustomMinSize:        5 * 1024 * 1024, // 5MB
		}

		if validationBuilder, ok := flutterBuilder.(interface {
			SetValidationConfig(*ArtifactValidationConfig)
		}); ok {
			validationBuilder.SetValidationConfig(validationConfig)
		}

		// 执行构建
		err := flutterBuilder.Run()
		if err != nil {
			t.Fatalf("构建失败: %v", err)
		}

		// 查找IPA文件
		ipaDir := filepath.Join(projectPath, "build", "ios", "ipa")
		ipaFiles, err := filepath.Glob(filepath.Join(ipaDir, "*.ipa"))
		if err != nil || len(ipaFiles) == 0 {
			t.Errorf("找不到IPA文件在目录: %s", ipaDir)
		}

		// 手动执行验证以确认结果
		validator := NewArtifactValidator()
		artifactConfig := &ArtifactConfig{
			Platform:         PlatformIOS,
			SourcePath:       projectPath,
			IOSConfig:        iosConfig,
			ValidateIntegrity: true,
			ValidationConfig: validationConfig,
		}

		result, err := validator.ValidateArtifact(artifactConfig)
		if err != nil {
			t.Errorf("验证执行失败: %v", err)
		}

		if !result.Success {
			t.Errorf("验证失败: %v", result.Error)
		}

		if result.FileSize < 5*1024*1024 {
			t.Errorf("IPA文件大小 %d 小于预期的最小值 5MB", result.FileSize)
		}
	})
}

// TestIntegration_ValidationConfigurationScenarios 测试不同验证配置场景
func TestIntegration_ValidationConfigurationScenarios(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("跳过集成测试，设置 INTEGRATION_TEST=1 来启用")
	}

	projectPath := os.Getenv("FLUTTER_TEST_PROJECT")
	if projectPath == "" {
		t.Skip("跳过集成测试，设置 FLUTTER_TEST_PROJECT 环境变量")
	}

	t.Run("禁用验证配置", func(t *testing.T) {
		flutterBuilder := builder.NewFlutterBuilder("apk", nil, projectPath)

		// 禁用验证
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     false,
			EnableIntegrityCheck: false,
		}

		if validationBuilder, ok := flutterBuilder.(interface {
			SetValidationConfig(*ArtifactValidationConfig)
		}); ok {
			validationBuilder.SetValidationConfig(validationConfig)
		}

		// 执行构建
		err := flutterBuilder.Run()
		if err != nil {
			t.Fatalf("构建失败: %v", err)
		}

		// 即使禁用验证，构建也应该成功
		// 这测试了验证不会干扰正常的构建流程
	})

	t.Run("仅启用存在性检查", func(t *testing.T) {
		flutterBuilder := builder.NewFlutterBuilder("apk", nil, projectPath)

		// 只启用验证，禁用完整性检查
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: false,
			CustomMinSize:        1024, // 1KB，很小的要求
		}

		if validationBuilder, ok := flutterBuilder.(interface {
			SetValidationConfig(*ArtifactValidationConfig)
		}); ok {
			validationBuilder.SetValidationConfig(validationConfig)
		}

		// 执行构建
		err := flutterBuilder.Run()
		if err != nil {
			t.Fatalf("构建失败: %v", err)
		}

		// 验证APK文件存在
		apkPath := filepath.Join(projectPath, "build", "app", "outputs", "flutter-apk", "app-release.apk")
		if _, err := os.Stat(apkPath); os.IsNotExist(err) {
			t.Errorf("APK文件不存在: %s", apkPath)
		}
	})

	t.Run("自定义大小限制", func(t *testing.T) {
		flutterBuilder := builder.NewFlutterBuilder("apk", nil, projectPath)

		// 设置很小的最小大小要求
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: true,
			CustomMinSize:        100,           // 100字节
			CustomMaxSize:        1024 * 1024 * 1024, // 1GB
		}

		if validationBuilder, ok := flutterBuilder.(interface {
			SetValidationConfig(*ArtifactValidationConfig)
		}); ok {
			validationBuilder.SetValidationConfig(validationConfig)
		}

		// 执行构建
		err := flutterBuilder.Run()
		if err != nil {
			t.Fatalf("构建失败: %v", err)
		}

		// 手动验证大小限制被正确应用
		validator := NewArtifactValidator()
		artifactConfig := &ArtifactConfig{
			Platform:         PlatformAPK,
			SourcePath:       projectPath,
			ValidateIntegrity: true,
			ValidationConfig: validationConfig,
		}

		result, err := validator.ValidateArtifact(artifactConfig)
		if err != nil {
			t.Errorf("验证执行失败: %v", err)
		}

		if !result.Success {
			t.Errorf("验证失败: %v", result.Error)
		}

		// 验证使用了自定义大小限制
		if result.FileSize < 100 {
			t.Errorf("文件大小 %d 小于自定义最小值 100字节", result.FileSize)
		}
	})
}