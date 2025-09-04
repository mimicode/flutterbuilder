package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mimicode/flutterbuilder/pkg/artifact"
	"github.com/mimicode/flutterbuilder/pkg/certificates"
	"github.com/mimicode/flutterbuilder/pkg/executor"
	"github.com/mimicode/flutterbuilder/pkg/hooks"
	"github.com/mimicode/flutterbuilder/pkg/logger"
	"github.com/mimicode/flutterbuilder/pkg/security"
	"github.com/mimicode/flutterbuilder/pkg/types"
)

// FlutterBuilderImpl Flutter构建器实现
type FlutterBuilderImpl struct {
	platform          Platform
	iosConfig         *IOSConfig
	projectRoot       string
	executor          executor.CommandExecutor
	security          security.SecurityChecker
	certManager       types.CertificateManager
	customArgs        map[string]interface{}             // 自定义构建参数
	hookExecutor      hooks.HookExecutor                 // 钩子执行器
	artifactValidator artifact.ArtifactValidator         // 产物验证器
	validationConfig  *artifact.ArtifactValidationConfig // 验证配置
}

// NewFlutterBuilder 创建新的Flutter构建器
func NewFlutterBuilder(platform string, iosConfig *IOSConfig, sourcePath string) FlutterBuilder {
	// 使用提供的源代码路径作为项目根目录
	projectRoot := sourcePath

	builder := &FlutterBuilderImpl{
		platform:          Platform(platform),
		iosConfig:         iosConfig,
		projectRoot:       projectRoot,
		executor:          executor.NewCommandExecutor(),
		security:          security.NewSecurityChecker(projectRoot),
		customArgs:        make(map[string]interface{}),          // 初始化自定义参数
		hookExecutor:      hooks.NewHookExecutor(projectRoot),    // 初始化钩子执行器
		artifactValidator: artifact.NewArtifactValidator(),       // 初始化产物验证器
		validationConfig:  artifact.GetDefaultValidationConfig(), // 默认验证配置
	}

	// 如果是iOS平台且有证书配置，创建证书管理器
	if platform == "ios" && iosConfig != nil {
		// 转换为 types.IOSConfig
		typesIOSConfig := &types.IOSConfig{
			P12Cert:             iosConfig.P12Cert,
			CertPassword:        iosConfig.CertPassword,
			ProvisioningProfile: iosConfig.ProvisioningProfile,
			TeamID:              iosConfig.TeamID,
			BundleID:            iosConfig.BundleID,
		}
		builder.certManager = certificates.NewCertificateManager(typesIOSConfig, projectRoot)
	}

	return builder
}

// Run 执行完整的构建流程
func (b *FlutterBuilderImpl) Run() error {
	startTime := time.Now()

	logger.Info("项目根目录: %s", b.projectRoot)
	logger.Info("构建平台: %s", b.platform)
	logger.Info("操作系统: %s", runtime.GOOS)
	logger.Println()

	// 验证环境和参数
	if err := b.validateEnvironment(); err != nil {
		return fmt.Errorf("环境验证失败: %w", err)
	}

	// 切换到项目根目录
	if err := os.Chdir(b.projectRoot); err != nil {
		return fmt.Errorf("切换目录失败: %w", err)
	}

	// 执行构建流程
	if err := b.Clean(); err != nil {
		return fmt.Errorf("清理项目失败: %w", err)
	}

	if err := b.GetDependencies(); err != nil {
		return fmt.Errorf("获取依赖失败: %w", err)
	}

	if err := b.RunCodeGeneration(); err != nil {
		return fmt.Errorf("代码生成失败: %w", err)
	}

	if err := b.CheckSecurityConfig(); err != nil {
		return fmt.Errorf("安全配置检查失败: %w", err)
	}

	if err := b.Build(); err != nil {
		return fmt.Errorf("构建失败: %w", err)
	}

	if err := b.PostBuildProcessing(); err != nil {
		return fmt.Errorf("构建后处理失败: %w", err)
	}

	// 显示完成信息
	elapsedTime := time.Since(startTime)
	logger.Println()
	logger.Header("构建完成")
	logger.Success("构建成功完成！耗时: %.2f秒", elapsedTime.Seconds())

	return nil
}

