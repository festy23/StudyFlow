package handler

import (
	"common_library/logging"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"time"
)

var (
	NilMethodError     = errors.New("grpc method is nil")
	BadRequestError    = errors.New("bad request")
	WrongGrpcTypeError = errors.New("wrong grpc request type")
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, data []byte, ttl time.Duration)
	Delete(ctx context.Context, key string)
}

func mapErr(err error) int {
	if errors.Is(err, BadRequestError) {
		return http.StatusBadRequest
	}
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			return http.StatusBadRequest
		case codes.AlreadyExists:
			return http.StatusConflict
		case codes.PermissionDenied:
			return http.StatusForbidden
		case codes.NotFound:
			return http.StatusNotFound
		}
	}
	return http.StatusInternalServerError
}

func Handle[Req any, Resp any](
	method func(context.Context, *Req, ...grpc.CallOption) (*Resp, error),
	reqParser func(context.Context, *http.Request, *Req) error,
	parseBody bool,
) (http.HandlerFunc, error) {
	if method == nil {
		return nil, NilMethodError
	}

	if _, ok := any(new(Req)).(proto.Message); !ok {
		return nil, WrongGrpcTypeError
	}

	if _, ok := any(new(Resp)).(proto.Message); !ok {
		return nil, WrongGrpcTypeError
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs())
		if id := r.Header.Get("X-User-Id"); id != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-user-id", id)
		}

		if role := r.Header.Get("X-User-Role"); role != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-user-role", role)
		}

		grpcReq := new(Req)

		if parseBody {
			reqJson, err := io.ReadAll(r.Body)
			if err != nil {
				if logger, ok := logging.GetFromContext(r.Context()); ok {
					logger.Error(ctx, "Failed to read request body", zap.Error(err))
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if err := protojson.Unmarshal(reqJson, any(grpcReq).(proto.Message)); err != nil {
				if logger, ok := logging.GetFromContext(r.Context()); ok {
					logger.Error(ctx, "Failed to parse request body", zap.Error(err))
				}
				w.WriteHeader(mapErr(err))
				return
			}
		}

		if reqParser != nil {
			if err := reqParser(ctx, r, grpcReq); err != nil {
				if logger, ok := logging.GetFromContext(r.Context()); ok {
					logger.Error(ctx, "Failed to parse request path and query", zap.Error(err))
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if logger, ok := logging.GetFromContext(r.Context()); ok {
			logger.Debug(ctx, "req after parsing path", zap.Any("req", grpcReq))
		}

		if logger, ok := logging.GetFromContext(r.Context()); ok {
			logger.Debug(ctx, "Sending request", zap.Any("grpcReq", grpcReq))
		}

		grpcResp, err := method(ctx, grpcReq)
		if err != nil {
			if logger, ok := logging.GetFromContext(r.Context()); ok {
				logger.Error(ctx, "grpc request failed", zap.Error(err))
			}
			w.WriteHeader(mapErr(err))
			return
		}

		if logger, ok := logging.GetFromContext(r.Context()); ok {
			logger.Debug(ctx, "Recieved response", zap.Any("grpcResp", grpcResp))
		}

		data, err := protojson.Marshal(any(grpcResp).(proto.Message))
		if err != nil {
			if logger, ok := logging.GetFromContext(r.Context()); ok {
				logger.Error(ctx, "Failed to parse response message", zap.Error(err))
			}
			w.WriteHeader(mapErr(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}, nil
}

func HandleWithCache[Req any, Resp any](
	method func(context.Context, *Req, ...grpc.CallOption) (*Resp, error),
	reqParser func(context.Context, *http.Request, *Req) error,
	parseBody bool,
	cache Cache,
	keyFunc func(r *http.Request) (string, error),
	ttl time.Duration,
) (http.HandlerFunc, error) {
	if method == nil {
		return nil, NilMethodError
	}
	if _, ok := any(new(Req)).(proto.Message); !ok {
		return nil, WrongGrpcTypeError
	}
	if _, ok := any(new(Resp)).(proto.Message); !ok {
		return nil, WrongGrpcTypeError
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs())
		if id := r.Header.Get("X-User-Id"); id != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-user-id", id)
		}
		if role := r.Header.Get("X-User-Role"); role != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-user-role", role)
		}

		key, err := keyFunc(r)
		if err == nil {
			if data, ok := cache.Get(ctx, key); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(data)
				return
			}
		}

		grpcReq := new(Req)
		if parseBody {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := protojson.Unmarshal(body, any(grpcReq).(proto.Message)); err != nil {
				w.WriteHeader(mapErr(err))
				return
			}
		}
		if reqParser != nil {
			if err := reqParser(ctx, r, grpcReq); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		grpcResp, err := method(ctx, grpcReq)
		if err != nil {
			w.WriteHeader(mapErr(err))
			return
		}

		data, err := protojson.Marshal(any(grpcResp).(proto.Message))
		if err != nil {
			w.WriteHeader(mapErr(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)

		if key != "" {
			cache.Set(ctx, key, data, ttl)
		}
	}, nil
}

func parsePathParam(r *http.Request, key string) (string, error) {
	val := chi.URLParam(r, key)
	if val == "" {
		return "", fmt.Errorf("missing path param: %s", key)
	}
	return val, nil
}
