package certificates

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/mimicode/flutterbuilder/pkg/executor"
	"github.com/mimicode/flutterbuilder/pkg/logger"
	"github.com/mimicode/flutterbuilder/pkg/types"
)

// CertificateManagerImpl iOS证书管理器实现
type CertificateManagerImpl struct {
	iosConfig        *types.IOSConfig
	projectRoot      string
	executor         executor.CommandExecutor
	tempKeychainPath string
}

// NewCertificateManager 创建新的证书管理器
func NewCertificateManager(iosConfig *types.IOSConfig, projectRoot string) types.CertificateManager {
	return &CertificateManagerImpl{
		iosConfig:   iosConfig,
		projectRoot: projectRoot,
		executor:    executor.NewCommandExecutor(),
	}
}

// SetupCertificates 设置iOS证书
func (c *CertificateManagerImpl) SetupCertificates() error {
	if c.iosConfig == nil {
		return nil
	}

	logger.Info("设置iOS证书和描述文件...")

	// 设置P12证书
	if c.iosConfig.P12Cert != "" && c.iosConfig.CertPassword != "" {
		if err := c.setupTemporaryKeychain(); err != nil {
			return fmt.Errorf("设置临时钥匙串失败: %w", err)
		}
	}

	// 安装描述文件
	if c.iosConfig.ProvisioningProfile != "" {
		if err := c.installProvisioningProfile(); err != nil {
			return fmt.Errorf("安装描述文件失败: %w", err)
		}
	}

	return nil
}

// CleanupCertificates 清理iOS证书配置
func (c *CertificateManagerImpl) CleanupCertificates() error {
	if c.tempKeychainPath == "" {
		return nil
	}

	logger.Info("清理临时钥匙串...")

	// 删除临时钥匙串
	if _, err := os.Stat(c.tempKeychainPath); err == nil {
		cleanupCmd := []string{"security", "delete-keychain", c.tempKeychainPath}
		if err := c.executor.RunCommand(cleanupCmd, c.projectRoot); err != nil {
			logger.Warning("清理临时钥匙串失败: %v", err)
		} else {
			logger.Success("临时钥匙串已删除")
		}
	}

	c.tempKeychainPath = ""
	return nil
}

// CreateExportOptionsPlist 创建导出选项plist文件
func (c *CertificateManagerImpl) CreateExportOptionsPlist() (string, error) {
	if c.iosConfig == nil {
		return "", fmt.Errorf("iOS配置为空")
	}

	exportOptions := map[string]interface{}{
		"method":         "app-store",
		"teamID":         c.iosConfig.TeamID,
		"uploadBitcode":  true,
		"uploadSymbols":  true,
		"compileBitcode": true,
	}

	// 如果指定了Bundle ID，添加签名配置
	if c.iosConfig.BundleID != "" {
		exportOptions["signingStyle"] = "manual"
		exportOptions["provisioningProfiles"] = map[string]string{
			c.iosConfig.BundleID: c.getProvisioningProfileName(),
		}
	}

	// 创建临时plist文件
	plistContent := c.dictToPlist(exportOptions)
	tempPlist := filepath.Join(c.projectRoot, "build", "export_options.plist")

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(tempPlist), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(tempPlist, []byte(plistContent), 0644); err != nil {
		return "", fmt.Errorf("写入plist文件失败: %w", err)
	}

	return tempPlist, nil
}

// 私有方法实现
func (c *CertificateManagerImpl) setupTemporaryKeychain() error {
	// 生成临时钥匙串名称和密码
	keychainName := fmt.Sprintf("flutter_build_%s.keychain", generateRandomString(8))
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("获取当前用户失败: %w", err)
	}

	c.tempKeychainPath = filepath.Join(currentUser.HomeDir, "Library", "Keychains", keychainName)
	keychainPassword := generateRandomString(16)

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

	// 复制描述文件到系统目录
	targetPath := filepath.Join(ppDir, filepath.Base(ppPath))
	if err := copyFile(ppPath, targetPath); err != nil {
		return fmt.Errorf("复制描述文件失败: %w", err)
	}

	logger.Success("描述文件安装成功: %s", filepath.Base(ppPath))
	return nil
}

func (c *CertificateManagerImpl) getProvisioningProfileName() string {
	if c.iosConfig.ProvisioningProfile == "" {
		return ""
	}
	return strings.TrimSuffix(filepath.Base(c.iosConfig.ProvisioningProfile), filepath.Ext(c.iosConfig.ProvisioningProfile))
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

// 辅助函数
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
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
