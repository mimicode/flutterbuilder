package main

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/mimicode/flutterbuilder/api"
	"github.com/mimicode/flutterbuilder/pkg/hooks"
)

func TestHooks(t *testing.T) {
	fmt.Println("Flutter Builder 钩子功能示例")
	fmt.Println("==============================")

	// 创建钩子配置
	hooksConfig := &hooks.HooksConfig{
		Hooks: map[hooks.HookType][]*hooks.HookConfig{
			// 在清理前执行示例钩子
			hooks.HookPreClean: {
				{
					ScriptPath:      "example/hooks/example_hook.dart",
					Args:            []string{"pre-clean-arg"},
					Timeout:         30 * time.Second,
					ContinueOnError: true,
				},
			},
			// 在构建后记录日志
			hooks.HookPostBuild: {
				{
					ScriptPath:      "example/hooks/log_hook.dart",
					Args:            []string{},
					Timeout:         10 * time.Second,
					ContinueOnError: true,
				},
			},
			// 在后处理完成后执行清理
			hooks.HookPostPostProcess: {
				{
					ScriptPath:      "example/hooks/example_hook.dart",
					Args:            []string{"cleanup"},
					Timeout:         15 * time.Second,
					ContinueOnError: true,
				},
			},
		},
	}

	// 创建构建配置
	config := &api.BuildConfig{
		Platform:    api.PlatformAPK, // 示例使用APK构建
		SourcePath:  ".",             // 使用当前目录作为Flutter项目路径（实际使用时请指定真实的Flutter项目路径）
		HooksConfig: hooksConfig,
		Verbose:     true,
	}

	fmt.Println("开始执行带钩子的构建流程...")
	fmt.Printf("平台: %s\n", config.Platform)
	fmt.Printf("源路径: %s\n", config.SourcePath)
	fmt.Printf("钩子数量: %d 类型\n", len(hooksConfig.Hooks))

	// 创建构建器
	builder := api.NewFlutterBuilder()

	// 执行构建
	result, err := builder.Build(config)
	if err != nil {
		log.Fatalf("构建失败: %v", err)
	}

	// 显示构建结果
	fmt.Println("\n构建结果:")
	fmt.Printf("成功: %v\n", result.Success)
	fmt.Printf("平台: %s\n", result.Platform)
	fmt.Printf("构建时间: %v\n", result.BuildTime)
	fmt.Printf("输出路径: %s\n", result.OutputPath)

	if result.Error != nil {
		fmt.Printf("错误: %v\n", result.Error)
	}

	fmt.Println("\n钩子功能演示完成！")
}
