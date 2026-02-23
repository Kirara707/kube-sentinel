package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.SugaredLogger for dependency injection
type Logger struct {
	*zap.SugaredLogger
}

// Log is the global logger instance for backward compatibility
// Deprecated: Use dependency injection instead
var Log *zap.SugaredLogger

// New creates a new Logger instance with the specified log level
// This is the recommended way to create a logger for dependency injection
func New(level string) (*Logger, error) {
	// 解析日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 编码器配置：控制台友好的彩色输出
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 彩色日志级别
		EncodeTime:     zapcore.ISO8601TimeEncoder,       // 2006-01-02T15:04:05
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
	sugared := logger.Sugar()

	// Also set the global logger for backward compatibility
	Log = sugared

	return &Logger{SugaredLogger: sugared}, nil
}

// Init initializes the global logger instance
// Deprecated: Use logger.New() for dependency injection instead
func Init(level string) {
	_, _ = New(level)
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
