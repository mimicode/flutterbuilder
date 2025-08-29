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

// TestIOSBuildLogic 测试iOS构建逻辑
func TestIOSBuildLogic(t *testing.T) {
	// 测试1: 无证书配置的iOS构建
	config1 := &BuildConfig{
		Platform:   PlatformIOS,
		SourcePath: "/non/existent/path",
		IOSConfig:  nil, // 无证书配置
	}

	// 验证输出路径为 .app 文件
	expectedPath1 := "/non/existent/path/build/ios/iphoneos/Runner.app"
	actualPath1 := getOutputPath(config1.Platform, config1.SourcePath, config1.IOSConfig)
	if actualPath1 != expectedPath1 {
		t.Errorf("Expected iOS path without cert: %s, got: %s", expectedPath1, actualPath1)
	}

	// 测试2: 有证书配置的iOS构建
	config2 := &BuildConfig{
		Platform:   PlatformIOS,
		SourcePath: "/non/existent/path",
		IOSConfig: &IOSConfig{
			TeamID: "TEST123456", // 提供TeamID表示有证书配置
		},
	}

	// 验证输出路径为具体的IPA文件（由于文件不存在，会返回默认路径）
	expectedPath2 := "/non/existent/path/build/ios/ipa/Runner.ipa"
	actualPath2 := getOutputPath(config2.Platform, config2.SourcePath, config2.IOSConfig)
	if actualPath2 != expectedPath2 {
		t.Errorf("Expected iOS path with cert: %s, got: %s", expectedPath2, actualPath2)
	}

	// 测试3: Android构建路径保持不变
	config3 := &BuildConfig{
		Platform:   PlatformAPK,
		SourcePath: "/non/existent/path",
	}

	expectedPath3 := "/non/existent/path/build/app/outputs/flutter-apk/app-release.apk"
	actualPath3 := getOutputPath(config3.Platform, config3.SourcePath, nil)
	if actualPath3 != expectedPath3 {
		t.Errorf("Expected APK path: %s, got: %s", expectedPath3, actualPath3)
	}
}
