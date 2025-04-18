package service

import (
	"context"
	"github.com/google/uuid"
)

type UserClient interface {
	IsPair(ctx context.Context, tutorID, studentID uuid.UUID) (bool, error)
}

type FileClient interface {
	GetFileURL(ctx context.Context, fileID uuid.UUID) (string, error)
}
