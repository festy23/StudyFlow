package mocks

import (
	"context"
	"homework_service/internal/domain"
	"time"

	"github.com/stretchr/testify/mock"
)

type AssignmentRepository struct {
	mock.Mock
}

func (m *AssignmentRepository) Create(ctx context.Context, assignment *domain.Assignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *AssignmentRepository) GetByID(ctx context.Context, id string) (*domain.Assignment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Assignment), args.Error(1)
}

func (m *AssignmentRepository) FindAssignmentsDueSoon(ctx context.Context, duration time.Duration) ([]*domain.Assignment, error) {
	args := m.Called(ctx, duration)
	return args.Get(0).([]*domain.Assignment), args.Error(1)
}
