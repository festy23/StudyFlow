package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockFileClient struct {
	mock.Mock
}

func (m *MockFileClient) FileExists(ctx context.Context, fileID string) bool {
	args := m.Called(ctx, fileID)
	return args.Bool(0)
}

func (m *MockFileClient) GetFileURL(ctx context.Context, fileID string, userID string) (string, error) {
	args := m.Called(ctx, fileID, userID)
	return args.String(0), args.Error(1)
}

func (m *MockFileClient) GetFile(ctx context.Context, id string) (*File, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*File), args.Error(1)
}

type File struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
	OwnerID   string    `json:"owner_id"`
}

type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) UserExists(ctx context.Context, userID string) bool {
	args := m.Called(ctx, userID)
	return args.Bool(0)
}

func (m *MockUserClient) IsPair(ctx context.Context, tutorID, studentID string) bool {
	args := m.Called(ctx, tutorID, studentID)
	return args.Bool(0)
}

func (m *MockUserClient) GetUserRole(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}
