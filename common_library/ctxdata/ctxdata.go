package ctxdata

import (
	"context"
)

type traceIDKey struct{}
type userIDKey struct{}
type userRoleKey struct{}

var (
	traceIDKeyInstance  = traceIDKey{}
	userIDKeyInstance   = userIDKey{}
	userRoleKeyInstance = userRoleKey{}
)

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKeyInstance, traceID)
}

func GetTraceID(ctx context.Context) (string, bool) {
	v := ctx.Value(traceIDKeyInstance)
	traceID, ok := v.(string)
	return traceID, ok
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKeyInstance, userID)
}

func GetUserID(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKeyInstance)
	userID, ok := v.(string)
	return userID, ok
}

func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, userRoleKeyInstance, role)
}

func GetUserRole(ctx context.Context) (string, bool) {
	v := ctx.Value(userRoleKeyInstance)
	role, ok := v.(string)
	return role, ok
}
