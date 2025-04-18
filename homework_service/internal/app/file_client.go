package app

import (
	"common_library/ctxdata"
	"context"
	filePb "fileservice/pkg/api"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type FileClient struct {
	client filePb.FileServiceClient
}

func NewFileClient(conn *grpc.ClientConn) *FileClient {
	return &FileClient{client: filePb.NewFileServiceClient(conn)}
}

func (c *FileClient) GetFileURL(ctx context.Context, fileID uuid.UUID) (string, error) {
	outCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs())
	if id, ok := ctxdata.GetUserID(ctx); ok {
		outCtx = metadata.AppendToOutgoingContext(outCtx, "x-user-id", id)
	}
	if role, ok := ctxdata.GetUserRole(ctx); ok {
		outCtx = metadata.AppendToOutgoingContext(outCtx, "x-user-role", role)
	}
	resp, err := c.client.GenerateDownloadURL(ctx, &filePb.GenerateDownloadURLRequest{FileId: fileID.String()})
	if err != nil {
		return "", err
	}
	return resp.Url, nil
}
