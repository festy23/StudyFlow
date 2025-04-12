package mocks

import (
	"context"
	"homework_service/internal/domain"
	"time"

	"github.com/stretchr/testify/mock"
)

type AssignmentService struct {
	mock.Mock
}

type MockAssignmentRepository struct {
	mock.Mock
}

type AssignmentServiceMock struct {
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

func (m *AssignmentServiceMock) CreateAssignment(ctx context.Context, assignment *domain.Assignment) (*domain.Assignment, error) {
	args := m.Called(ctx, assignment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Assignment), args.Error(1)
}

func (m *AssignmentService) GetAssignment(ctx context.Context, id string) (*domain.Assignment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Assignment), args.Error(1)
}

func (m *AssignmentService) UpdateAssignment(ctx context.Context, assignment *domain.Assignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *AssignmentService) ListAssignments(ctx context.Context, filter domain.AssignmentFilter) ([]*domain.Assignment, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Assignment), args.Error(1)
}
