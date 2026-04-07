package pkg

import "context"

type contextKey string

const RequestIDKey contextKey = "request_id"

// WithRequestID returns a new context with the given request ID
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey, id)
}

// GetRequestID extracts request ID from context, returns empty string if not found
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
