package clients

import (
	"common_library/logging"
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func New(ctx context.Context, url string) (*grpc.ClientConn, func()) {
	client, err := grpc.NewClient(
		url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		if logger, ok := logging.GetFromContext(ctx); ok {
			logger.Fatal(ctx, "cannot create user grpc client", zap.Error(err))
		}
	}

	closeFunc := func() {
		err := client.Close()
		if err != nil {
			if logger, ok := logging.GetFromContext(ctx); ok {
				logger.Fatal(ctx, "cannot close user grpc client", zap.Error(err))
			}
		}
	}

	return client, closeFunc
}
