package mocks

import (
	"context"
	"homework_service/internal/domain"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockAssignmentService struct {
	mock.Mock
}

type MockAssignmentRepository struct {
	mock.Mock
}

func (m *MockAssignmentRepository) Create(ctx context.Context, assignment *domain.Assignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *MockAssignmentRepository) GetByID(ctx context.Context, id string) (*domain.Assignment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Assignment), args.Error(1)
}

func (m *MockAssignmentRepository) FindAssignmentsDueSoon(ctx context.Context, duration time.Duration) ([]*domain.Assignment, error) {
	args := m.Called(ctx, duration)
	return args.Get(0).([]*domain.Assignment), args.Error(1)
}

func (m *MockAssignmentService) CreateAssignment(ctx context.Context, assignment *domain.Assignment) (*domain.Assignment, error) {
	args := m.Called(ctx, assignment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Assignment), args.Error(1)
}

func (m *MockAssignmentService) GetAssignment(ctx context.Context, id string) (*domain.Assignment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Assignment), args.Error(1)
}

func (m *MockAssignmentService) UpdateAssignment(ctx context.Context, assignment *domain.Assignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *MockAssignmentService) ListAssignments(ctx context.Context, filter domain.AssignmentFilter) ([]*domain.Assignment, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Assignment), args.Error(1)
}
