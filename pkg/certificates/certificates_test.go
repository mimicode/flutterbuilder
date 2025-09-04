package certificates

import (
	"testing"

	"github.com/mimicode/flutterbuilder/pkg/types"
)

func TestIdentifierGenerator(t *testing.T) {
	tests := []struct {
		name     string
		teamID   string
		bundleID string
		expected string
	}{
		{
			name:     "基本测试",
			teamID:   "ABCD123456",
			bundleID: "com.example.app",
			expected: "abcd123456_com_example_app",
		},
		{
			name:     "特殊字符测试",
			teamID:   "ABC-123",
			bundleID: "com.test.app-beta",
			expected: "abc_123_com_test_app_beta",
		},
		{
			name:     "长字符串截断测试",
			teamID:   "VERYLONGTEAMIDTHATEXCEEDSTHERECOMMENDEDLENGTH",
			bundleID: "com.verylongbundleid.application.name.test",
			expected: "verylongteamidthatexceedstherecommendedlength_com_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewIdentifierGenerator(tt.teamID, tt.bundleID)
			result := generator.Generate()
			
			if result != tt.expected {
				t.Errorf("生成的标识符不匹配，期望: %s, 实际: %s", tt.expected, result)
			}
		})
	}
}

func TestCertificateManager(t *testing.T) {
	// 创建测试配置
	config := &types.IOSConfig{
		TeamID:   "TESTTEAM123",
		BundleID: "com.test.app",
	}

	// 创建证书管理器
	manager := NewCertificateManager(config, "/tmp/test")
	
	// 测试获取唯一标识符
	identifier := manager.GetUniqueIdentifier()
	expected := "testteam123_com_test_app"
	
	if identifier != expected {
		t.Errorf("标识符生成失败，期望: %s, 实际: %s", expected, identifier)
	}
}

func TestCleanupRegistry(t *testing.T) {
	registry := NewCleanupRegistry()
	identifier := "test_identifier"
	
	resources := []types.CleanupResource{
		{
			Type:        types.ResourceKeychain,
			Path:        "/test/keychain.keychain",
			Description: "测试钥匙串",
		},
		{
			Type:        types.ResourceProvisioningProfile,
			Path:        "/test/profile.mobileprovision",
			Description: "测试描述文件",
		},
	}
	
	// 注册资源
	err := registry.Register(identifier, resources)
	if err != nil {
		t.Errorf("注册资源失败: %v", err)
	}
	
	// 检查资源是否已注册
	registered := registry.GetRegisteredResources()
	if _, exists := registered[identifier]; !exists {
		t.Errorf("资源未正确注册")
	}
	
	if len(registered[identifier]) != 2 {
		t.Errorf("注册的资源数量不正确，期望: 2, 实际: %d", len(registered[identifier]))
	}
}