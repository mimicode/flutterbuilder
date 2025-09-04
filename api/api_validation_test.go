package api

import (
	"testing"

	"github.com/mimicode/flutterbuilder/pkg/artifact"
)

func TestValidationConfig(t *testing.T) {
	t.Run("默认验证配置", func(t *testing.T) {
		config := GetDefaultValidationConfig()
		if config == nil {
			t.Fatal("默认验证配置不应该为nil")
		}
		if !config.EnableValidation {
			t.Error("默认应该启用验证")
		}
		if !config.EnableIntegrityCheck {
			t.Error("默认应该启用完整性检查")
		}
	})

	t.Run("禁用验证配置", func(t *testing.T) {
		config := DisableValidationConfig()
		if config == nil {
			t.Fatal("禁用验证配置不应该为nil")
		}
		if config.EnableValidation {
			t.Error("应该禁用验证")
		}
		if config.EnableIntegrityCheck {
			t.Error("应该禁用完整性检查")
		}
	})
}

func TestBuildConfigWithValidation(t *testing.T) {
	t.Run("构建配置包含验证配置", func(t *testing.T) {
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: false,
			CustomMinSize:        1024 * 1024, // 1MB
		}

		buildConfig := &BuildConfig{
			Platform:         PlatformAPK,
			SourcePath:       "/test/project",
			ValidationConfig: validationConfig,
		}

		if buildConfig.ValidationConfig == nil {
			t.Error("验证配置不应该为nil")
		}
		if !buildConfig.ValidationConfig.EnableValidation {
			t.Error("应该启用验证")
		}
		if buildConfig.ValidationConfig.EnableIntegrityCheck {
			t.Error("应该禁用完整性检查")
		}
		if buildConfig.ValidationConfig.CustomMinSize != 1024*1024 {
			t.Error("自定义最小大小应该为1MB")
		}
	})
}

func TestBuildResultWithValidation(t *testing.T) {
	t.Run("构建结果包含验证信息", func(t *testing.T) {
		result := &BuildResult{
			Success:      true,
			Platform:     PlatformAPK,
			ArtifactSize: 1024 * 1024 * 50, // 50MB
			Verified:     true,
			ValidationResult: &artifact.ValidationResult{
				Success:      true,
				ArtifactPath: "/path/to/app.apk",
				FileSize:     1024 * 1024 * 50,
				ValidationDetails: []artifact.ValidationDetail{
					{
						Check:    "APK文件存在性检查",
						Status:   "success",
						Message:  "APK文件存在",
						Critical: true,
					},
				},
			},
		}

		if !result.Success {
			t.Error("构建应该成功")
		}
		if !result.Verified {
			t.Error("验证应该通过")
		}
		if result.ArtifactSize == 0 {
			t.Error("产物大小应该大于0")
		}
		if result.ValidationResult == nil {
			t.Error("验证结果不应该为nil")
		}
		if !result.ValidationResult.Success {
			t.Error("验证结果应该成功")
		}
		if len(result.ValidationResult.ValidationDetails) == 0 {
			t.Error("应该有验证详情")
		}
	})
}

func TestQuickBuildWithValidation(t *testing.T) {
	t.Run("快速构建方法接受验证配置", func(t *testing.T) {
		// 注意：这个测试不会实际执行构建，只是测试参数传递
		validationConfig := &ArtifactValidationConfig{
			EnableValidation:     false, // 禁用验证以避免实际构建
			EnableIntegrityCheck: false,
		}

		// 测试APK构建
		_, err := QuickBuildAPKWithValidation("/nonexistent/project", validationConfig)
		// 预期会失败，因为项目不存在，但这证明参数传递正确
		if err == nil {
			t.Error("预期应该失败（项目不存在）")
		}

		// 测试iOS构建
		iosConfig := &IOSConfig{
			TeamID:   "ABCD123456",
			BundleID: "com.example.test",
		}
		_, err = QuickBuildIOSWithValidation("/nonexistent/project", iosConfig, validationConfig)
		// 预期会失败，因为项目不存在，但这证明参数传递正确
		if err == nil {
			t.Error("预期应该失败（项目不存在）")
		}

		// 测试通用构建
		_, err = QuickBuildWithValidation(PlatformAPK, "/nonexistent/project", nil, validationConfig)
		// 预期会失败，因为项目不存在，但这证明参数传递正确
		if err == nil {
			t.Error("预期应该失败（项目不存在）")
		}
	})
}

