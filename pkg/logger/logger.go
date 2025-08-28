package logger

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

// LogLevel 日志级别
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
)

var (
	currentLevel  = InfoLevel
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
)

// 初始化日志记录器
func init() {
	debugLogger = log.New(os.Stdout, "", 0)
	infoLogger = log.New(os.Stdout, "", 0)
	warningLogger = log.New(os.Stdout, "", 0)
	errorLogger = log.New(os.Stderr, "", 0)
}

// SetLevel 设置日志级别
func SetLevel(level LogLevel) {
	currentLevel = level
}

// Debug 调试日志
func Debug(format string, args ...interface{}) {
	if currentLevel <= DebugLevel {
		message := fmt.Sprintf(format, args...)
		if color.NoColor {
			debugLogger.Printf("[DEBUG] %s", message)
		} else {
			color.Cyan("🔍 %s", message)
		}
	}
}

// Info 信息日志
func Info(format string, args ...interface{}) {
	if currentLevel <= InfoLevel {
		message := fmt.Sprintf(format, args...)
		if color.NoColor {
			infoLogger.Printf("[INFO] %s", message)
		} else {
			color.Blue("ℹ %s", message)
		}
	}
}

// Warning 警告日志
func Warning(format string, args ...interface{}) {
	if currentLevel <= WarningLevel {
		message := fmt.Sprintf(format, args...)
		if color.NoColor {
			warningLogger.Printf("[WARNING] %s", message)
		} else {
			color.Yellow("⚠ %s", message)
		}
	}
}

// Error 错误日志
func Error(format string, args ...interface{}) {
	if currentLevel <= ErrorLevel {
		message := fmt.Sprintf(format, args...)
		if color.NoColor {
			errorLogger.Printf("[ERROR] %s", message)
		} else {
			color.Red("✗ %s", message)
		}
	}
}

// Success 成功日志
func Success(format string, args ...interface{}) {
	if currentLevel <= InfoLevel {
		message := fmt.Sprintf(format, args...)
		if color.NoColor {
			infoLogger.Printf("[SUCCESS] %s", message)
		} else {
			color.Green("✓ %s", message)
		}
	}
}

// Header 标题日志
func Header(title string) {
	separator := strings.Repeat("=", 50)
	if color.NoColor {
		fmt.Println(separator)
		fmt.Println(title)
		fmt.Println(separator)
	} else {
		color.Cyan(separator)
		color.Cyan(title)
		color.Cyan(separator)
	}
}

// Println 普通输出
func Println(args ...interface{}) {
	fmt.Println(args...)
}

// Printf 格式化输出
func Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Fprintf 格式化输出到指定writer
func Fprintf(w *os.File, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
}
