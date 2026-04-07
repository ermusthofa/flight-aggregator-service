package pkg

import "context"

type Logger interface {
	Info(ctx context.Context, format string, v ...interface{})
	Error(ctx context.Context, format string, v ...interface{})
	Warn(ctx context.Context, format string, v ...interface{})
}
