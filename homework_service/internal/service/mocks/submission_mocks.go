package mocks

import (
	"context"
	"homework_service/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockSubmissionRepository struct {
	mock.Mock
}

func (m *MockSubmissionRepository) Create(ctx context.Context, submission *domain.Submission) error {
	args := m.Called(ctx, submission)
	return args.Error(0)
}

func (m *MockSubmissionRepository) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Submission), args.Error(1)
}

func (m *MockSubmissionRepository) ListByFilter(ctx context.Context, filter domain.SubmissionFilter) ([]*domain.Submission, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.Submission), args.Error(1)
}

type SubmissionService struct {
	mock.Mock
}

func (m *SubmissionService) CreateSubmission(ctx context.Context, submission *domain.Submission) (*domain.Submission, error) {
	args := m.Called(ctx, submission)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Submission), args.Error(1)
}

func (m *SubmissionService) GetSubmission(ctx context.Context, id string) (*domain.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Submission), args.Error(1)
}

func (m *SubmissionService) ListSubmissions(ctx context.Context, filter domain.SubmissionFilter) ([]*domain.Submission, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Submission), args.Error(1)
}
