package mocks

import (
	"context"
	"homework_service/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockFeedbackRepository struct {
	mock.Mock
}

func (m *MockFeedbackRepository) Create(ctx context.Context, feedback *domain.Feedback) error {
	args := m.Called(ctx, feedback)
	return args.Error(0)
}

func (m *MockFeedbackRepository) GetByID(ctx context.Context, id string) (*domain.Feedback, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Feedback), args.Error(1)
}

func (m *MockFeedbackRepository) Update(ctx context.Context, feedback *domain.Feedback) error {
	args := m.Called(ctx, feedback)
	return args.Error(0)
}

func (m *MockFeedbackRepository) ListByFilter(ctx context.Context, filter domain.FeedbackFilter) ([]*domain.Feedback, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.Feedback), args.Error(1)
}

type FeedbackService struct {
	mock.Mock
}

func (m *FeedbackService) CreateFeedback(ctx context.Context, feedback *domain.Feedback) (*domain.Feedback, error) {
	args := m.Called(ctx, feedback)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Feedback), args.Error(1)
}

func (m *FeedbackService) UpdateFeedback(ctx context.Context, feedback *domain.Feedback) (*domain.Feedback, error) {
	args := m.Called(ctx, feedback)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Feedback), args.Error(1)
}

func (m *FeedbackService) ListFeedbacks(ctx context.Context, filter domain.FeedbackFilter) ([]*domain.Feedback, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Feedback), args.Error(1)
}

func (m *FeedbackService) GetFeedback(ctx context.Context, id string) (*domain.Feedback, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Feedback), args.Error(1)
}
