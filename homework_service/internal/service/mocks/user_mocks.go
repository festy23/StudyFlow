package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

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

type MockFileClient struct {
	mock.Mock
}

func (m *MockFileClient) FileExists(ctx context.Context, fileID string) bool {
	args := m.Called(ctx, fileID)
	return args.Bool(0)
}

func (m *MockFileClient) GetFile(ctx context.Context, fileID string) (interface{}, error) {
	args := m.Called(ctx, fileID)
	return args.Get(0), args.Error(1)
}
