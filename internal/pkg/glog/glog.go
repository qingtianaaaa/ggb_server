package glog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

var Logger *zap.Logger

func InitLogger(logPath string) {
	// 日志编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel)

	// 日志输出配置
	var cores []zapcore.Core

	// 控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atomicLevel)
	cores = append(cores, consoleCore)

	// 文件输出
	if logPath != "" {
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    100, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		})
		fileCore := zapcore.NewCore(fileEncoder, fileWriter, atomicLevel)
		cores = append(cores, fileCore)
	}

	// 创建 Logger
	core := zapcore.NewTee(cores...)
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	defer Logger.Sync()
}

func WithRequestID(requestID string) *zap.Logger {
	return Logger.With(zap.String("requestID", requestID))
}
