package mocks

import (
	"context"
	"homework_service/internal/domain"
	"reflect"

	"github.com/golang/mock/gomock"
)

type UserClient interface {
	UserExists(ctx context.Context, userID string) bool
	IsPair(ctx context.Context, tutorID, studentID string) bool
}

type FileClient interface {
	FileExists(ctx context.Context, fileID string) bool
}

type MockAssignmentRepository struct {
	ctrl     *gomock.Controller
	recorder *MockAssignmentRepositoryMockRecorder
}

type MockAssignmentRepositoryMockRecorder struct {
	mock *MockAssignmentRepository
}

func NewMockAssignmentRepository(ctrl *gomock.Controller) *MockAssignmentRepository {
	mock := &MockAssignmentRepository{ctrl: ctrl}
	mock.recorder = &MockAssignmentRepositoryMockRecorder{mock}
	return mock
}

type MockAssignmentService struct {
	ctrl     *gomock.Controller
	recorder *MockAssignmentServiceMockRecorder
}

type MockAssignmentServiceMockRecorder struct {
	mock *MockAssignmentService
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

func (m *MockAssignmentService) GetAssignment(ctx context.Context, id string) (*domain.Assignment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAssignment", ctx, id)
	ret0, _ := ret[0].(*domain.Assignment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockAssignmentServiceMockRecorder) GetAssignment(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAssignment", reflect.TypeOf((*MockAssignmentService)(nil).GetAssignment), ctx, id)
}

func (m *MockAssignmentService) UpdateAssignment(ctx context.Context, assignment *domain.Assignment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAssignment", ctx, assignment)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockAssignmentServiceMockRecorder) UpdateAssignment(ctx, assignment interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAssignment", reflect.TypeOf((*MockAssignmentService)(nil).UpdateAssignment), ctx, assignment)
}

func (m *MockAssignmentService) DeleteAssignment(ctx context.Context, id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAssignment", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockAssignmentServiceMockRecorder) DeleteAssignment(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAssignment", reflect.TypeOf((*MockAssignmentService)(nil).DeleteAssignment), ctx, id)
}

func (m *MockAssignmentService) ListAssignmentsByTutor(ctx context.Context, tutorID string) ([]*domain.Assignment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAssignmentsByTutor", ctx, tutorID)
	ret0, _ := ret[0].([]*domain.Assignment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockAssignmentServiceMockRecorder) ListAssignmentsByTutor(ctx, tutorID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAssignmentsByTutor", reflect.TypeOf((*MockAssignmentService)(nil).ListAssignmentsByTutor), ctx, tutorID)
}

func (m *MockAssignmentService) ListAssignmentsByStudent(ctx context.Context, studentID string) ([]*domain.Assignment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAssignmentsByStudent", ctx, studentID)
	ret0, _ := ret[0].([]*domain.Assignment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockAssignmentServiceMockRecorder) ListAssignmentsByStudent(ctx, studentID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAssignmentsByStudent", reflect.TypeOf((*MockAssignmentService)(nil).ListAssignmentsByStudent), ctx, studentID)
}
