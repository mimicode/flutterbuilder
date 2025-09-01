package hooks

import "time"

// HookType 钩子类型
type HookType string

const (
	// 构建流程各阶段的前置钩子
	HookPreClean         HookType = "pre_clean"          // 清理前
	HookPreGetDeps       HookType = "pre_get_deps"       // 获取依赖前
	HookPreCodeGen       HookType = "pre_code_gen"       // 代码生成前
	HookPreSecurityCheck HookType = "pre_security_check" // 安全检查前
	HookPreBuild         HookType = "pre_build"          // 构建前
	HookPrePostProcess   HookType = "pre_post_process"   // 后处理前

	// 构建流程各阶段的后置钩子
	HookPostClean         HookType = "post_clean"          // 清理后
	HookPostGetDeps       HookType = "post_get_deps"       // 获取依赖后
	HookPostCodeGen       HookType = "post_code_gen"       // 代码生成后
	HookPostSecurityCheck HookType = "post_security_check" // 安全检查后
	HookPostBuild         HookType = "post_build"          // 构建后
	HookPostPostProcess   HookType = "post_post_process"   // 后处理后
)

// HookConfig 钩子配置
type HookConfig struct {
	// ScriptPath 脚本文件路径（相对于项目根目录）
	ScriptPath string `json:"script_path"`

	// Args 传递给脚本的参数
	Args []string `json:"args,omitempty"`

	// Timeout 脚本执行超时时间，默认30秒
	Timeout time.Duration `json:"timeout,omitempty"`

	// ContinueOnError 脚本执行失败时是否继续构建流程，默认为false
	ContinueOnError bool `json:"continue_on_error,omitempty"`

	// WorkingDir 脚本工作目录，默认为项目根目录
	WorkingDir string `json:"working_dir,omitempty"`

	// Environment 环境变量
	Environment map[string]string `json:"environment,omitempty"`
}

// HookRegistry 钩子注册表
type HookRegistry struct {
	hooks map[HookType][]*HookConfig
}

// HookContext 钩子执行上下文
type HookContext struct {
	// HookType 当前钩子类型
	HookType HookType

	// Platform 构建平台
	Platform string

	// ProjectRoot 项目根目录
	ProjectRoot string

	// BuildStage 当前构建阶段名称
	BuildStage string

	// StartTime 钩子开始执行时间
	StartTime time.Time

	// CustomArgs 自定义参数
	CustomArgs map[string]interface{}
}

// HookResult 钩子执行结果
type HookResult struct {
	// Success 是否执行成功
	Success bool

	// Error 错误信息
	Error error

	// Duration 执行耗时
	Duration time.Duration

	// Output 脚本输出
	Output string

	// ExitCode 退出码
	ExitCode int
}

// HookExecutor 钩子执行器接口
type HookExecutor interface {
	// ExecuteHooks 执行指定类型的所有钩子
	ExecuteHooks(hookType HookType, context *HookContext) ([]*HookResult, error)

	// RegisterHook 注册钩子
	RegisterHook(hookType HookType, config *HookConfig) error

	// UnregisterHook 注销钩子
	UnregisterHook(hookType HookType, scriptPath string) error

	// GetHooks 获取指定类型的所有钩子
	GetHooks(hookType HookType) []*HookConfig

	// SetHooks 设置指定类型的钩子列表
	SetHooks(hookType HookType, configs []*HookConfig)

	// ClearHooks 清空指定类型的钩子
	ClearHooks(hookType HookType)

	// ClearAllHooks 清空所有钩子
	ClearAllHooks()
}

// HooksConfig 钩子配置集合
type HooksConfig struct {
	Hooks map[HookType][]*HookConfig `json:"hooks"`
}