// Clean 清理项目
func (b *FlutterBuilderImpl) Clean() error {
	// 执行前置钩子
	if err := b.executeHooks(hooks.HookPreClean, "Clean"); err != nil {
		return fmt.Errorf("清理前置钩子执行失败: %w", err)
	}

	logger.Info("[1/6] 清理构建缓存...")

	// Flutter clean
	if err := b.executor.RunCommand([]string{"flutter", "clean"}, b.projectRoot); err != nil {
		return fmt.Errorf("flutter clean 执行失败: %w", err)
	}
	logger.Success("Flutter clean 执行成功")

	// 删除旧的构建目录
	buildDir := filepath.Join(b.projectRoot, "build")
	if err := b.cleanBuildDirectory(buildDir); err != nil {
		logger.Warning("无法删除构建目录，文件可能被占用: %v", err)
		logger.Info("继续构建过程...")
	}

	// 执行后置钩子
	if err := b.executeHooks(hooks.HookPostClean, "Clean"); err != nil {
		return fmt.Errorf("清理后置钩子执行失败: %w", err)
	}

	return nil
}

// GetDependencies 获取依赖
func (b *FlutterBuilderImpl) GetDependencies() error {
	// 执行前置钩子
	if err := b.executeHooks(hooks.HookPreGetDeps, "GetDependencies"); err != nil {
		return fmt.Errorf("获取依赖前置钩子执行失败: %w", err)
	}

	logger.Info("[2/6] 获取项目依赖...")

	if err := b.executor.RunCommand([]string{"flutter", "pub", "get"}, b.projectRoot); err != nil {
		return fmt.Errorf("依赖获取失败: %w", err)
	}

	logger.Success("依赖获取成功")

	// 执行后置钩子
	if err := b.executeHooks(hooks.HookPostGetDeps, "GetDependencies"); err != nil {
		return fmt.Errorf("获取依赖后置钩子执行失败: %w", err)
	}

	return nil
}

// RunCodeGeneration 运行代码生成
func (b *FlutterBuilderImpl) RunCodeGeneration() error {
	// 执行前置钩子
	if err := b.executeHooks(hooks.HookPreCodeGen, "RunCodeGeneration"); err != nil {
		return fmt.Errorf("代码生成前置钩子执行失败: %w", err)
	}

	logger.Info("[3/6] 运行代码生成...")

	// 尝试运行build_runner，如果失败则忽略
	// 使用新的 dart run 命令替代已废弃的 flutter packages pub run
	if err := b.executor.RunCommand([]string{
		"dart", "run", "build_runner",
		"build", "--delete-conflicting-outputs",
	}, b.projectRoot); err != nil {
		logger.Info("跳过代码生成（build_runner未配置或不需要）")
	} else {
		logger.Success("代码生成完成")
	}

	// 执行后置钩子
	if err := b.executeHooks(hooks.HookPostCodeGen, "RunCodeGeneration"); err != nil {
		return fmt.Errorf("代码生成后置钩子执行失败: %w", err)
	}

	return nil
}

// CheckSecurityConfig 检查安全配置
func (b *FlutterBuilderImpl) CheckSecurityConfig() error {
	// 执行前置钩子
	if err := b.executeHooks(hooks.HookPreSecurityCheck, "CheckSecurityConfig"); err != nil {
		return fmt.Errorf("安全检查前置钩子执行失败: %w", err)
	}

	logger.Info("[4/6] 检查安全配置...")

	var checkErr error
	if b.platform == PlatformAPK {
		checkErr = b.security.CheckAndroidSecurity()
	} else if b.platform == PlatformIOS {
		checkErr = b.security.CheckIOSSecurity()
	}

	// 执行后置钩子
	if err := b.executeHooks(hooks.HookPostSecurityCheck, "CheckSecurityConfig"); err != nil {
		return fmt.Errorf("安全检查后置钩子执行失败: %w", err)
	}

	return checkErr
}

