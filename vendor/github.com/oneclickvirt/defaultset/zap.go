package defaultset

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func getZapConfig() zap.Config {
	cfg := zap.Config{
		Encoding:         "console",                           // 日志编码格式
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel), // 日志级别
		OutputPaths:      []string{"ecs.log"},                 // 输出路径，可以是多个文件
		ErrorOutputPaths: []string{},                          // 错误输出路径，可以是多个文件
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "timestamp",                   // 时间字段名称
			LevelKey:      "level",                       // 日志级别字段名称
			NameKey:       "logger",                      // 日志记录器名称字段名称
			CallerKey:     "caller",                      // 调用者信息字段名称
			MessageKey:    "message",                     // 日志消息字段名称
			StacktraceKey: "stacktrace",                  // 堆栈跟踪信息字段名称
			EncodeLevel:   zapcore.LowercaseLevelEncoder, // 小写格式的日志级别编码器
			EncodeTime:    zapcore.ISO8601TimeEncoder,    // ISO8601 时间格式编码器
			EncodeCaller:  zapcore.ShortCallerEncoder,    // 编码短调用者信息
		},
	}
	return cfg
}

// InitLogger 初始化日志记录器
func InitLogger() {
	// 配置日志记录器
	cfg := getZapConfig()
	var err error
	Logger, err = cfg.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}
