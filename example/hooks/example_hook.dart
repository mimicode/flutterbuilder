// Flutter Builder 钩子示例脚本
// 这个脚本展示了如何在构建流程的每个阶段执行自定义操作

import 'dart:io';

void main(List<String> arguments) {
  // 从环境变量中获取构建上下文信息
  final hookType =
      Platform.environment['FLUTTER_BUILDER_HOOK_TYPE'] ?? 'unknown';
  final platform =
      Platform.environment['FLUTTER_BUILDER_PLATFORM'] ?? 'unknown';
  final projectRoot =
      Platform.environment['FLUTTER_BUILDER_PROJECT_ROOT'] ?? '';
  final buildStage =
      Platform.environment['FLUTTER_BUILDER_BUILD_STAGE'] ?? 'unknown';

  print('[钩子] 执行钩子脚本');
  print('[钩子] 钩子类型: $hookType');
  print('[钩子] 构建平台: $platform');
  print('[钩子] 项目根目录: $projectRoot');
  print('[钩子] 构建阶段: $buildStage');
  print('[钩子] 命令行参数: ${arguments.join(' ')}');

  // 根据钩子类型执行不同的操作
  switch (hookType) {
    case 'pre_clean':
      print('[钩子] 准备清理操作...');
      // 可以在这里备份重要文件或记录状态
      break;
    case 'post_clean':
      print('[钩子] 清理操作完成');
      break;
    case 'pre_get_deps':
      print('[钩子] 准备获取依赖...');
      // 可以在这里检查网络连接或设置代理
      break;
    case 'post_get_deps':
      print('[钩子] 依赖获取完成');
      // 可以在这里验证依赖版本或添加额外的依赖
      break;
    case 'pre_code_gen':
      print('[钩子] 准备代码生成...');
      // 可以在这里准备代码生成的输入文件
      break;
    case 'post_code_gen':
      print('[钩子] 代码生成完成');
      // 可以在这里验证生成的代码或进行后处理
      break;
    case 'pre_security_check':
      print('[钩子] 准备安全检查...');
      break;
    case 'post_security_check':
      print('[钩子] 安全检查完成');
      break;
    case 'pre_build':
      print('[钩子] 准备构建...');
      // 可以在这里设置构建环境变量或验证构建条件
      break;
    case 'post_build':
      print('[钩子] 构建完成');
      // 可以在这里处理构建产物，如签名、上传等
      break;
    case 'pre_post_process':
      print('[钩子] 准备后处理...');
      break;
    case 'post_post_process':
      print('[钩子] 后处理完成');
      // 可以在这里发送通知、生成报告等
      break;
    default:
      print('[钩子] 未知的钩子类型: $hookType');
  }

  // 模拟一些处理时间
  sleep(Duration(milliseconds: 500));

  print('[钩子] 钩子脚本执行完成');

  // 返回成功退出码
  exit(0);
}
