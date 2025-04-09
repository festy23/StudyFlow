package service

import "context"

type UserClient interface {
	UserExists(ctx context.Context, userID string) bool
	IsPair(ctx context.Context, tutorID, studentID string) bool
	GetUserRole(ctx context.Context, userID string) (string, error)
}

type FileClient interface {
	FileExists(ctx context.Context, fileID string) bool
	GetFileURL(ctx context.Context, fileID string, userID string) (string, error)
}
