//go:generate mockgen -source=fileServiceClient.go -destination=../mocks/fileService_mock.go -package=mocks

package clients

import (
	"context"
	api2 "fileservice/pkg/api"
	"google.golang.org/grpc"
)

type FileServiceClient interface {
	GenerateDownloadURL(ctx context.Context, req *api2.GenerateDownloadURLRequest, opts ...grpc.CallOption) (*api2.DownloadURL, error)
}
