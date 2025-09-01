package api

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/mimicode/flutterbuilder/pkg/builder"
	"github.com/mimicode/flutterbuilder/pkg/hooks"
	"github.com/mimicode/flutterbuilder/pkg/logger"
	"github.com/mimicode/flutterbuilder/pkg/types"
)

// Platform 构建平台
type Platform = builder.Platform

const (
	PlatformAPK = builder.PlatformAPK
	PlatformIOS = builder.PlatformIOS
)

// IOSConfig iOS构建配置
type IOSConfig = types.IOSConfig

// BuildConfig 构建配置
type BuildConfig struct {
	Platform    Platform               // 构建平台
	SourcePath  string                 // Flutter项目源代码路径
	IOSConfig   *IOSConfig             // iOS配置（可选）
	CustomArgs  map[string]interface{} // 自定义构建参数
	HooksConfig *hooks.HooksConfig     // 钩子配置（可选）
	Logger      Logger                 // 日志接口（可选）
	Verbose     bool                   // 是否显示详细日志
}

// BuildResult 构建结果
type BuildResult struct {
	Success    bool          // 是否成功
	Platform   Platform      // 构建平台
	BuildTime  time.Duration // 构建耗时
	OutputPath string        // 输出路径
	Error      error         // 错误信息
}

// Logger 日志接口
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
	Success(format string, args ...interface{})
	Header(title string)
	Println(args ...interface{})
	Printf(format string, args ...interface{})
}

// FlutterBuilder Flutter构建器公开接口
type FlutterBuilder interface {
	// Build 执行构建
	Build(config *BuildConfig) (*BuildResult, error)

	// SetLogger 设置日志接口
	SetLogger(logger Logger)

	// Validate 验证配置
	Validate(config *BuildConfig) error
}

// flutterBuilderImpl 内部实现
type flutterBuilderImpl struct {
	customLogger Logger
}

// NewFlutterBuilder 创建新的Flutter构建器
func NewFlutterBuilder() FlutterBuilder {
	return &flutterBuilderImpl{}
}

// SetLogger 设置自定义日志接口
func (fb *flutterBuilderImpl) SetLogger(customLogger Logger) {
	fb.customLogger = customLogger
}

// Validate 验证构建配置
func (fb *flutterBuilderImpl) Validate(config *BuildConfig) error {
	if config == nil {
		return fmt.Errorf("构建配置不能为空")
	}

	if config.SourcePath == "" {
		return fmt.Errorf("源代码路径不能为空")
	}

	if config.Platform != PlatformAPK && config.Platform != PlatformIOS {
		return fmt.Errorf("不支持的平台: %s", config.Platform)
	}

	return nil
}

// Build 执行构建
func (fb *flutterBuilderImpl) Build(config *BuildConfig) (*BuildResult, error) {
	startTime := time.Now()

	// 验证配置
	if err := fb.Validate(config); err != nil {
		return &BuildResult{
			Success:   false,
			Platform:  config.Platform,
			BuildTime: time.Since(startTime),
			Error:     err,
		}, err
	}

	// 设置日志接口
	if config.Logger != nil {
		logger.SetExternalLogger(config.Logger)
	} else if fb.customLogger != nil {
		logger.SetExternalLogger(fb.customLogger)
	}

	// 设置日志级别
	if config.Verbose {
		logger.SetLevel(logger.DebugLevel)
	}

	// 创建内部构建器
	var internalBuilder builder.FlutterBuilder
	if config.Platform == PlatformIOS {
		internalBuilder = builder.NewFlutterBuilder("ios", config.IOSConfig, config.SourcePath)
	} else {
		internalBuilder = builder.NewFlutterBuilder("apk", nil, config.SourcePath)
	}

	// 如果有自定义参数，需要传递给内部构建器
	if len(config.CustomArgs) > 0 {
		if customBuilder, ok := internalBuilder.(interface{ SetCustomArgs(map[string]interface{}) }); ok {
			customBuilder.SetCustomArgs(config.CustomArgs)
		}
	}

	// 如果有钩子配置，需要传递给内部构建器
	if config.HooksConfig != nil {
		if hooksBuilder, ok := internalBuilder.(interface {
			SetHooks(*hooks.HooksConfig) error
		}); ok {
			if err := hooksBuilder.SetHooks(config.HooksConfig); err != nil {
				return &BuildResult{
					Success:   false,
					Platform:  config.Platform,
					BuildTime: time.Since(startTime),
					Error:     fmt.Errorf("设置钩子配置失败: %w", err),
				}, fmt.Errorf("设置钩子配置失败: %w", err)
			}
		}
	}

	// 执行构建
	err := internalBuilder.Run()
	buildTime := time.Since(startTime)

	result := &BuildResult{
		Success:   err == nil,
		Platform:  config.Platform,
		BuildTime: buildTime,
	}

	if err != nil {
		result.Error = err
		return result, err
	}

	// 获取输出路径
	result.OutputPath = getOutputPath(config.Platform, config.SourcePath, config.IOSConfig)

	return result, nil
}

