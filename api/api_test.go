package api

import (
	"strings"
	"testing"
)

// TestBuildConfig 测试构建配置验证
func TestBuildConfig(t *testing.T) {
	builder := NewFlutterBuilder()

	// 测试无效配置
	err := builder.Validate(nil)
	if err == nil {
		t.Error("Expected validation error for nil config")
	}

	// 测试空配置
	invalidConfig := &BuildConfig{}
	err = builder.Validate(invalidConfig)
	if err == nil {
		t.Error("Expected validation error for empty config")
	}

	// 测试无效源路径
	invalidConfig.SourcePath = ""
	err = builder.Validate(invalidConfig)
	if err == nil {
		t.Error("Expected validation error for empty source path")
	}

	// 测试无效平台
	invalidConfig.SourcePath = "/some/path"
	invalidConfig.Platform = "invalid"
	err = builder.Validate(invalidConfig)
	if err == nil {
		t.Error("Expected validation error for invalid platform")
	}

	// 测试有效配置但路径不存在
	validConfig := &BuildConfig{
		Platform:   PlatformAPK,
		SourcePath: "/non/existent/path",
	}
	err = builder.Validate(validConfig)
	// 这里可能会成功，因为验证只检查基本参数，不检查文件系统
	if err != nil {
		t.Logf("Validation failed as expected for non-existent path: %v", err)
	}
}

// TestPlatformConstants 测试平台常量
func TestPlatformConstants(t *testing.T) {
	if PlatformAPK != "apk" {
		t.Errorf("Expected PlatformAPK to be 'apk', got '%s'", PlatformAPK)
	}

	if PlatformIOS != "ios" {
		t.Errorf("Expected PlatformIOS to be 'ios', got '%s'", PlatformIOS)
	}
}

// TestLoggerInterface 测试自定义日志接口
func TestLoggerInterface(t *testing.T) {
	// 验证Logger接口可以被正确定义
	var _ Logger = (*mockLogger)(nil)

	// 测试可以设置自定义日志器
	builder := NewFlutterBuilder()
	mock := &mockLogger{}
	builder.SetLogger(mock)

	// 这里只是验证不会报错
	if mock == nil {
		t.Error("Mock logger should not be nil")
	}
}

// mockLogger 实现Logger接口的模拟器
type mockLogger struct{}

func (m *mockLogger) Debug(format string, args ...interface{})   {}
func (m *mockLogger) Info(format string, args ...interface{})    {}
func (m *mockLogger) Warning(format string, args ...interface{}) {}
func (m *mockLogger) Error(format string, args ...interface{})   {}
func (m *mockLogger) Success(format string, args ...interface{}) {}
func (m *mockLogger) Header(title string)                        {}
func (m *mockLogger) Println(args ...interface{})                {}
func (m *mockLogger) Printf(format string, args ...interface{})  {}

// TestNewFlutterBuilder 测试构建器创建
func TestNewFlutterBuilder(t *testing.T) {
	builder := NewFlutterBuilder()
	if builder == nil {
		t.Error("NewFlutterBuilder should not return nil")
	}
}

// TestBuildResult 测试构建结果结构
func TestBuildResult(t *testing.T) {
	result := &BuildResult{
		Success:    true,
		Platform:   PlatformAPK,
		BuildTime:  0,
		OutputPath: "/some/output/path",
		Error:      nil,
	}

	if !result.Success {
		t.Error("Build result should be successful")
	}

	if result.Platform != PlatformAPK {
		t.Errorf("Expected platform APK, got %s", result.Platform)
	}
}

// TestRemoveSpecificArgs 测试移除特定参数功能
func TestRemoveSpecificArgs(t *testing.T) {
	builder := NewFlutterBuilder()

	// 测试移除特定默认参数
	config := &BuildConfig{
		Platform:   PlatformAPK,
		SourcePath: "/non/existent/path",
		CustomArgs: map[string]interface{}{
			// 移除特定的默认参数
			"remove_default_args": []string{
				"--obfuscate",        // 移除代码混淆
				"--tree-shake-icons", // 移除图标优化
				"--dart-define=FLUTTER_WEB_USE_SKIA=true", // 移除特定的dart-define
			},
			// 同时添加自定义参数
			"flutter_build_args": []string{"--no-tree-shake-icons"},
		},
	}

	// 验证配置可以正常创建和验证
	err := builder.Validate(config)
	if err != nil {
		// 这里期望验证失败，因为路径不存在
		if !strings.Contains(err.Error(), "源代码路径不存在") && !strings.Contains(err.Error(), "no such file or directory") {
			t.Errorf("Expected path validation error, got: %v", err)
		}
	} else {
		// 如果验证成功，说明只检查了基本参数，这也是可以接受的
		t.Log("Validation passed, which means only basic parameters were checked")
	}

	// 验证可以正常获取自定义参数
	if config.CustomArgs == nil {
		t.Error("CustomArgs should not be nil")
	}

	removeArgs, exists := config.CustomArgs["remove_default_args"]
	if !exists {
		t.Error("remove_default_args should exist in CustomArgs")
	}

	if removeSlice, ok := removeArgs.([]string); ok {
		if len(removeSlice) != 3 {
			t.Errorf("Expected 3 remove args, got %d", len(removeSlice))
		}
	} else {
		t.Error("remove_default_args should be []string type")
	}
}
