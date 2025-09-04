package builder

import (
	"time"

	"github.com/mimicode/flutterbuilder/pkg/artifact"
	"github.com/mimicode/flutterbuilder/pkg/hooks"
	"github.com/mimicode/flutterbuilder/pkg/types"
)

// Platform 构建平台
type Platform string

const (
	PlatformAPK Platform = "apk"
	PlatformIOS Platform = "ios"
)

// IOSConfig iOS构建配置
type IOSConfig = types.IOSConfig

// BuildResult 构建结果
type BuildResult struct {
	Success    bool
	Platform   Platform
	BuildTime  time.Duration
	OutputPath string
	Error      error
}

// FlutterBuilder Flutter构建器接口
type FlutterBuilder interface {
	Run() error
	Clean() error
	GetDependencies() error
	RunCodeGeneration() error
	CheckSecurityConfig() error
	Build() error
	PostBuildProcessing() error
	// 自定义参数方法
	SetCustomArgs(args map[string]interface{}) // 设置自定义构建参数
	// 钩子相关方法
	SetHooks(hooksConfig *hooks.HooksConfig) error                        // 设置钩子配置
	RegisterHook(hookType hooks.HookType, config *hooks.HookConfig) error // 注册单个钩子
	UnregisterHook(hookType hooks.HookType, scriptPath string) error      // 注销钩子
	GetHooks(hookType hooks.HookType) []*hooks.HookConfig                 // 获取指定类型的钩子
	ClearHooks(hookType hooks.HookType)                                   // 清空指定类型的钩子
	ClearAllHooks()                                                       // 清空所有钩子
	// 验证相关方法
	SetValidationConfig(config *artifact.ArtifactValidationConfig)       // 设置验证配置
	GetValidationConfig() *artifact.ArtifactValidationConfig             // 获取验证配置
}

// CommandRunner 命令运行器接口
type CommandRunner interface {
	RunCommand(cmd []string, cwd string) error
	RunCommandWithOutput(cmd []string, cwd string) (string, error)
}

// SecurityChecker 安全配置检查器接口
type SecurityChecker interface {
	CheckAndroidSecurity() error
	CheckIOSSecurity() error
}
