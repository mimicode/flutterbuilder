package certificates

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/mimicode/flutterbuilder/pkg/executor"
	"github.com/mimicode/flutterbuilder/pkg/logger"
	"github.com/mimicode/flutterbuilder/pkg/types"
)

// CertificateManagerImpl iOS证书管理器实现
type CertificateManagerImpl struct {
	iosConfig         *types.IOSConfig
	projectRoot       string
	executor          executor.CommandExecutor
	uniqueIdentifier  string
	cleanupRegistry   CleanupRegistry
	tempKeychainPath  string
	installedPPPath   string
	tempPlistPath     string
	cleanupRegistered bool
}

// NewCertificateManager 创建新的证书管理器
func NewCertificateManager(iosConfig *types.IOSConfig, projectRoot string) types.CertificateManager {
	if iosConfig == nil {
		return &CertificateManagerImpl{
			projectRoot:     projectRoot,
			executor:        executor.NewCommandExecutor(),
			cleanupRegistry: NewCleanupRegistry(),
		}
	}

	generator := NewIdentifierGenerator(iosConfig.TeamID, iosConfig.BundleID)

	return &CertificateManagerImpl{
		iosConfig:        iosConfig,
		projectRoot:      projectRoot,
		executor:         executor.NewCommandExecutor(),
		uniqueIdentifier: generator.Generate(),
		cleanupRegistry:  NewCleanupRegistry(),
	}
}

// SetupCertificates 设置iOS证书
func (c *CertificateManagerImpl) SetupCertificates() error {
	if c.iosConfig == nil {
		return nil
	}

	logger.Info("设置iOS证书和描述文件 [标识符: %s]", c.uniqueIdentifier)

	// 注册清理资源
	if err := c.registerCleanupResources(); err != nil {
		return fmt.Errorf("注册清理资源失败: %w", err)
	}

	// 设置信号处理器确保异常情况下的清理
	c.setupSignalHandler()

	// 设置P12证书
	if c.iosConfig.P12Cert != "" && c.iosConfig.CertPassword != "" {
		if err := c.setupTemporaryKeychain(); err != nil {
			c.ForceCleanupAll() // 确保清理
			return fmt.Errorf("设置临时钥匙串失败: %w", err)
		}
	}

	// 安装描述文件
	if c.iosConfig.ProvisioningProfile != "" {
		if err := c.installProvisioningProfile(); err != nil {
			c.ForceCleanupAll() // 确保清理
			return fmt.Errorf("安装描述文件失败: %w", err)
		}
	}

	return nil
}

// CleanupCertificates 清理iOS证书配置
func (c *CertificateManagerImpl) CleanupCertificates() error {
	startTime := time.Now()
	logger.Info("开始清理证书资源 [标识符: %s]", c.uniqueIdentifier)

	defer func() {
		duration := time.Since(startTime)
		logger.Info("证书清理完成，耗时: %v [标识符: %s]", duration, c.uniqueIdentifier)
	}()

	return c.ForceCleanupAll()
}

// CreateExportOptionsPlist 创建导出选项plist文件
func (c *CertificateManagerImpl) CreateExportOptionsPlist() (string, error) {
	if c.iosConfig == nil {
		return "", fmt.Errorf("iOS配置为空")
	}

	exportOptions := map[string]interface{}{
		"method":        "app-store",
		"teamID":        c.iosConfig.TeamID,
		"uploadSymbols": false,
	}

	// 如果指定了Bundle ID，添加签名配置
	if c.iosConfig.BundleID != "" {
		exportOptions["signingStyle"] = "manual"
		exportOptions["provisioningProfiles"] = map[string]string{
			c.iosConfig.BundleID: c.GetUniqueIdentifier(),
		}
	}

	// 使用标识符创建临时plist文件
	plistContent := c.dictToPlist(exportOptions)
	plistFileName := fmt.Sprintf("export_options_%s.plist", c.uniqueIdentifier)
	c.tempPlistPath = filepath.Join(c.projectRoot, "build", plistFileName)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(c.tempPlistPath), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(c.tempPlistPath, []byte(plistContent), 0644); err != nil {
		return "", fmt.Errorf("写入plist文件失败: %w", err)
	}

	return c.tempPlistPath, nil
}

// GetUniqueIdentifier 获取唯一标识符
func (c *CertificateManagerImpl) GetUniqueIdentifier() string {
	return c.uniqueIdentifier
}

