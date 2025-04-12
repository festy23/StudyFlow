package mocks

import (
	"context"
	"homework_service/internal/domain"
	"reflect"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
)

type MockAssignmentService struct {
	ctrl     *gomock.Controller
	recorder *MockAssignmentServiceMockRecorder
}

type MockAssignmentServiceMockRecorder struct {
	mock *MockAssignmentService
}

type AssignmentService struct {
	mock.Mock
}

type MockAssignmentRepository struct {
	mock.Mock
}

type AssignmentServiceMock struct {
	mock.Mock
}

func NewMockAssignmentService(ctrl *gomock.Controller) *MockAssignmentService {
	mock := &MockAssignmentService{ctrl: ctrl}
	mock.recorder = &MockAssignmentServiceMockRecorder{mock}
	return mock
}

func (m *MockAssignmentService) EXPECT() *MockAssignmentServiceMockRecorder {
	return m.recorder
}

func (m *MockAssignmentService) CreateAssignment(ctx context.Context, assignment *domain.Assignment) (*domain.Assignment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAssignment", ctx, assignment)
	ret0, _ := ret[0].(*domain.Assignment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockAssignmentServiceMockRecorder) CreateAssignment(ctx, assignment interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAssignment", reflect.TypeOf((*MockAssignmentService)(nil).CreateAssignment), ctx, assignment)
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
