package middleware

import (
	"common_library/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	userpb "userservice/pkg/api"
)

func NewAuthMiddleware(userClient userpb.UserServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			header := r.Header.Get("Authorization")
			if header == "" {
				if logger, ok := logging.GetFromContext(ctx); ok {
					logger.Info(ctx, "no authorization header", zap.Any("request", r))
				}
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			req := &userpb.AuthorizeByAuthHeaderRequest{AuthorizationHeader: header}
			resp, err := userClient.AuthorizeByAuthHeader(ctx, req)
			if err != nil {
				if status.Code(err) == codes.PermissionDenied {
					if logger, ok := logging.GetFromContext(ctx); ok {
						logger.Info(ctx, "permission denied", zap.Any("request", r))
					}
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if logger, ok := logging.GetFromContext(ctx); ok {
					logger.Error(
						ctx, "error while sending grpc auth request",
						zap.Any("request", r),
						zap.Error(err),
						zap.Any("grpc request", req),
						zap.Any("grpc response", resp),
					)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Header.Set("X-User-Id", resp.Id)
			r.Header.Set("X-User-Role", resp.Role)
			next.ServeHTTP(w, r)
		})
	}
}