// ForceCleanupAll 强制清理所有资源
func (c *CertificateManagerImpl) ForceCleanupAll() error {
	logger.Info("强制清理所有资源 [标识符: %s]", c.uniqueIdentifier)

	var errors []error

	// 清理临时钥匙串
	if c.tempKeychainPath != "" {
		if err := c.cleanupKeychain(); err != nil {
			errors = append(errors, err)
		}
	}

	// 清理描述文件
	if c.installedPPPath != "" {
		if err := c.cleanupProvisioningProfile(); err != nil {
			errors = append(errors, err)
		}
	}

	// 清理临时plist文件
	if c.tempPlistPath != "" {
		if err := os.Remove(c.tempPlistPath); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("删除临时plist文件失败: %w", err))
		}
	}

	// 从注册表中移除
	if c.cleanupRegistry != nil {
		c.cleanupRegistry.Cleanup(c.uniqueIdentifier)
	}

	if len(errors) > 0 {
		return fmt.Errorf("清理过程中发生错误: %v", errors)
	}

	logger.Success("资源清理完成 [标识符: %s]", c.uniqueIdentifier)
	return nil
}

// 私有方法实现
func (c *CertificateManagerImpl) setupTemporaryKeychain() error {
	// 使用标识符生成钥匙串名称
	keychainName := fmt.Sprintf("flutter_%s.keychain", c.uniqueIdentifier)
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("获取当前用户失败: %w", err)
	}

	c.tempKeychainPath = filepath.Join(currentUser.HomeDir, "Library", "Keychains", keychainName)
	keychainPassword := c.iosConfig.CertPassword

	logger.Info("创建临时钥匙串: %s", keychainName)

	// 创建钥匙串
	createCmd := []string{"security", "create-keychain", "-p", keychainPassword, c.tempKeychainPath}
	if err := c.executor.RunCommand(createCmd, c.projectRoot); err != nil {
		return fmt.Errorf("创建钥匙串失败: %w", err)
	}

	// 解锁钥匙串
	unlockCmd := []string{"security", "unlock-keychain", "-p", keychainPassword, c.tempKeychainPath}
	if err := c.executor.RunCommand(unlockCmd, c.projectRoot); err != nil {
		return fmt.Errorf("解锁钥匙串失败: %w", err)
	}

	// 设置钥匙串搜索列表
	listCmd := []string{"security", "list-keychains", "-d", "user"}
	output, err := c.executor.RunCommandWithOutput(listCmd, c.projectRoot)
	if err != nil {
		return fmt.Errorf("获取钥匙串列表失败: %w", err)
	}

	// 解析当前钥匙串列表
	currentKeychains := parseKeychainList(output)
	if len(currentKeychains) > 0 {
		logger.Info("当前钥匙串列表:\n%s", strings.Join(currentKeychains, "\n"))
	}

	// 添加临时钥匙串到搜索列表
	newKeychains := append([]string{c.tempKeychainPath}, currentKeychains...)
	setCmd := []string{"security", "list-keychains", "-d", "user", "-s"}
	setCmd = append(setCmd, newKeychains...)

	if err := c.executor.RunCommand(setCmd, c.projectRoot); err != nil {
		return fmt.Errorf("设置钥匙串搜索列表失败: %w", err)
	}

	// 导入P12证书
	logger.Info("导入P12证书到临时钥匙串...")
	importCmd := []string{"security", "import", c.iosConfig.P12Cert, "-k", c.tempKeychainPath, "-P", c.iosConfig.CertPassword, "-T", "/usr/bin/codesign"}
	if err := c.executor.RunCommand(importCmd, c.projectRoot); err != nil {
		return fmt.Errorf("导入P12证书失败: %w", err)
	}

	// 设置证书访问权限
	partitionCmd := []string{"security", "set-key-partition-list", "-S", "apple-tool:,apple:", "-s", "-k", keychainPassword, c.tempKeychainPath}
	if err := c.executor.RunCommand(partitionCmd, c.projectRoot); err != nil {
		return fmt.Errorf("设置证书访问权限失败: %w", err)
	}

	logger.Success("P12证书导入成功")
	return nil
}

func (c *CertificateManagerImpl) installProvisioningProfile() error {
	ppPath := c.iosConfig.ProvisioningProfile
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("获取当前用户失败: %w", err)
	}

	ppDir := filepath.Join(currentUser.HomeDir, "Library", "MobileDevice", "Provisioning Profiles")
	if err := os.MkdirAll(ppDir, 0755); err != nil {
		return fmt.Errorf("创建描述文件目录失败: %w", err)
	}

	// 使用标识符命名描述文件
	targetFileName := fmt.Sprintf("%s.mobileprovision", c.uniqueIdentifier)
	c.installedPPPath = filepath.Join(ppDir, targetFileName)
	if err := copyFile(ppPath, c.installedPPPath); err != nil {
		return fmt.Errorf("复制描述文件失败: %w", err)
	}

	logger.Success("描述文件安装成功: %s", targetFileName)
	return nil
}

// setupSignalHandler 设置信号处理器
func (c *CertificateManagerImpl) setupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Warning("接收到终止信号，正在清理资源...")
		c.ForceCleanupAll()
		os.Exit(1)
	}()
}

