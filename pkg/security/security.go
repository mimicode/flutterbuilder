package security

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mimicode/flutterbuilder/pkg/logger"
)

// SecurityChecker 安全配置检查器接口
type SecurityChecker interface {
	CheckAndroidSecurity() error
	CheckIOSSecurity() error
}

// SecurityCheckerImpl 安全配置检查器实现
type SecurityCheckerImpl struct {
	projectRoot string
}

// NewSecurityChecker 创建新的安全配置检查器
func NewSecurityChecker(projectRoot string) SecurityChecker {
	return &SecurityCheckerImpl{
		projectRoot: projectRoot,
	}
}

// CheckAndroidSecurity 检查Android安全配置
func (s *SecurityCheckerImpl) CheckAndroidSecurity() error {
	// 检查ProGuard规则文件
	proguardFile := filepath.Join(s.projectRoot, "android", "app", "proguard-rules.pro")
	if _, err := os.Stat(proguardFile); err == nil {
		logger.Success("ProGuard配置文件存在")
	} else {
		logger.Warning("ProGuard规则文件未找到")
	}

	// 检查签名配置
	keyProperties := filepath.Join(s.projectRoot, "android", "key.properties")
	if _, err := os.Stat(keyProperties); err == nil {
		logger.Success("发布签名配置存在")
	} else {
		logger.Warning("签名配置文件未找到 - 将使用调试签名")
	}

	return nil
}

// CheckIOSSecurity 检查iOS安全配置
func (s *SecurityCheckerImpl) CheckIOSSecurity() error {
	iosProject := filepath.Join(s.projectRoot, "ios", "Runner.xcodeproj")
	if _, err := os.Stat(iosProject); err == nil {
		logger.Success("iOS项目配置存在")
	} else {
		return fmt.Errorf("iOS项目未找到")
	}

	// 检查动态证书配置
	if runtime.GOOS != "darwin" {
		logger.Warning("iOS构建需要macOS环境")
	}

	return nil
}
