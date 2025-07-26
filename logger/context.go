package logger

import (
	"context"
	"log/slog"
)

type attrsCtxKey struct{}

func AddAttrsToCtx(ctx context.Context, key string, value any) context.Context {
	attrs := append(AttrsFromCtx(ctx), slog.Any(key, value))
	return context.WithValue(ctx, attrsCtxKey{}, attrs)
}

func AttrsFromCtx(ctx context.Context) []slog.Attr {
	attrs, ok := ctx.Value(attrsCtxKey{}).([]slog.Attr)
	if !ok {
		return []slog.Attr{}
	}
	return attrs
}