// registerCleanupResources 注册清理资源
func (c *CertificateManagerImpl) registerCleanupResources() error {
	if c.cleanupRegistered {
		return nil
	}

	var resources []types.CleanupResource

	// 注册钥匙串资源
	if c.iosConfig.P12Cert != "" {
		keychainName := fmt.Sprintf("flutter_%s.keychain", c.uniqueIdentifier)
		currentUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("获取当前用户失败: %w", err)
		}
		keychainPath := filepath.Join(currentUser.HomeDir, "Library", "Keychains", keychainName)

		resources = append(resources, types.CleanupResource{
			Type:        types.ResourceKeychain,
			Path:        keychainPath,
			Description: "临时钥匙串",
		})
	}

	// 注册描述文件资源
	if c.iosConfig.ProvisioningProfile != "" {
		currentUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("获取当前用户失败: %w", err)
		}
		ppPath := filepath.Join(currentUser.HomeDir, "Library", "MobileDevice", "Provisioning Profiles", fmt.Sprintf("%s.mobileprovision", c.uniqueIdentifier))

		resources = append(resources, types.CleanupResource{
			Type:        types.ResourceProvisioningProfile,
			Path:        ppPath,
			Description: "描述文件",
		})
	}

	// 注册 plist 文件资源
	plistPath := filepath.Join(c.projectRoot, "build", fmt.Sprintf("export_options_%s.plist", c.uniqueIdentifier))
	resources = append(resources, types.CleanupResource{
		Type:        types.ResourcePlistFile,
		Path:        plistPath,
		Description: "导出选项文件",
	})

	// 注册到清理注册表
	if err := c.cleanupRegistry.Register(c.uniqueIdentifier, resources); err != nil {
		return fmt.Errorf("注册清理资源失败: %w", err)
	}

	c.cleanupRegistered = true
	return nil
}

// cleanupKeychain 清理钥匙串
func (c *CertificateManagerImpl) cleanupKeychain() error {
	// 检查钥匙串文件是否存在
	if _, err := os.Stat(c.tempKeychainPath); os.IsNotExist(err) {
		return nil // 文件不存在，无需清理
	}

	// 使用 security 命令删除钥匙串
	cleanupCmd := []string{"security", "delete-keychain", c.tempKeychainPath}
	if err := c.executor.RunCommand(cleanupCmd, c.projectRoot); err != nil {
		logger.Warning("删除钥匙串失败: %s, 错误: %v", c.tempKeychainPath, err)
		// 尝试直接删除文件
		if removeErr := os.Remove(c.tempKeychainPath); removeErr != nil {
			return fmt.Errorf("删除钥匙串文件失败: %w", removeErr)
		}
	}

	logger.Info("已删除钥匙串: %s", c.tempKeychainPath)
	c.tempKeychainPath = ""
	return nil
}

// cleanupProvisioningProfile 清理描述文件
func (c *CertificateManagerImpl) cleanupProvisioningProfile() error {
	// 检查描述文件是否存在
	if _, err := os.Stat(c.installedPPPath); os.IsNotExist(err) {
		return nil // 文件不存在，无需清理
	}

	// 直接删除描述文件
	if err := os.Remove(c.installedPPPath); err != nil {
		return fmt.Errorf("删除描述文件失败: %w", err)
	}

	logger.Info("已删除描述文件: %s", c.installedPPPath)
	c.installedPPPath = ""
	return nil
}

func (c *CertificateManagerImpl) dictToPlist(data map[string]interface{}) string {
	var plistLines []string

	plistLines = append(plistLines,
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`,
		`<plist version="1.0">`,
		`<dict>`,
	)

	for key, value := range data {
		plistLines = append(plistLines, fmt.Sprintf(`    <key>%s</key>`, key))

		switch v := value.(type) {
		case bool:
			if v {
				plistLines = append(plistLines, `    <true/>`)
			} else {
				plistLines = append(plistLines, `    <false/>`)
			}
		case string:
			plistLines = append(plistLines, fmt.Sprintf(`    <string>%s</string>`, v))
		case map[string]string:
			plistLines = append(plistLines, `    <dict>`)
			for subKey, subValue := range v {
				plistLines = append(plistLines, fmt.Sprintf(`        <key>%s</key>`, subKey))
				plistLines = append(plistLines, fmt.Sprintf(`        <string>%s</string>`, subValue))
			}
			plistLines = append(plistLines, `    </dict>`)
		}
	}

	plistLines = append(plistLines, `</dict>`, `</plist>`)
	return strings.Join(plistLines, "\n")
}

func parseKeychainList(output string) []string {
	lines := strings.Split(output, "\n")
	var keychains []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// 移除引号
			line = strings.Trim(line, `"`)
			keychains = append(keychains, line)
		}
	}

	return keychains
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
