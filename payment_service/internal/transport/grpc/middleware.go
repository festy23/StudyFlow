package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	userRoleKey contextKey = "user_role"
)

func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	userID := md.Get("user_id")
	if len(userID) == 0 {
		return nil, status.Error(codes.Unauthenticated, "user_id is not provided")
	}

	userRole := md.Get("user_role")
	if len(userRole) == 0 {
		return nil, status.Error(codes.Unauthenticated, "user_role is not provided")
	}

	ctx = context.WithValue(ctx, userIDKey, userID[0])
	ctx = context.WithValue(ctx, userRoleKey, userRole[0])

	return handler(ctx, req)
}

func getUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func getUserRoleFromContext(ctx context.Context) (string, bool) {
	userRole, ok := ctx.Value(userRoleKey).(string)
	return userRole, ok
}