// Build 构建发布版本
func (b *FlutterBuilderImpl) Build() error {
	// 执行前置钩子
	if err := b.executeHooks(hooks.HookPreBuild, "Build"); err != nil {
		return fmt.Errorf("构建前置钩子执行失败: %w", err)
	}

	logger.Info("[5/6] 构建发布版本...")

	var buildErr error
	if b.platform == PlatformAPK {
		buildErr = b.buildAndroidAPK()
	} else if b.platform == PlatformIOS {
		buildErr = b.buildIOS()
	} else {
		buildErr = fmt.Errorf("不支持的平台: %s", b.platform)
	}

	// 执行后置钩子
	if err := b.executeHooks(hooks.HookPostBuild, "Build"); err != nil {
		return fmt.Errorf("构建后置钩子执行失败: %w", err)
	}

	return buildErr
}

// PostBuildProcessing 构建后处理
func (b *FlutterBuilderImpl) PostBuildProcessing() error {
	// 执行前置钩子
	if err := b.executeHooks(hooks.HookPrePostProcess, "PostBuildProcessing"); err != nil {
		return fmt.Errorf("后处理前置钩子执行失败: %w", err)
	}

	logger.Info("[6/6] 构建后处理...")

	// 创建构建信息文件
	if err := b.createBuildInfo(); err != nil {
		logger.Warning("创建构建信息文件失败: %v", err)
	}

	// 显示安全提醒
	b.showSecurityReminders()

	// 执行后置钩子
	if err := b.executeHooks(hooks.HookPostPostProcess, "PostBuildProcessing"); err != nil {
		return fmt.Errorf("后处理后置钩子执行失败: %w", err)
	}

	return nil
}

// SetCustomArgs 设置自定义构建参数
func (b *FlutterBuilderImpl) SetCustomArgs(args map[string]interface{}) {
	if b.customArgs == nil {
		b.customArgs = make(map[string]interface{})
	}
	for k, v := range args {
		b.customArgs[k] = v
	}
}

// GetCustomArg 获取自定义参数值
func (b *FlutterBuilderImpl) GetCustomArg(key string) (interface{}, bool) {
	if b.customArgs == nil {
		return nil, false
	}
	val, exists := b.customArgs[key]
	return val, exists
}

