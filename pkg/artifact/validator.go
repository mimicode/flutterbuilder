package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mimicode/flutterbuilder/pkg/types"
)

// ArtifactValidatorImpl 产物验证器实现
type ArtifactValidatorImpl struct {
	// 移除logger字段，由于循环依赖问题
}

// NewArtifactValidator 创建新的产物验证器
func NewArtifactValidator() ArtifactValidator {
	return &ArtifactValidatorImpl{}
}

// ValidateArtifact 验证构建产物
func (v *ArtifactValidatorImpl) ValidateArtifact(config *ArtifactConfig) (*ValidationResult, error) {
	if config == nil {
		return nil, fmt.Errorf("验证配置不能为空")
	}

	// 如果禁用验证，直接返回成功
	if config.ValidationConfig != nil && !config.ValidationConfig.EnableValidation {
		return &ValidationResult{
			Success:           true,
			ValidationDetails: []ValidationDetail{{Check: "验证跳过", Status: "success", Message: "产物验证已禁用"}},
		}, nil
	}

	// 获取预期的产物路径
	expectedPaths, err := v.GetExpectedPaths(config.Platform, config.SourcePath, config.IOSConfig)
	if err != nil {
		return nil, fmt.Errorf("获取预期产物路径失败: %w", err)
	}

	// 设置默认大小限制
	v.setDefaultSizeLimits(config)

	// 根据平台执行相应的验证
	switch config.Platform {
	case PlatformAPK:
		return v.validateAndroidArtifacts(expectedPaths, config)
	case PlatformIOS:
		return v.validateIOSArtifacts(expectedPaths, config)
	default:
		return nil, fmt.Errorf("不支持的平台: %s", config.Platform)
	}
}

// GetExpectedPaths 获取预期的产物路径
func (v *ArtifactValidatorImpl) GetExpectedPaths(platform Platform, sourcePath string, iosConfig *types.IOSConfig) ([]string, error) {
	switch platform {
	case PlatformAPK:
		return []string{
			filepath.Join(sourcePath, "build", "app", "outputs", "flutter-apk", "app-release.apk"),
		}, nil
	case PlatformIOS:
		if iosConfig != nil && iosConfig.TeamID != "" {
			// 有证书配置，构建IPA
			ipaDir := filepath.Join(sourcePath, "build", "ios", "ipa")
			return []string{ipaDir}, nil // 返回目录，稍后搜索.ipa文件
		} else {
			// 无证书配置，仅构建iOS App
			return []string{
				filepath.Join(sourcePath, "build", "ios", "iphoneos", "Runner.app"),
			}, nil
		}
	default:
		return nil, fmt.Errorf("不支持的平台: %s", platform)
	}
}

// setDefaultSizeLimits 设置默认大小限制
func (v *ArtifactValidatorImpl) setDefaultSizeLimits(config *ArtifactConfig) {
	// 如果有自定义配置，优先使用自定义值
	if config.ValidationConfig != nil {
		if config.ValidationConfig.CustomMinSize > 0 {
			config.MinFileSize = config.ValidationConfig.CustomMinSize
		}
		if config.ValidationConfig.CustomMaxSize > 0 {
			config.MaxFileSize = config.ValidationConfig.CustomMaxSize
		}
	}

	// 如果仍未设置，使用平台默认值
	if config.MinFileSize == 0 || config.MaxFileSize == 0 {
		switch config.Platform {
		case PlatformAPK:
			if config.MinFileSize == 0 {
				config.MinFileSize = DefaultAndroidMinSize
			}
			if config.MaxFileSize == 0 {
				config.MaxFileSize = DefaultAndroidMaxSize
			}
		case PlatformIOS:
			if config.IOSConfig != nil && config.IOSConfig.TeamID != "" {
				// IPA 文件
				if config.MinFileSize == 0 {
					config.MinFileSize = DefaultIOSIPAMinSize
				}
				if config.MaxFileSize == 0 {
					config.MaxFileSize = DefaultIOSIPAMaxSize
				}
			} else {
				// iOS App 目录
				if config.MinFileSize == 0 {
					config.MinFileSize = DefaultIOSAppMinSize
				}
				if config.MaxFileSize == 0 {
					config.MaxFileSize = DefaultIOSAppMinSize * 10 // 500MB 上限
				}
			}
		}
	}
}

// validateAndroidArtifacts 验证Android产物
func (v *ArtifactValidatorImpl) validateAndroidArtifacts(expectedPaths []string, config *ArtifactConfig) (*ValidationResult, error) {
	apkPath := expectedPaths[0]
	return v.ValidateAPK(apkPath, config)
}

// validateIOSArtifacts 验证iOS产物
func (v *ArtifactValidatorImpl) validateIOSArtifacts(expectedPaths []string, config *ArtifactConfig) (*ValidationResult, error) {
	if config.IOSConfig != nil && config.IOSConfig.TeamID != "" {
		// 验证IPA文件
		ipaDir := expectedPaths[0]
		ipaPath, err := v.findIPAFile(ipaDir)
		if err != nil {
			return &ValidationResult{
				Success: false,
				Error:   fmt.Errorf("查找IPA文件失败: %w", err),
				ValidationDetails: []ValidationDetail{
					{Check: "IPA文件查找", Status: "failed", Message: err.Error(), Critical: true},
				},
			}, err
		}
		return v.ValidateIPA(ipaPath, config)
	} else {
		// 验证iOS App目录
		appPath := expectedPaths[0]
		return v.ValidateIOSApp(appPath, config)
	}
}

// findIPAFile 在指定目录中查找IPA文件
func (v *ArtifactValidatorImpl) findIPAFile(ipaDir string) (string, error) {
	files, err := os.ReadDir(ipaDir)
	if err != nil {
		return "", fmt.Errorf("读取IPA目录失败: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".ipa") {
			return filepath.Join(ipaDir, file.Name()), nil
		}
	}

	return "", fmt.Errorf("在目录 %s 中未找到.ipa文件", ipaDir)
}

// checkFileExists 检查文件是否存在
func (v *ArtifactValidatorImpl) checkFileExists(path string) (bool, os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, info, nil
}

// checkFileSize 检查文件大小是否在合理范围内
func (v *ArtifactValidatorImpl) checkFileSize(fileSize, minSize, maxSize int64) (bool, string) {
	if fileSize < minSize {
		return false, fmt.Sprintf("文件过小: %.2f MB，最小要求: %.2f MB",
			float64(fileSize)/(1024*1024), float64(minSize)/(1024*1024))
	}
	if fileSize > maxSize {
		return false, fmt.Sprintf("文件过大: %.2f MB，最大限制: %.2f MB",
			float64(fileSize)/(1024*1024), float64(maxSize)/(1024*1024))
	}
	return true, fmt.Sprintf("文件大小正常: %.2f MB", float64(fileSize)/(1024*1024))
}

// createValidationResult 创建验证结果
func (v *ArtifactValidatorImpl) createValidationResult(success bool, artifactPath string, fileSize int64, details []ValidationDetail, err error) *ValidationResult {
	return &ValidationResult{
		Success:           success,
		ArtifactPath:      artifactPath,
		FileSize:          fileSize,
		ValidationDetails: details,
		Error:             err,
	}
}