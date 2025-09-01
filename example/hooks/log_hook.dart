// 简单的钩子脚本示例 - 记录构建日志

import 'dart:io';

void main(List<String> arguments) {
  final hookType =
      Platform.environment['FLUTTER_BUILDER_HOOK_TYPE'] ?? 'unknown';
  final timestamp = DateTime.now().toIso8601String();

  // 记录钩子执行日志到文件
  final logFile = File('build_hooks.log');
  final logEntry = '[$timestamp] Hook executed: $hookType\n';

  logFile.writeAsStringSync(logEntry, mode: FileMode.append);

  print('钩子执行记录已写入 build_hooks.log');
}
