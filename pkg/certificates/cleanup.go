package certificates

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/mimicode/flutterbuilder/pkg/logger"
	"github.com/mimicode/flutterbuilder/pkg/types"
)

// CleanupRegistry 清理注册表接口
type CleanupRegistry interface {
	Register(identifier string, resources []types.CleanupResource) error
	Cleanup(identifier string) error
	CleanupAll() error
	GetRegisteredResources() map[string][]types.CleanupResource
}

// CleanupRegistryImpl 清理注册表实现
type CleanupRegistryImpl struct {
	mu        sync.RWMutex
	resources map[string][]types.CleanupResource
}

// NewCleanupRegistry 创建清理注册表
func NewCleanupRegistry() CleanupRegistry {
	return &CleanupRegistryImpl{
		resources: make(map[string][]types.CleanupResource),
	}
}

// Register 注册清理资源
func (cr *CleanupRegistryImpl) Register(identifier string, resources []types.CleanupResource) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	
	cr.resources[identifier] = resources
	logger.Info("已注册清理资源 [标识符: %s, 资源数量: %d]", identifier, len(resources))
	return nil
}

// Cleanup 清理指定标识符的资源
func (cr *CleanupRegistryImpl) Cleanup(identifier string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	
	resources, exists := cr.resources[identifier]
	if !exists {
		return nil
	}
	
	var errors []error
	for _, resource := range resources {
		if err := cr.cleanupResource(resource); err != nil {
			errors = append(errors, err)
		}
	}
	
	// 从注册表中移除
	delete(cr.resources, identifier)
	
	if len(errors) > 0 {
		return fmt.Errorf("清理过程中发生错误: %v", errors)
	}
	
	logger.Success("资源清理完成 [标识符: %s]", identifier)
	return nil
}

// CleanupAll 清理所有资源
func (cr *CleanupRegistryImpl) CleanupAll() error {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	
	var allErrors []error
	for identifier := range cr.resources {
		if err := cr.Cleanup(identifier); err != nil {
			allErrors = append(allErrors, err)
		}
	}
	
	if len(allErrors) > 0 {
		return fmt.Errorf("批量清理过程中发生错误: %v", allErrors)
	}
	
	return nil
}

// GetRegisteredResources 获取已注册的资源
func (cr *CleanupRegistryImpl) GetRegisteredResources() map[string][]types.CleanupResource {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	
	result := make(map[string][]types.CleanupResource)
	for k, v := range cr.resources {
		result[k] = v
	}
	return result
}

// cleanupResource 清理单个资源
func (cr *CleanupRegistryImpl) cleanupResource(resource types.CleanupResource) error {
	switch resource.Type {
	case types.ResourceKeychain:
		return cr.cleanupKeychain(resource.Path)
	case types.ResourceProvisioningProfile:
		return cr.cleanupProvisioningProfile(resource.Path)
	case types.ResourcePlistFile, types.ResourceTempDirectory:
		return cr.cleanupFile(resource.Path)
	default:
		return fmt.Errorf("未知的资源类型: %d", resource.Type)
	}
}

// cleanupKeychain 清理钥匙串
func (cr *CleanupRegistryImpl) cleanupKeychain(keychainPath string) error {
	// 检查钥匙串文件是否存在
	if _, err := os.Stat(keychainPath); os.IsNotExist(err) {
		return nil // 文件不存在，无需清理
	}
	
	// 使用 security 命令删除钥匙串
	cmd := exec.Command("security", "delete-keychain", keychainPath)
	if err := cmd.Run(); err != nil {
		logger.Warning("删除钥匙串失败: %s, 错误: %v", keychainPath, err)
		// 尝试直接删除文件
		if removeErr := os.Remove(keychainPath); removeErr != nil {
			return fmt.Errorf("删除钥匙串文件失败: %w", removeErr)
		}
	}
	
	logger.Info("已删除钥匙串: %s", keychainPath)
	return nil
}

// cleanupProvisioningProfile 清理描述文件
func (cr *CleanupRegistryImpl) cleanupProvisioningProfile(profilePath string) error {
	// 检查描述文件是否存在
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return nil // 文件不存在，无需清理
	}
	
	// 直接删除描述文件
	if err := os.Remove(profilePath); err != nil {
		return fmt.Errorf("删除描述文件失败: %w", err)
	}
	
	logger.Info("已删除描述文件: %s", profilePath)
	return nil
}

// cleanupFile 清理普通文件
func (cr *CleanupRegistryImpl) cleanupFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // 文件不存在，无需清理
	}
	
	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	
	logger.Info("已删除文件: %s", filePath)
	return nil
}