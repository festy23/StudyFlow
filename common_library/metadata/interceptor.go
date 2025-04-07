package metadata

import (
	"common_library/ctxdata"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewMetadataUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-trace-id"); len(values) > 0 {
				ctx = ctxdata.WithTraceID(ctx, values[0])
			}
			if values := md.Get("x-user-id"); len(values) > 0 {
				ctx = ctxdata.WithUserID(ctx, values[0])
			}
			if values := md.Get("x-user-role"); len(values) > 0 {
				ctx = ctxdata.WithUserRole(ctx, values[0])
			}
		}

		return handler(ctx, req)
	}
}
