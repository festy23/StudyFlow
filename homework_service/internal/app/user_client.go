package app

import (
	"common_library/ctxdata"
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	userPb "userservice/pkg/api"
)

type UserClient struct {
	client userPb.UserServiceClient
}

func NewUserClient(conn *grpc.ClientConn) *UserClient {
	return &UserClient{client: userPb.NewUserServiceClient(conn)}
}

func (c *UserClient) IsPair(ctx context.Context, tutorID, studentID uuid.UUID) (bool, error) {
	req := &userPb.GetTutorStudentRequest{
		TutorId:   tutorID.String(),
		StudentId: studentID.String(),
	}
	outCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs())
	if id, ok := ctxdata.GetUserID(ctx); ok {
		outCtx = metadata.AppendToOutgoingContext(outCtx, "x-user-id", id)
	}
	if role, ok := ctxdata.GetUserRole(ctx); ok {
		outCtx = metadata.AppendToOutgoingContext(outCtx, "x-user-role", role)
	}
	resp, err := c.client.GetTutorStudent(outCtx, req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return resp.Status == "active", nil
}
