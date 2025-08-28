package logger

import (
	"testing"
)

func TestLogLevels(t *testing.T) {
	// 测试不同日志级别
	SetLevel(DebugLevel)

	// 这些应该都能输出
	Debug("调试信息")
	Info("信息")
	Warning("警告")
	Error("错误")
	Success("成功")

	// 设置为Info级别
	SetLevel(InfoLevel)

	// Debug不应该输出
	Debug("这条调试信息不应该显示")

	// 其他应该输出
	Info("信息应该显示")
	Warning("警告应该显示")
	Error("错误应该显示")
	Success("成功应该显示")
}

func TestHeader(t *testing.T) {
	// 测试标题输出
	Header("测试标题")
}

func TestPrintFunctions(t *testing.T) {
	// 测试基本打印函数
	Println("测试Println")
	Printf("测试Printf: %s", "参数")
}
