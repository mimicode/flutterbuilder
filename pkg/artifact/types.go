package artifact

import (
	"github.com/mimicode/flutterbuilder/pkg/types"
)

// Platform 构建平台
type Platform string

const (
	PlatformAPK Platform = "apk"
	PlatformIOS Platform = "ios"
)

// ArtifactValidationConfig 产物验证配置
type ArtifactValidationConfig struct {
	EnableValidation     bool  // 是否启用产物验证（默认: true）
	EnableIntegrityCheck bool  // 是否启用完整性检查（默认: true）
	CustomMinSize        int64 // 自定义最小文件大小（0表示使用默认值）
	CustomMaxSize        int64 // 自定义最大文件大小（0表示使用默认值）
}

// ArtifactConfig 产物验证配置
type ArtifactConfig struct {
	Platform          Platform                  // 构建平台
	SourcePath        string                    // 项目根路径
	IOSConfig         *types.IOSConfig          // iOS配置（可选）
	ExpectedPaths     []string                  // 预期的产物路径
	MinFileSize       int64                     // 最小文件大小
	MaxFileSize       int64                     // 最大文件大小
	ValidateIntegrity bool                      // 是否验证完整性
	ValidationConfig  *ArtifactValidationConfig // 验证配置（可选）
}

// ValidationResult 验证结果
type ValidationResult struct {
	Success           bool                // 验证是否成功
	ArtifactPath      string              // 产物文件路径
	FileSize          int64               // 文件大小
	ValidationDetails []ValidationDetail  // 验证详情
	Error             error               // 错误信息
}

// ValidationDetail 验证详情
type ValidationDetail struct {
	Check    string // 检查项名称
	Status   string // 检查状态 (success, failed, warning)
	Message  string // 详细信息
	Critical bool   // 是否为关键检查项
}

// 默认配置常量
const (
	// Android APK 默认大小限制
	DefaultAndroidMinSize = 5 * 1024 * 1024       // 5MB
	DefaultAndroidMaxSize = 2 * 1024 * 1024 * 1024 // 2GB

	// iOS App 默认大小限制
	DefaultIOSAppMinSize = 50 * 1024 * 1024 // 50MB

	// iOS IPA 默认大小限制
	DefaultIOSIPAMinSize = 10 * 1024 * 1024      // 10MB
	DefaultIOSIPAMaxSize = 4 * 1024 * 1024 * 1024 // 4GB
)

// GetDefaultValidationConfig 获取默认验证配置
func GetDefaultValidationConfig() *ArtifactValidationConfig {
	return &ArtifactValidationConfig{
		EnableValidation:     true,
		EnableIntegrityCheck: true,
		CustomMinSize:        0, // 使用平台默认值
		CustomMaxSize:        0, // 使用平台默认值
	}
}

// ArtifactValidator 产物验证器接口
type ArtifactValidator interface {
	// ValidateArtifact 验证构建产物
	ValidateArtifact(config *ArtifactConfig) (*ValidationResult, error)

	// GetExpectedPaths 获取预期的产物路径
	GetExpectedPaths(platform Platform, sourcePath string, iosConfig *types.IOSConfig) ([]string, error)

	// ValidateAPK 验证Android APK文件
	ValidateAPK(apkPath string, config *ArtifactConfig) (*ValidationResult, error)

	// ValidateIPA 验证iOS IPA文件
	ValidateIPA(ipaPath string, config *ArtifactConfig) (*ValidationResult, error)

	// ValidateIOSApp 验证iOS App目录
	ValidateIOSApp(appPath string, config *ArtifactConfig) (*ValidationResult, error)
}