// getOutputPath 获取构建输出路径
func getOutputPath(platform Platform, sourcePath string, iosConfig *IOSConfig) string {
	switch platform {
	case PlatformAPK:
		return fmt.Sprintf("%s/build/app/outputs/flutter-apk/app-release.apk", sourcePath)
	case PlatformIOS:
		// 如果提供了证书配置，返回IPA文件路径；否则返回构建目录
		if iosConfig != nil && iosConfig.TeamID != "" {
			// 构建IPA，尝试获取实际的IPA文件路径
			return getActualIPAPath(sourcePath)
		} else {
			// 仅构建iOS，返回.app文件路径
			return fmt.Sprintf("%s/build/ios/iphoneos/Runner.app", sourcePath)
		}
	default:
		return ""
	}
}

// getActualIPAPath 获取实际生成的IPA文件路径
func getActualIPAPath(sourcePath string) string {
	ipaDir := fmt.Sprintf("%s/build/ios/ipa", sourcePath)

	// 尝试读取目录中的IPA文件
	if files, err := filepath.Glob(fmt.Sprintf("%s/*.ipa", ipaDir)); err == nil && len(files) > 0 {
		// 返回找到的第一个IPA文件路径
		return files[0]
	}

	// 如果找不到实际文件，返回默认的IPA路径格式
	// 通常Flutter会生成以项目名命名的IPA文件
	return fmt.Sprintf("%s/Runner.ipa", ipaDir)
}

// QuickBuild 快速构建函数（便捷方法）
func QuickBuild(platform Platform, sourcePath string) (*BuildResult, error) {
	builder := NewFlutterBuilder()
	config := &BuildConfig{
		Platform:   platform,
		SourcePath: sourcePath,
	}
	return builder.Build(config)
}

// QuickBuildWithHooks 带钩子的快速构建（便捷方法）
func QuickBuildWithHooks(platform Platform, sourcePath string, hooksConfig *hooks.HooksConfig) (*BuildResult, error) {
	builder := NewFlutterBuilder()
	config := &BuildConfig{
		Platform:    platform,
		SourcePath:  sourcePath,
		HooksConfig: hooksConfig,
	}
	return builder.Build(config)
}

// QuickBuildAPK 快速构建APK（便捷方法）
func QuickBuildAPK(sourcePath string) (*BuildResult, error) {
	return QuickBuild(PlatformAPK, sourcePath)
}

// QuickBuildAPKWithHooks 带钩子的快速构建 APK（便捷方法）
func QuickBuildAPKWithHooks(sourcePath string, hooksConfig *hooks.HooksConfig) (*BuildResult, error) {
	return QuickBuildWithHooks(PlatformAPK, sourcePath, hooksConfig)
}

// QuickBuildIOS 快速构建iOS（便捷方法）
func QuickBuildIOS(sourcePath string, iosConfig *IOSConfig) (*BuildResult, error) {
	builder := NewFlutterBuilder()
	config := &BuildConfig{
		Platform:   PlatformIOS,
		SourcePath: sourcePath,
		IOSConfig:  iosConfig,
	}
	return builder.Build(config)
}

// QuickBuildIOSWithHooks 带钩子的快速构建 iOS（便捷方法）
func QuickBuildIOSWithHooks(sourcePath string, iosConfig *IOSConfig, hooksConfig *hooks.HooksConfig) (*BuildResult, error) {
	builder := NewFlutterBuilder()
	config := &BuildConfig{
		Platform:    PlatformIOS,
		SourcePath:  sourcePath,
		IOSConfig:   iosConfig,
		HooksConfig: hooksConfig,
	}
	return builder.Build(config)
}
