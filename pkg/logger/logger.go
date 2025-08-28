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

// ExternalLogger 外部日志接口
type ExternalLogger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
	Success(format string, args ...interface{})
	Header(title string)
	Println(args ...interface{})
	Printf(format string, args ...interface{})
}

var (
	currentLevel   = InfoLevel
	debugLogger    *log.Logger
	infoLogger     *log.Logger
	warningLogger  *log.Logger
	errorLogger    *log.Logger
	externalLogger ExternalLogger // 外部日志接口
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

// SetExternalLogger 设置外部日志接口
func SetExternalLogger(logger ExternalLogger) {
	externalLogger = logger
}

// ClearExternalLogger 清除外部日志接口
func ClearExternalLogger() {
	externalLogger = nil
}

// Debug 调试日志
func Debug(format string, args ...interface{}) {
	if externalLogger != nil {
		externalLogger.Debug(format, args...)
		return
	}

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
	if externalLogger != nil {
		externalLogger.Info(format, args...)
		return
	}

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
	if externalLogger != nil {
		externalLogger.Warning(format, args...)
		return
	}

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
	if externalLogger != nil {
		externalLogger.Error(format, args...)
		return
	}

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
	if externalLogger != nil {
		externalLogger.Success(format, args...)
		return
	}

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
	if externalLogger != nil {
		externalLogger.Header(title)
		return
	}

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
	if externalLogger != nil {
		externalLogger.Println(args...)
		return
	}
	fmt.Println(args...)
}

// Printf 格式化输出
func Printf(format string, args ...interface{}) {
	if externalLogger != nil {
		externalLogger.Printf(format, args...)
		return
	}
	fmt.Printf(format, args...)
}

// Fprintf 格式化输出到指定writer
func Fprintf(w *os.File, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
}