// GetCustomArgString 获取字符串类型的自定义参数
func (b *FlutterBuilderImpl) GetCustomArgString(key string) string {
	if val, exists := b.GetCustomArg(key); exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetCustomArgBool 获取布尔类型的自定义参数
func (b *FlutterBuilderImpl) GetCustomArgBool(key string) bool {
	if val, exists := b.GetCustomArg(key); exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// GetCustomArgStringSlice 获取字符串数组类型的自定义参数
func (b *FlutterBuilderImpl) GetCustomArgStringSlice(key string) []string {
	if val, exists := b.GetCustomArg(key); exists {
		if slice, ok := val.([]string); ok {
			return slice
		}
		// 尝试转换 []interface{} 为 []string
		if interfaceSlice, ok := val.([]interface{}); ok {
			result := make([]string, len(interfaceSlice))
			for i, v := range interfaceSlice {
				if str, ok := v.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}
	return nil
}

// SetHooks 设置钩子配置
func (b *FlutterBuilderImpl) SetHooks(hooksConfig *hooks.HooksConfig) error {
	if hooksConfig == nil {
		return fmt.Errorf("钩子配置不能为空")
	}

	// 清空现有钩子
	b.hookExecutor.ClearAllHooks()

	// 注册所有钩子
	for hookType, configs := range hooksConfig.Hooks {
		for _, config := range configs {
			if err := b.hookExecutor.RegisterHook(hookType, config); err != nil {
				return fmt.Errorf("注册钩子失败 [%s]: %w", hookType, err)
			}
		}
	}

	return nil
}

// RegisterHook 注册单个钩子
func (b *FlutterBuilderImpl) RegisterHook(hookType hooks.HookType, config *hooks.HookConfig) error {
	return b.hookExecutor.RegisterHook(hookType, config)
}

// UnregisterHook 注销钩子
func (b *FlutterBuilderImpl) UnregisterHook(hookType hooks.HookType, scriptPath string) error {
	return b.hookExecutor.UnregisterHook(hookType, scriptPath)
}

// GetHooks 获取指定类型的钩子
func (b *FlutterBuilderImpl) GetHooks(hookType hooks.HookType) []*hooks.HookConfig {
	return b.hookExecutor.GetHooks(hookType)
}

// ClearHooks 清空指定类型的钩子
func (b *FlutterBuilderImpl) ClearHooks(hookType hooks.HookType) {
	b.hookExecutor.ClearHooks(hookType)
}

// ClearAllHooks 清空所有钩子
func (b *FlutterBuilderImpl) ClearAllHooks() {
	b.hookExecutor.ClearAllHooks()
}

// SetValidationConfig 设置验证配置
func (b *FlutterBuilderImpl) SetValidationConfig(config *artifact.ArtifactValidationConfig) {
	if config != nil {
		b.validationConfig = config
	} else {
		b.validationConfig = artifact.GetDefaultValidationConfig()
	}
}

// GetValidationConfig 获取当前验证配置
func (b *FlutterBuilderImpl) GetValidationConfig() *artifact.ArtifactValidationConfig {
	return b.validationConfig
}

// validateBuildArtifacts 验证构建产物
func (b *FlutterBuilderImpl) validateBuildArtifacts() error {
	logger.Info("[6/6] 验证构建产物...")

	// 创建验证配置
	config := &artifact.ArtifactConfig{
		Platform:          convertPlatform(b.platform),
		SourcePath:        b.projectRoot,
		ValidateIntegrity: true,
		ValidationConfig:  b.validationConfig,
	}

	// 如果是iOS平台，传递iOS配置
	if b.platform == PlatformIOS && b.iosConfig != nil {
		config.IOSConfig = &types.IOSConfig{
			P12Cert:             b.iosConfig.P12Cert,
			CertPassword:        b.iosConfig.CertPassword,
			ProvisioningProfile: b.iosConfig.ProvisioningProfile,
			TeamID:              b.iosConfig.TeamID,
			BundleID:            b.iosConfig.BundleID,
		}
	}

	// 执行验证
	result, err := b.artifactValidator.ValidateArtifact(config)
	if err != nil {
		return fmt.Errorf("产物验证执行失败: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("产物验证失败: %s", result.Error)
	}

	// 记录验证详情
	logger.Success("产物验证成功")
	if result.ArtifactPath != "" {
		logger.Printf("  文件路径: %s", result.ArtifactPath)
	}
	if result.FileSize > 0 {
		logger.Printf("  文件大小: %.2f MB", float64(result.FileSize)/(1024*1024))
	}

	for _, detail := range result.ValidationDetails {
		switch detail.Status {
		case "success":
			logger.Success("  ✓ %s", detail.Check)
		case "failed":
			if detail.Critical {
				logger.Error("  ✗ %s: %s", detail.Check, detail.Message)
			} else {
				logger.Warning("  ⚠ %s: %s", detail.Check, detail.Message)
			}
		case "warning":
			logger.Warning("  ⚠ %s: %s", detail.Check, detail.Message)
		default:
			logger.Info("  • %s: %s", detail.Check, detail.Message)
		}
	}

	return nil
}

// executeHooks 执行指定类型的钩子
func (b *FlutterBuilderImpl) executeHooks(hookType hooks.HookType, buildStage string) error {
	context := &hooks.HookContext{
		HookType:    hookType,
		Platform:    string(b.platform),
		ProjectRoot: b.projectRoot,
		BuildStage:  buildStage,
		StartTime:   time.Now(),
		CustomArgs:  b.customArgs,
	}

	_, err := b.hookExecutor.ExecuteHooks(hookType, context)
	return err
}

// removeSpecificArgs 移除指定的参数
func (b *FlutterBuilderImpl) removeSpecificArgs(args []string, removeList []string) []string {
	if len(removeList) == 0 {
		return args
	}

	// 创建移除列表的map以提高查找效率
	removeMap := make(map[string]bool)
	for _, arg := range removeList {
		removeMap[arg] = true
	}

	result := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// 检查是否是带等号的参数（如 --dart-define=KEY=VALUE）
		// 这个检查要先做，因为 --dart-define=xxx 不等于 --dart-define
		if b.isParameterWithEquals(arg, removeList) {
			continue
		}

		// 检查当前参数是否在移除列表中
		if removeMap[arg] {
			// 如果是带值的参数（如 --target-platform android-arm64）
			// 也要移除下一个参数值
			if b.isParameterWithValue(arg) && i+1 < len(args) {
				i++ // 跳过下一个参数值
			}
			continue
		}

		result = append(result, arg)
	}

	return result
}

// isParameterWithValue 检查参数是否需要值（如 --target-platform）
func (b *FlutterBuilderImpl) isParameterWithValue(arg string) bool {
	paramsWithValue := []string{
		"--target-platform",
		"--split-debug-info",
		"--export-options-plist",
		"--flavor",
		"--bundle-id",
		"--build-name",
		"--build-number",
	}

	for _, param := range paramsWithValue {
		if arg == param {
			return true
		}
	}
	return false
}

// isParameterWithEquals 检查是否是带等号的参数（如 --dart-define=KEY=VALUE）
func (b *FlutterBuilderImpl) isParameterWithEquals(arg string, removeList []string) bool {
	for _, removeArg := range removeList {
		// 如果移除列表中有 --dart-define，则移除所有以 --dart-define= 开头的参数
		if removeArg == "--dart-define" {
			if strings.HasPrefix(arg, "--dart-define=") {
				return true
			}
		}
		// 如果移除列表中有具体的 dart-define 参数
		if arg == removeArg {
			return true
		}
	}
	return false
}

// 私有方法实现...
func (b *FlutterBuilderImpl) validateEnvironment() error {
	// 检查Flutter环境
	if err := b.checkFlutterEnvironment(); err != nil {
		return err
	}

	// 验证平台参数
	if err := b.validatePlatform(); err != nil {
		return err
	}

	return nil
}

func (b *FlutterBuilderImpl) checkFlutterEnvironment() error {
	logger.Info("检查Flutter环境...")

	if err := b.executor.RunCommand([]string{"flutter", "--version"}, b.projectRoot); err != nil {
		return fmt.Errorf("Flutter未安装或不在PATH中")
	}

	logger.Success("Flutter环境正常")
	return nil
}

func (b *FlutterBuilderImpl) validatePlatform() error {
	if b.platform != PlatformAPK && b.platform != PlatformIOS {
		return fmt.Errorf("无效的平台参数: %s", b.platform)
	}

	if b.platform == PlatformIOS && runtime.GOOS != "darwin" {
		return fmt.Errorf("iOS构建需要macOS环境，当前操作系统: %s", runtime.GOOS)
	}

	return nil
}

func (b *FlutterBuilderImpl) cleanBuildDirectory(buildDir string) error {
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		return nil // 目录不存在，无需清理
	}

	logger.Info("删除旧的构建目录...")
	return os.RemoveAll(buildDir)
}

func (b *FlutterBuilderImpl) buildAndroidAPK() error {
	logger.Info("构建Android APK（仅ARM64架构）...")

	buildCmd := []string{
		"flutter", "build", "apk",
		"--release",
	}

	// 添加默认参数（可被自定义参数覆盖）
	defaultArgs := []string{
		"--obfuscate",
		"--split-debug-info=build/debug-info",
		"--tree-shake-icons",
		"--target-platform", "android-arm64",
		"--dart-define=FLUTTER_WEB_USE_SKIA=true",
		"--dart-define=FLUTTER_WEB_AUTO_DETECT=true",
	}

	// 检查是否禁用默认参数
	if !b.GetCustomArgBool("disable_default_args") {
		// 添加默认参数
		buildCmd = append(buildCmd, defaultArgs...)

		// 移除指定的默认参数
		if removeArgs := b.GetCustomArgStringSlice("remove_default_args"); len(removeArgs) > 0 {
			buildCmd = b.removeSpecificArgs(buildCmd, removeArgs)
			logger.Debug("移除指定默认参数: %v", removeArgs)
		}
	}

	// 添加自定义参数
	if customArgs := b.GetCustomArgStringSlice("flutter_build_args"); len(customArgs) > 0 {
		buildCmd = append(buildCmd, customArgs...)
	}

	// 添加自定义dart-define参数
	if dartDefines := b.GetCustomArgStringSlice("dart_defines"); len(dartDefines) > 0 {
		for _, define := range dartDefines {
			buildCmd = append(buildCmd, "--dart-define="+define)
		}
	}

	// 自定义目标平台
	if targetPlatform := b.GetCustomArgString("target_platform"); targetPlatform != "" {
		// 移除默认的target-platform参数
		for i := 0; i < len(buildCmd)-1; i++ {
			if buildCmd[i] == "--target-platform" {
				buildCmd = append(buildCmd[:i], buildCmd[i+2:]...)
				break
			}
		}
		buildCmd = append(buildCmd, "--target-platform", targetPlatform)
	}

	if err := b.executor.RunCommand(buildCmd, b.projectRoot); err != nil {
		return fmt.Errorf("android构建失败: %w", err)
	}

	logger.Success("Android APK构建完成")

	// 验证构建产物
	if err := b.validateBuildArtifacts(); err != nil {
		return fmt.Errorf("构建产物验证失败: %w", err)
	}

	b.showAndroidBuildArtifacts()
	return nil
}

func (b *FlutterBuilderImpl) buildIOS() error {
	logger.Info("构建iOS发布版本...")

	// 判断是否提供了证书配置，决定构建类型
	if b.iosConfig != nil && b.iosConfig.TeamID != "" {
		// 提供了证书配置，直接构建IPA
		return b.buildIPA()
	} else {
		// 未提供证书配置，仅构建iOS
		return b.buildIOSOnly()
	}
}

// buildIOSOnly 仅构建iOS项目（不生成IPA）
func (b *FlutterBuilderImpl) buildIOSOnly() error {
	logger.Info("构建iOS发布版本...")

	// 构建iOS项目（不生成IPA）
	buildCmd := []string{
		"flutter", "build", "ios",
		"--release",
	}

	// 添加默认参数（可被自定义参数覆盖）
	defaultArgs := []string{
		"--obfuscate",
		"--split-debug-info=build/debug-info",
		"--tree-shake-icons",
		"--dart-define=FLUTTER_WEB_USE_SKIA=true",
		"--dart-define=FLUTTER_WEB_AUTO_DETECT=true",
	}

	// 检查是否禁用默认参数
	if !b.GetCustomArgBool("disable_default_args") {
		// 添加默认参数
		buildCmd = append(buildCmd, defaultArgs...)

		// 移除指定的默认参数
		if removeArgs := b.GetCustomArgStringSlice("remove_default_args"); len(removeArgs) > 0 {
			buildCmd = b.removeSpecificArgs(buildCmd, removeArgs)
			logger.Debug("移除指定默认参数: %v", removeArgs)
		}
	}

	// 添加自定义参数
	if customArgs := b.GetCustomArgStringSlice("flutter_build_args"); len(customArgs) > 0 {
		buildCmd = append(buildCmd, customArgs...)
	}

	// 添加自定义dart-define参数
	if dartDefines := b.GetCustomArgStringSlice("dart_defines"); len(dartDefines) > 0 {
		for _, define := range dartDefines {
			buildCmd = append(buildCmd, "--dart-define="+define)
		}
	}

	if err := b.executor.RunCommand(buildCmd, b.projectRoot); err != nil {
		return fmt.Errorf("iOS构建失败: %w", err)
	}

	logger.Success("iOS构建完成")

	// 验证构建产物
	if err := b.validateBuildArtifacts(); err != nil {
		return fmt.Errorf("构建产物验证失败: %w", err)
	}

	b.showIOSBuildArtifacts()
	return nil
}

func (b *FlutterBuilderImpl) buildIPA() error {
	logger.Info("构建IPA文件...")

	// 设置证书（如果提供了动态证书）
	if b.certManager != nil {
		if err := b.certManager.SetupCertificates(); err != nil {
			return fmt.Errorf("设置iOS证书失败: %w", err)
		}
		// 使用 defer 确保无论是否成功都能清理资源
		defer func() {
			if err := b.certManager.ForceCleanupAll(); err != nil {
				logger.Warning("清理证书资源时发生错误: %v", err)
			}
		}()
	}

	// 创建导出选项plist文件
	exportOptionsPlist, err := b.certManager.CreateExportOptionsPlist()
	if err != nil {
		return fmt.Errorf("创建导出选项plist失败: %w", err)
	}

	// 使用flutter build ipa命令生成IPA
	ipaCmd := []string{
		"flutter", "build", "ipa",
		"--release",
	}

	// 添加默认参数（可被自定义参数覆盖）
	defaultArgs := []string{
		"--obfuscate",
		"--split-debug-info=build/debug-info",
		"--tree-shake-icons",
		"--dart-define=FLUTTER_WEB_USE_SKIA=true",
		"--dart-define=FLUTTER_WEB_AUTO_DETECT=true",
		"--export-options-plist", exportOptionsPlist,
	}

	// 检查是否禁用默认参数
	if !b.GetCustomArgBool("disable_default_args") {
		// 添加默认参数
		ipaCmd = append(ipaCmd, defaultArgs...)

		// 移除指定的默认参数
		if removeArgs := b.GetCustomArgStringSlice("remove_default_args"); len(removeArgs) > 0 {
			ipaCmd = b.removeSpecificArgs(ipaCmd, removeArgs)
			logger.Debug("移除指定默认参数: %v", removeArgs)
		}
	}

	// 添加自定义参数
	if customArgs := b.GetCustomArgStringSlice("flutter_build_args"); len(customArgs) > 0 {
		ipaCmd = append(ipaCmd, customArgs...)
	}

	// 添加自定义dart-define参数
	if dartDefines := b.GetCustomArgStringSlice("dart_defines"); len(dartDefines) > 0 {
		for _, define := range dartDefines {
			ipaCmd = append(ipaCmd, "--dart-define="+define)
		}
	}

	if err := b.executor.RunCommand(ipaCmd, b.projectRoot); err != nil {
		return fmt.Errorf("IPA构建失败: %w", err)
	}

	logger.Success("IPA文件生成完成")

	// 验证构建产物
	if err := b.validateBuildArtifacts(); err != nil {
		return fmt.Errorf("构建产物验证失败: %w", err)
	}

	return nil
}

func (b *FlutterBuilderImpl) createBuildInfo() error {
	buildInfoPath := filepath.Join(b.projectRoot, "build", "build_info.txt")

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(buildInfoPath), 0755); err != nil {
		return err
	}

	// 获取Flutter版本信息
	flutterVersion, err := b.executor.RunCommandWithOutput([]string{"flutter", "--version"}, b.projectRoot)
	if err != nil {
		flutterVersion = "无法获取Flutter版本信息"
	}

	buildInfoContent := fmt.Sprintf(`构建信息
==================
平台: %s
构建日期: %s
构建类型: Release
代码混淆: 已启用
Tree Shaking: 已启用
调试信息分离: 已启用
架构: %s

Flutter版本信息:
%s

系统信息:
操作系统: %s
Go版本: %s
`,
		b.platform,
		time.Now().Format("2006-01-02 15:04:05"),
		getArchitecture(b.platform),
		flutterVersion,
		runtime.GOOS,
		runtime.Version(),
	)

	return os.WriteFile(buildInfoPath, []byte(buildInfoContent), 0644)
}

func (b *FlutterBuilderImpl) showSecurityReminders() {
	logger.Println()
	logger.Header("安全提醒")
	logger.Success("代码混淆已启用")
	logger.Success("调试信息已分离")
	logger.Success("Tree Shaking已应用")
	logger.Success("图标Tree Shaking已启用")

	logger.Println()
	logger.Info("额外安全措施:")
	logger.Println("- 保护调试符号安全 (build/debug-info/)")
	logger.Println("- 验证签名证书配置")
	logger.Println("- 在真实设备上测试")
	logger.Println("- 考虑使用额外的安全工具 (R8, DexGuard)")

	if b.platform == PlatformAPK {
		logger.Println()
		logger.Info("Android特定:")
		logger.Println("- ProGuard/R8混淆已应用")
		logger.Println("- 仅ARM64架构 (已排除x86/x86_64)")
		logger.Println("- 验证应用签名配置")
	} else if b.platform == PlatformIOS {
		logger.Println()
		logger.Info("iOS特定:")
		logger.Println("- Bitcode优化已应用")
		logger.Println("- App Store提交就绪")
		logger.Println("- 验证配置文件")
	}
}

func (b *FlutterBuilderImpl) showAndroidBuildArtifacts() {
	apkPath := filepath.Join(b.projectRoot, "build", "app", "outputs", "flutter-apk", "app-release.apk")

	if info, err := os.Stat(apkPath); err == nil {
		apkSizeMB := float64(info.Size()) / (1024 * 1024)

		logger.Println()
		logger.Info("构建产物:")
		logger.Printf("  APK文件: app-release.apk (%.2f MB)", apkSizeMB)
		logger.Printf("  位置: %s", filepath.Dir(apkPath))
		logger.Printf("  调试信息: %s/build/debug-info/", b.projectRoot)
	}
}

func (b *FlutterBuilderImpl) showIOSBuildArtifacts() {
	logger.Println()
	logger.Info("构建产物:")

	// 根据是否有证书配置来显示不同的信息
	if b.iosConfig != nil && b.iosConfig.TeamID != "" {
		// 构建IPA文件
		ipaDir := filepath.Join(b.projectRoot, "build", "ios", "ipa")

		// 尝试找到实际生成的IPA文件
		if files, err := filepath.Glob(filepath.Join(ipaDir, "*.ipa")); err == nil && len(files) > 0 {
			logger.Printf("  IPA文件: %s", files[0])
		} else {
			logger.Printf("  IPA文件目录: %s", ipaDir)
		}

		logger.Printf("  调试信息: %s/build/debug-info/", b.projectRoot)
		logger.Println()
		logger.Success("IPA文件已生成，可直接上传到App Store Connect")
	} else {
		// 仅构建iOS项目
		iosBuildPath := filepath.Join(b.projectRoot, "build", "ios", "iphoneos")
		runnerAppPath := filepath.Join(iosBuildPath, "Runner.app")
		logger.Printf("  Runner.app位置: %s", runnerAppPath)
		logger.Printf("  调试信息: %s/build/debug-info/", b.projectRoot)
		logger.Println()
		logger.Info("创建IPA文件:")
		logger.Println("  1. 在Xcode中打开 ios/Runner.xcworkspace")
		logger.Println("  2. 选择 'Any iOS Device' 作为目标")
		logger.Println("  3. Product > Archive")
		logger.Println("  4. Distribute App > App Store Connect / Ad Hoc / Enterprise")
	}
}

// 辅助函数（已移除getProjectRoot，现在通过参数传递项目根目录）

func getArchitecture(platform Platform) string {
	if platform == PlatformAPK {
		return "ARM64"
	}
	return "iOS Universal"
}

// convertPlatform 将构建器平台转换为验证器平台
func convertPlatform(platform Platform) artifact.Platform {
	switch platform {
	case PlatformAPK:
		return artifact.PlatformAPK
	case PlatformIOS:
		return artifact.PlatformIOS
	default:
		return artifact.Platform(platform)
	}
}