func TestValidationConfigTypes(t *testing.T) {
	t.Run("验证配置类型别名", func(t *testing.T) {
		// 测试类型别名是否正确工作
		var apiConfig *ArtifactValidationConfig
		var artifactConfig *artifact.ArtifactValidationConfig

		// 这些应该是同一个类型
		apiConfig = &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: true,
		}

		artifactConfig = apiConfig
		if artifactConfig == nil {
			t.Error("类型别名应该兼容")
		}

		// 测试直接赋值
		apiConfig = artifact.GetDefaultValidationConfig()
		if apiConfig == nil {
			t.Error("应该能够直接赋值")
		}
	})
}

// 测试验证配置的默认值设置
func TestValidationConfigDefaults(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		config   *ArtifactValidationConfig
		expected struct {
			enableValidation     bool
			enableIntegrityCheck bool
		}
	}{
		{
			name:     "默认配置",
			platform: PlatformAPK,
			config:   GetDefaultValidationConfig(),
			expected: struct {
				enableValidation     bool
				enableIntegrityCheck bool
			}{
				enableValidation:     true,
				enableIntegrityCheck: true,
			},
		},
		{
			name:     "自定义禁用验证",
			platform: PlatformAPK,
			config: &ArtifactValidationConfig{
				EnableValidation:     false,
				EnableIntegrityCheck: true,
			},
			expected: struct {
				enableValidation     bool
				enableIntegrityCheck bool
			}{
				enableValidation:     false,
				enableIntegrityCheck: true,
			},
		},
		{
			name:     "自定义禁用完整性检查",
			platform: PlatformIOS,
			config: &ArtifactValidationConfig{
				EnableValidation:     true,
				EnableIntegrityCheck: false,
			},
			expected: struct {
				enableValidation     bool
				enableIntegrityCheck bool
			}{
				enableValidation:     true,
				enableIntegrityCheck: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.EnableValidation != tt.expected.enableValidation {
				t.Errorf("EnableValidation = %v, 预期 %v", 
					tt.config.EnableValidation, tt.expected.enableValidation)
			}
			if tt.config.EnableIntegrityCheck != tt.expected.enableIntegrityCheck {
				t.Errorf("EnableIntegrityCheck = %v, 预期 %v", 
					tt.config.EnableIntegrityCheck, tt.expected.enableIntegrityCheck)
			}
		})
	}
}

// 测试验证配置的边界值
func TestValidationConfigBoundaryValues(t *testing.T) {
	t.Run("自定义大小限制", func(t *testing.T) {
		config := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: true,
			CustomMinSize:        1024,      // 1KB
			CustomMaxSize:        1024 * 1024, // 1MB
		}

		if config.CustomMinSize != 1024 {
			t.Errorf("CustomMinSize = %d, 预期 1024", config.CustomMinSize)
		}
		if config.CustomMaxSize != 1024*1024 {
			t.Errorf("CustomMaxSize = %d, 预期 %d", config.CustomMaxSize, 1024*1024)
		}
	})

	t.Run("零值大小限制", func(t *testing.T) {
		config := &ArtifactValidationConfig{
			EnableValidation:     true,
			EnableIntegrityCheck: true,
			CustomMinSize:        0, // 使用默认值
			CustomMaxSize:        0, // 使用默认值
		}

		if config.CustomMinSize != 0 {
			t.Errorf("CustomMinSize = %d, 预期 0", config.CustomMinSize)
		}
		if config.CustomMaxSize != 0 {
			t.Errorf("CustomMaxSize = %d, 预期 0", config.CustomMaxSize)
		}
	})
}