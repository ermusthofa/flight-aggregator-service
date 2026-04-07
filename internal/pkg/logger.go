package pkg

import (
	"context"
	"log"
	"os"
)

var stdLogger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

// Info logs with request ID if present in context
func Info(ctx context.Context, format string, v ...interface{}) {
	id := GetRequestID(ctx)
	if id != "" {
		stdLogger.Printf("[INFO] [req=%s] "+format, append([]interface{}{id}, v...)...)
	} else {
		stdLogger.Printf("[INFO] "+format, v...)
	}
}

// Error logs with request ID
func Error(ctx context.Context, format string, v ...interface{}) {
	id := GetRequestID(ctx)
	if id != "" {
		stdLogger.Printf("[ERROR] [req=%s] "+format, append([]interface{}{id}, v...)...)
	} else {
		stdLogger.Printf("[ERROR] "+format, v...)
	}
}

// Warn logs with request ID
func Warn(ctx context.Context, format string, v ...interface{}) {
	id := GetRequestID(ctx)
	if id != "" {
		stdLogger.Printf("[WARN] [req=%s] "+format, append([]interface{}{id}, v...)...)
	} else {
		stdLogger.Printf("[WARN] "+format, v...)
	}
}
