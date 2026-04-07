package pkg

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger *zap.SugaredLogger
}

func NewZapLogger() (*ZapLogger, error) {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)
	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	return &ZapLogger{logger: zapLogger.Sugar()}, nil
}

func (l *ZapLogger) Info(ctx context.Context, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	id := GetRequestID(ctx)
	if id != "" {
		l.logger.Infow(msg, "request_id", id)
	} else {
		l.logger.Info(msg)
	}
}

func (l *ZapLogger) Error(ctx context.Context, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	id := GetRequestID(ctx)
	if id != "" {
		l.logger.Errorw(msg, "request_id", id)
	} else {
		l.logger.Error(msg)
	}
}

func (l *ZapLogger) Warn(ctx context.Context, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	id := GetRequestID(ctx)
	if id != "" {
		l.logger.Warnw(msg, "request_id", id)
	} else {
		l.logger.Warn(msg)
	}
}
