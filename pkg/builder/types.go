package builder

import (
	"time"

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
