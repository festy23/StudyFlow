package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"homework_service/internal/domain"
	"homework_service/internal/repository"
	"homework_service/internal/service"
	v1 "homework_service/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockAssignmentService struct {
	mock.Mock
}

func (m *MockAssignmentService) CreateAssignment(ctx context.Context, assignment *domain.Assignment) (*domain.Assignment, error) {
	args := m.Called(ctx, assignment)
	return args.Get(0).(*domain.Assignment), args.Error(1)
}

// Add other required methods from service.AssignmentService interface
func (m *MockAssignmentService) GetAssignment(ctx context.Context, id string) (*domain.Assignment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Assignment), args.Error(1)
}

// Add any other methods that are part of the service.AssignmentService interface

func TestHomeworkHandler_CreateAssignment(t *testing.T) {
	now := time.Now()
	dueDate := now.Add(24 * time.Hour)

	tests := []struct {
		name        string
		req         *v1.CreateAssignmentRequest
		ctx         context.Context
		mockSetup   func(*MockAssignmentService)
		want        *v1.Assignment
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful creation",
			req: &v1.CreateAssignmentRequest{
				TutorId:     "tutor1",
				StudentId:   "student1",
				Title:       "Test Assignment",
				Description: "Test Description",
				FileId:      "file1",
				DueDate:     timestamppb.New(dueDate),
			},
			ctx: context.WithValue(context.Background(), "user_id", "tutor1"),
			mockSetup: func(m *MockAssignmentService) {
				m.On("CreateAssignment", mock.Anything, mock.AnythingOfType("*domain.Assignment")).
					Return(&domain.Assignment{
						ID:          "assign1",
						TutorID:     "tutor1",
						StudentID:   "student1",
						Title:       "Test Assignment",
						Description: "Test Description",
						FileID:      "file1",
						DueDate:     &dueDate,
						CreatedAt:   now,
						EditedAt:    now,
					}, nil)
			},
			want: &v1.Assignment{
				Id:          "assign1",
				TutorId:     "tutor1",
				StudentId:   "student1",
				Title:       "Test Assignment",
				Description: "Test Description",
				FileId:      "file1",
				DueDate:     timestamppb.New(dueDate),
				CreatedAt:   timestamppb.New(now),
				EditedAt:    timestamppb.New(now),
			},
			wantErr: false,
		},
		{
			name: "missing user id in context",
			req: &v1.CreateAssignmentRequest{
				TutorId: "tutor1",
			},
			ctx:         context.Background(),
			mockSetup:   func(m *MockAssignmentService) {},
			wantErr:     true,
			expectedErr: status.Error(codes.Unauthenticated, "user id not found"),
		},
		{
			name: "permission denied - creating for another tutor",
			req: &v1.CreateAssignmentRequest{
				TutorId: "tutor2",
			},
			ctx:         context.WithValue(context.Background(), "user_id", "tutor1"),
			mockSetup:   func(m *MockAssignmentService) {},
			wantErr:     true,
			expectedErr: status.Error(codes.PermissionDenied, "can only create assignments for yourself"),
		},
		{
			name: "service returns not found error",
			req: &v1.CreateAssignmentRequest{
				TutorId:   "tutor1",
				StudentId: "student1",
			},
			ctx: context.WithValue(context.Background(), "user_id", "tutor1"),
			mockSetup: func(m *MockAssignmentService) {
				m.On("CreateAssignment", mock.Anything, mock.Anything).
					Return(&domain.Assignment{}, repository.ErrNotFound)
			},
			wantErr:     true,
			expectedErr: status.Error(codes.NotFound, repository.ErrNotFound.Error()),
		},
		{
			name: "service returns permission denied error",
			req: &v1.CreateAssignmentRequest{
				TutorId:   "tutor1",
				StudentId: "student1",
			},
			ctx: context.WithValue(context.Background(), "user_id", "tutor1"),
			mockSetup: func(m *MockAssignmentService) {
				m.On("CreateAssignment", mock.Anything, mock.Anything).
					Return(&domain.Assignment{}, service.ErrPermissionDenied)
			},
			wantErr:     true,
			expectedErr: status.Error(codes.PermissionDenied, service.ErrPermissionDenied.Error()),
		},
		{
			name: "service returns internal error",
			req: &v1.CreateAssignmentRequest{
				TutorId:   "tutor1",
				StudentId: "student1",
			},
			ctx: context.WithValue(context.Background(), "user_id", "tutor1"),
			mockSetup: func(m *MockAssignmentService) {
				m.On("CreateAssignment", mock.Anything, mock.Anything).
					Return(&domain.Assignment{}, errors.New("some unexpected error"))
			},
			wantErr:     true,
			expectedErr: status.Error(codes.Internal, "internal server error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAssignmentService)
			tt.mockSetup(mockService)

			handler := &HomeworkHandler{
				assignmentService: mockService,
			}

			got, err := handler.CreateAssignment(tt.ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestToGRPCError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected error
	}{
		{
			name:     "not found error",
			input:    repository.ErrNotFound,
			expected: status.Error(codes.NotFound, repository.ErrNotFound.Error()),
		},
		{
			name:     "permission denied error",
			input:    service.ErrPermissionDenied,
			expected: status.Error(codes.PermissionDenied, service.ErrPermissionDenied.Error()),
		},
		{
			name:     "invalid argument error",
			input:    service.ErrInvalidArgument,
			expected: status.Error(codes.InvalidArgument, service.ErrInvalidArgument.Error()),
		},
		{
			name:     "generic error",
			input:    errors.New("some error"),
			expected: status.Error(codes.Internal, "internal server error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toGRPCError(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToProtoAssignment(t *testing.T) {
	now := time.Now()
	dueDate := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		input    *domain.Assignment
		expected *v1.Assignment
	}{
		{
			name: "complete assignment",
			input: &domain.Assignment{
				ID:          "assign1",
				TutorID:     "tutor1",
				StudentID:   "student1",
				Title:       "Test",
				Description: "Description",
				FileID:      "file1",
				DueDate:     &dueDate,
				CreatedAt:   now,
				EditedAt:    now,
			},
			expected: &v1.Assignment{
				Id:          "assign1",
				TutorId:     "tutor1",
				StudentId:   "student1",
				Title:       "Test",
				Description: "Description",
				FileId:      "file1",
				DueDate:     timestamppb.New(dueDate),
				CreatedAt:   timestamppb.New(now),
				EditedAt:    timestamppb.New(now),
			},
		},
		{
			name: "assignment without optional fields",
			input: &domain.Assignment{
				ID:          "assign1",
				TutorID:     "tutor1",
				StudentID:   "student1",
				Title:       "Test",
				Description: "Description",
				CreatedAt:   now,
				EditedAt:    now,
			},
			expected: &v1.Assignment{
				Id:          "assign1",
				TutorId:     "tutor1",
				StudentId:   "student1",
				Title:       "Test",
				Description: "Description",
				CreatedAt:   timestamppb.New(now),
				EditedAt:    timestamppb.New(now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toProtoAssignment(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
