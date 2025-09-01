package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mimicode/flutterbuilder/pkg/logger"
)

// DefaultTimeout 默认超时时间
const DefaultTimeout = 30 * time.Second

// HookExecutorImpl 钩子执行器实现
type HookExecutorImpl struct {
	registry    *HookRegistry
	projectRoot string
}

// NewHookExecutor 创建新的钩子执行器
func NewHookExecutor(projectRoot string) HookExecutor {
	return &HookExecutorImpl{
		registry: &HookRegistry{
			hooks: make(map[HookType][]*HookConfig),
		},
		projectRoot: projectRoot,
	}
}

// ExecuteHooks 执行指定类型的所有钩子
func (h *HookExecutorImpl) ExecuteHooks(hookType HookType, context *HookContext) ([]*HookResult, error) {
	hooks := h.registry.hooks[hookType]
	if len(hooks) == 0 {
		return nil, nil // 没有钩子需要执行
	}

	logger.Info("执行 %s 钩子 (%d个)...", hookType, len(hooks))

	var results []*HookResult
	for i, hook := range hooks {
		logger.Info("  [%d/%d] 执行钩子: %s", i+1, len(hooks), hook.ScriptPath)

		result := h.executeHook(hook, context)
		results = append(results, result)

		if !result.Success && !hook.ContinueOnError {
			logger.Error("钩子执行失败，终止构建流程: %s", hook.ScriptPath)
			return results, fmt.Errorf("钩子执行失败: %s", hook.ScriptPath)
		}

		if !result.Success {
			logger.Warning("钩子执行失败但继续构建: %s (错误: %v)", hook.ScriptPath, result.Error)
		} else {
			logger.Success("钩子执行成功: %s (耗时: %.2fs)", hook.ScriptPath, result.Duration.Seconds())
		}
	}

	return results, nil
}

// executeHook 执行单个钩子
func (h *HookExecutorImpl) executeHook(hook *HookConfig, context *HookContext) *HookResult {
	startTime := time.Now()
	result := &HookResult{
		Success: false,
	}

	// 解析脚本路径
	scriptPath := filepath.Join(h.projectRoot, hook.ScriptPath)
	if !filepath.IsAbs(scriptPath) {
		scriptPath = filepath.Clean(scriptPath)
	}

	// 检查脚本文件是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		result.Error = fmt.Errorf("钩子脚本不存在: %s", scriptPath)
		result.Duration = time.Since(startTime)
		return result
	}

	// 准备命令
	var cmd *exec.Cmd
	args := []string{"run", scriptPath}
	args = append(args, hook.Args...)
	cmd = exec.Command("dart", args...)

	// 设置工作目录
	workingDir := h.projectRoot
	if hook.WorkingDir != "" {
		if filepath.IsAbs(hook.WorkingDir) {
			workingDir = hook.WorkingDir
		} else {
			workingDir = filepath.Join(h.projectRoot, hook.WorkingDir)
		}
	}
	cmd.Dir = workingDir

	// 设置环境变量
	cmd.Env = os.Environ()

	// 添加构建上下文环境变量
	cmd.Env = append(cmd.Env, fmt.Sprintf("FLUTTER_BUILDER_HOOK_TYPE=%s", context.HookType))
	cmd.Env = append(cmd.Env, fmt.Sprintf("FLUTTER_BUILDER_PLATFORM=%s", context.Platform))
	cmd.Env = append(cmd.Env, fmt.Sprintf("FLUTTER_BUILDER_PROJECT_ROOT=%s", context.ProjectRoot))
	cmd.Env = append(cmd.Env, fmt.Sprintf("FLUTTER_BUILDER_BUILD_STAGE=%s", context.BuildStage))

	// 添加自定义环境变量
	if hook.Environment != nil {
		for key, value := range hook.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// 设置超时时间
	timeout := hook.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	// 执行命令
	output, err := h.runCommandWithTimeout(cmd, timeout)
	result.Duration = time.Since(startTime)
	result.Output = output

	if err != nil {
		result.Error = err
		// 尝试获取退出码
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
	} else {
		result.Success = true
		result.ExitCode = 0
	}

	return result
}

// runCommandWithTimeout 带超时的命令执行
func (h *HookExecutorImpl) runCommandWithTimeout(cmd *exec.Cmd, timeout time.Duration) (string, error) {
	// 创建超时上下文
	done := make(chan error, 1)
	var output strings.Builder

	// 设置输出
	cmd.Stdout = &output
	cmd.Stderr = &output

	// 启动命令
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("启动钩子脚本失败: %w", err)
	}

	// 异步等待命令完成
	go func() {
		done <- cmd.Wait()
	}()

	// 等待完成或超时
	select {
	case err := <-done:
		return output.String(), err
	case <-time.After(timeout):
		// 超时，杀死进程
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return output.String(), fmt.Errorf("钩子脚本执行超时 (%v)", timeout)
	}
}

// RegisterHook 注册钩子
func (h *HookExecutorImpl) RegisterHook(hookType HookType, config *HookConfig) error {
	if config == nil {
		return fmt.Errorf("钩子配置不能为空")
	}

	if config.ScriptPath == "" {
		return fmt.Errorf("钩子脚本路径不能为空")
	}

	// 检查脚本文件是否存在
	scriptPath := filepath.Join(h.projectRoot, config.ScriptPath)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("钩子脚本不存在: %s", scriptPath)
	}

	if h.registry.hooks[hookType] == nil {
		h.registry.hooks[hookType] = make([]*HookConfig, 0)
	}

	h.registry.hooks[hookType] = append(h.registry.hooks[hookType], config)
	return nil
}

// UnregisterHook 注销钩子
func (h *HookExecutorImpl) UnregisterHook(hookType HookType, scriptPath string) error {
	hooks := h.registry.hooks[hookType]
	if hooks == nil {
		return nil
	}

	for i, hook := range hooks {
		if hook.ScriptPath == scriptPath {
			// 删除指定的钩子
			h.registry.hooks[hookType] = append(hooks[:i], hooks[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("未找到指定的钩子: %s", scriptPath)
}

// GetHooks 获取指定类型的所有钩子
func (h *HookExecutorImpl) GetHooks(hookType HookType) []*HookConfig {
	hooks := h.registry.hooks[hookType]
	if hooks == nil {
		return make([]*HookConfig, 0)
	}

	// 返回副本以避免外部修改
	result := make([]*HookConfig, len(hooks))
	copy(result, hooks)
	return result
}

// SetHooks 设置指定类型的钩子列表
func (h *HookExecutorImpl) SetHooks(hookType HookType, configs []*HookConfig) {
	if configs == nil {
		h.registry.hooks[hookType] = make([]*HookConfig, 0)
	} else {
		// 创建副本以避免外部修改
		h.registry.hooks[hookType] = make([]*HookConfig, len(configs))
		copy(h.registry.hooks[hookType], configs)
	}
}

// ClearHooks 清空指定类型的钩子
func (h *HookExecutorImpl) ClearHooks(hookType HookType) {
	h.registry.hooks[hookType] = make([]*HookConfig, 0)
}

// ClearAllHooks 清空所有钩子
func (h *HookExecutorImpl) ClearAllHooks() {
	h.registry.hooks = make(map[HookType][]*HookConfig)
}
