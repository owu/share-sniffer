package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var (
	// CurrentLevel 当前日志级别
	CurrentLevel = LevelInfo
	// logger 标准库logger实例
	logger *log.Logger
	// levelNames 日志级别名称映射
	levelNames = map[LogLevel]string{
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
		LevelFatal: "FATAL",
	}
)

func init() {
	// 初始化logger
	logger = log.New(os.Stdout, "", 0)
}

// SetLogLevel 设置日志级别
func SetLogLevel(level LogLevel) {
	CurrentLevel = level
}

// Debug 记录调试日志
func Debug(format string, args ...interface{}) {
	if CurrentLevel <= LevelDebug {
		logMessage(LevelDebug, format, args...)
	}
}

// Info 记录信息日志
func Info(format string, args ...interface{}) {
	if CurrentLevel <= LevelInfo {
		logMessage(LevelInfo, format, args...)
	}
}

// Warn 记录警告日志
func Warn(format string, args ...interface{}) {
	if CurrentLevel <= LevelWarn {
		logMessage(LevelWarn, format, args...)
	}
}

// Error 记录错误日志
func Error(format string, args ...interface{}) {
	if CurrentLevel <= LevelError {
		logMessage(LevelError, format, args...)
	}
}

// Fatal 记录致命错误并退出
func Fatal(format string, args ...interface{}) {
	logMessage(LevelFatal, format, args...)
	os.Exit(1)
}

// logMessage 内部日志记录函数
func logMessage(level LogLevel, format string, args ...interface{}) {
	// 获取调用信息
	pc, file, line, ok := runtime.Caller(3) // 3表示调用链上的第3层
	functionName := "unknown"
	if ok {
		function := runtime.FuncForPC(pc)
		if function != nil {
			functionName = function.Name()
		}
	}

	// 格式化消息
	message := fmt.Sprintf(format, args...)

	// 构造完整日志行
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("[%s] [%s] [%s:%d] [%s] %s\n",
		timestamp,
		levelNames[level],
		file,
		line,
		functionName,
		message,
	)

	// 输出日志
	logger.Print(logLine)
}

// WithFields 记录带字段的日志（简单实现）
func WithFields(fields map[string]interface{}) *Entry {
	return &Entry{
		fields: fields,
	}
}

// Entry 日志条目
type Entry struct {
	fields map[string]interface{}
}

// Debug 记录带字段的调试日志
func (q *Entry) Debug(format string, args ...interface{}) {
	if CurrentLevel <= LevelDebug {
		q.logMessage(LevelDebug, format, args...)
	}
}

// Info 记录带字段的信息日志
func (q *Entry) Info(format string, args ...interface{}) {
	if CurrentLevel <= LevelInfo {
		q.logMessage(LevelInfo, format, args...)
	}
}

// Error 记录带字段的错误日志
func (q *Entry) Error(format string, args ...interface{}) {
	if CurrentLevel <= LevelError {
		q.logMessage(LevelError, format, args...)
	}
}

// logMessage 记录带字段的日志
func (q *Entry) logMessage(level LogLevel, format string, args ...interface{}) {
	// 格式化字段
	fieldsStr := ""
	for k, v := range q.fields {
		fieldsStr += fmt.Sprintf(" %s=%v", k, v)
	}

	// 获取调用信息
	pc, file, line, ok := runtime.Caller(4) // 4表示调用链上的第4层
	functionName := "unknown"
	if ok {
		function := runtime.FuncForPC(pc)
		if function != nil {
			functionName = function.Name()
		}
	}

	// 格式化消息
	message := fmt.Sprintf(format, args...)

	// 构造完整日志行
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("[%s] [%s] [%s:%d] [%s]%s %s\n",
		timestamp,
		levelNames[level],
		file,
		line,
		functionName,
		fieldsStr,
		message,
	)

	// 输出日志
	logger.Print(logLine)
}
