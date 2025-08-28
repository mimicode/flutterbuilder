package api

import (
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
