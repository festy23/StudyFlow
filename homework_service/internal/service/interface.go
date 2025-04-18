package service

import (
	"context"
	"github.com/google/uuid"
)

type UserClient interface {
	UserExists(ctx context.Context, userID uuid.UUID) bool
	IsPair(ctx context.Context, tutorID, studentID uuid.UUID) bool
	GetUserRole(ctx context.Context, userID uuid.UUID) (string, error)
}

type FileClient interface {
	GetFileURL(ctx context.Context, fileID uuid.UUID) (string, error)
}
