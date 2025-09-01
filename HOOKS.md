# Flutter Builder 钩子系统

Flutter Builder 现在支持钩子系统，允许在构建流程的每个阶段前后执行自定义的 Dart 脚本。

## 钩子类型

钩子系统支持在以下构建阶段的前后执行脚本：

1. **Clean** (清理)
   - `pre_clean`: 清理前执行
   - `post_clean`: 清理后执行

2. **GetDependencies** (获取依赖)
   - `pre_get_deps`: 获取依赖前执行
   - `post_get_deps`: 获取依赖后执行

3. **RunCodeGeneration** (代码生成)
   - `pre_code_gen`: 代码生成前执行
   - `post_code_gen`: 代码生成后执行

4. **CheckSecurityConfig** (安全检查)
   - `pre_security_check`: 安全检查前执行
   - `post_security_check`: 安全检查后执行

5. **Build** (构建)
   - `pre_build`: 构建前执行
   - `post_build`: 构建后执行

6. **PostBuildProcessing** (后处理)
   - `pre_post_process`: 后处理前执行
   - `post_post_process`: 后处理后执行

## 钩子配置

### 基本配置结构

```go
import "github.com/mimicode/flutterbuilder/pkg/hooks"

hooksConfig := &hooks.HooksConfig{
    Hooks: map[hooks.HookType][]*hooks.HookConfig{
        hooks.HookPreClean: {
            {
                ScriptPath:      "scripts/pre_clean.dart",
                Args:           []string{"arg1", "arg2"},
                Timeout:        30 * time.Second,
                ContinueOnError: false,
                WorkingDir:     "", // 默认为项目根目录
                Environment: map[string]string{
                    "CUSTOM_VAR": "value",
                },
            },
        },
    },
}
```

### 配置选项说明

- `ScriptPath`: 脚本文件路径（相对于项目根目录）
- `Args`: 传递给脚本的参数
- `Timeout`: 脚本执行超时时间（默认30秒）
- `ContinueOnError`: 脚本执行失败时是否继续构建流程（默认false）
- `WorkingDir`: 脚本工作目录（默认为项目根目录）
- `Environment`: 自定义环境变量

## 脚本上下文

钩子脚本执行时会自动设置以下环境变量：

- `FLUTTER_BUILDER_HOOK_TYPE`: 钩子类型
- `FLUTTER_BUILDER_PLATFORM`: 构建平台 (apk/ios)
- `FLUTTER_BUILDER_PROJECT_ROOT`: 项目根目录
- `FLUTTER_BUILDER_BUILD_STAGE`: 构建阶段名称

## 使用示例

### 1. 通过API使用

```go
package main

import (
    "github.com/mimicode/flutterbuilder/api"
    "github.com/mimicode/flutterbuilder/pkg/hooks"
)

func main() {
    // 创建钩子配置
    hooksConfig := &hooks.HooksConfig{
        Hooks: map[hooks.HookType][]*hooks.HookConfig{
            hooks.HookPreBuild: {
                {
                    ScriptPath: "scripts/notify_start.dart",
                    Args:       []string{"build-starting"},
                },
            },
            hooks.HookPostBuild: {
                {
                    ScriptPath: "scripts/upload_artifact.dart",
                    Args:       []string{"--upload-to", "s3"},
                },
            },
        },
    }

    // 创建构建配置
    config := &api.BuildConfig{
        Platform:    api.PlatformAPK,
        SourcePath:  "/path/to/flutter/project",
        HooksConfig: hooksConfig,
    }

    // 执行构建
    builder := api.NewFlutterBuilder()
    result, err := builder.Build(config)
    // ... 处理结果
}
```

### 2. 便捷方法

```go
// 带钩子的快速构建
result, err := api.QuickBuildAPKWithHooks("/path/to/project", hooksConfig)

// 带钩子的iOS构建
result, err := api.QuickBuildIOSWithHooks("/path/to/project", iosConfig, hooksConfig)
```

### 3. 动态注册钩子

```go
builder := builder.NewFlutterBuilder("apk", nil, "/path/to/project")

// 注册单个钩子
hookConfig := &hooks.HookConfig{
    ScriptPath: "scripts/custom_hook.dart",
    Timeout:    15 * time.Second,
}
builder.RegisterHook(hooks.HookPreBuild, hookConfig)

// 执行构建
err := builder.Run()
```

## 钩子脚本示例

### 1. 基本钩子脚本

```dart
// scripts/example_hook.dart
import 'dart:io';

void main(List<String> arguments) {
    final hookType = Platform.environment['FLUTTER_BUILDER_HOOK_TYPE'];
    final platform = Platform.environment['FLUTTER_BUILDER_PLATFORM'];
    
    print('执行钩子: $hookType，平台: $platform');
    
    // 执行自定义逻辑
    switch (hookType) {
        case 'pre_build':
            print('准备构建...');
            break;
        case 'post_build':
            print('构建完成，处理产物...');
            break;
    }
}
```

### 2. 日志记录钩子

```dart
// scripts/log_hook.dart
import 'dart:io';

void main(List<String> arguments) {
    final hookType = Platform.environment['FLUTTER_BUILDER_HOOK_TYPE'];
    final timestamp = DateTime.now().toIso8601String();
    
    final logEntry = '[$timestamp] Hook: $hookType\n';
    File('build_hooks.log').writeAsStringSync(logEntry, mode: FileMode.append);
}
```

### 3. 构建通知钩子

```dart
// scripts/notify_hook.dart
import 'dart:io';
import 'dart:convert';

void main(List<String> arguments) async {
    final hookType = Platform.environment['FLUTTER_BUILDER_HOOK_TYPE'];
    
    if (hookType == 'post_build') {
        // 发送构建完成通知
        final notification = {
            'message': 'Flutter build completed',
            'timestamp': DateTime.now().toIso8601String(),
            'platform': Platform.environment['FLUTTER_BUILDER_PLATFORM'],
        };
        
        // 这里可以调用webhook API或发送邮件
        print('构建完成通知: ${jsonEncode(notification)}');
    }
}
```

## 最佳实践

1. **错误处理**: 对于关键钩子，设置 `ContinueOnError: false`；对于可选钩子，设置 `ContinueOnError: true`
2. **超时设置**: 根据脚本复杂度合理设置超时时间
3. **日志记录**: 在钩子脚本中添加适当的日志输出
4. **版本控制**: 将钩子脚本纳入版本控制，与项目代码一起管理
5. **环境隔离**: 使用环境变量区分不同的构建环境（开发/测试/生产）

## 故障排除

- 确保 Dart 运行时已安装并在 PATH 中
- 检查脚本路径是否正确（相对于项目根目录）
- 验证脚本文件权限
- 查看钩子执行日志输出
- 使用 `ContinueOnError: true` 进行调试

## 向后兼容性

钩子系统是完全可选的。不配置钩子时，构建流程与之前完全相同，确保向后兼容性。