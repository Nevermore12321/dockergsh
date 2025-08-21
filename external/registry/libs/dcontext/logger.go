package dcontext

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
	"sync"
)

var (
	// logrus WithField 可以向日志输出中添加一些字段
	defaultLogger      *logrus.Entry = logrus.StandardLogger().WithField("go.version", runtime.Version())
	defaultLoggerMutex sync.RWMutex
)

type loggerKey struct{}
type Logger interface {
}

// WithLogger 在 ctx 中添加 logger 属性
func WithLogger(ctx context.Context, logger Logger) context.Context {
	ctx = context.WithValue(ctx, loggerKey{}, logger)
	return ctx
}

func GetLogger(ctx context.Context, keys ...interface{}) Logger {
	return getLogrusLogger(ctx, keys...)
}

// 从 ctx 中获取 logrus logger，如果传入了 keys 参数，那么会从 ctx 中对应的 key，添加到 logger 中的 WithField 中
func getLogrusLogger(ctx context.Context, keys ...interface{}) *logrus.Entry {
	var logger *logrus.Entry

	// 从 ctx 中获取 logger
	loggerInterface := ctx.Value(loggerKey{})

	// 如果 context 中有 logger
	if loggerInterface != nil {
		if lgr, ok := loggerInterface.(*logrus.Entry); ok {
			logger = lgr
		}
	}

	// 如果 context 中没有 logger，使用默认的 defailtLogger
	if logger == nil {
		fields := logrus.Fields{}

		// 添加 instance id
		instanceId := ctx.Value("instance.id")
		if instanceId != nil {
			fields["instance.id"] = instanceId
		}

		// 设置到 defaultLogger 中的 fields 中
		defaultLoggerMutex.RLock()
		logger = defaultLogger.WithFields(fields)
		defaultLoggerMutex.Unlock()
	}

	// 自定义的 keys fields 也添加到 logger 中
	fields := logrus.Fields{}
	for _, key := range keys {
		val := ctx.Value(key)
		if val != nil {
			fields[fmt.Sprint(key)] = val
		}
	}

	return logger.WithFields(fields)
}
