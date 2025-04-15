package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"homework_service/internal/domain"
	"homework_service/internal/service"
	"homework_service/internal/service/mocks"

	"homework_service/internal/repository"
	v1 "homework_service/pkg/api"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHomeworkHandler_CreateAssignment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAssignmentService := mocks.NewMockAssignmentService(ctrl)
	handler := NewHomeworkHandler(mockAssignmentService, nil, nil, nil, nil)

	now := time.Now()
	dueDate := now.Add(24 * time.Hour)
	validRequest := &v1.CreateAssignmentRequest{
		TutorId:     "tutor-123",
		StudentId:   "student-456",
		Title:       "Test Assignment",
		Description: "Test Description",
		FileId:      "file-789",
		DueDate:     timestamppb.New(dueDate),
	}

	tests := []struct {
		name          string
		ctx           context.Context
		req           *v1.CreateAssignmentRequest
		mockSetup     func()
		expectedResp  *v1.Assignment
		expectedError error
	}{
		{
			name: "Success",
			ctx:  context.WithValue(context.Background(), "user_id", "tutor-123"),
			req:  validRequest,
			mockSetup: func() {
				mockAssignmentService.EXPECT().CreateAssignment(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, a *domain.Assignment) (*domain.Assignment, error) {
						assert.Equal(t, "tutor-123", a.TutorID)
						assert.Equal(t, "student-456", a.StudentID)
						assert.Equal(t, "Test Assignment", a.Title)
						assert.Equal(t, "Test Description", a.Description)
						assert.Equal(t, "file-789", a.FileID)
						assert.Equal(t, dueDate, *a.DueDate)

						return &domain.Assignment{
							ID:          "assignment-123",
							TutorID:     a.TutorID,
							StudentID:   a.StudentID,
							Title:       a.Title,
							Description: a.Description,
							FileID:      a.FileID,
							DueDate:     a.DueDate,
							CreatedAt:   now,
							EditedAt:    now,
						}, nil
					})
			},
			expectedResp: &v1.Assignment{
				Id:          "assignment-123",
				TutorId:     "tutor-123",
				StudentId:   "student-456",
				Title:       "Test Assignment",
				Description: "Test Description",
				FileId:      "file-789",
				DueDate:     timestamppb.New(dueDate),
				CreatedAt:   timestamppb.New(now),
				EditedAt:    timestamppb.New(now),
			},
		},
		{
			name:          "Unauthenticated - no user_id in context",
			ctx:           context.Background(),
			req:           validRequest,
			mockSetup:     func() {},
			expectedError: status.Error(codes.Unauthenticated, "user id not found"),
		},
		{
			name:          "PermissionDenied - tutor_id != user_id",
			ctx:           context.WithValue(context.Background(), "user_id", "different-tutor"),
			req:           validRequest,
			mockSetup:     func() {},
			expectedError: status.Error(codes.PermissionDenied, "can only create assignments for yourself"),
		},
		{
			name: "Service error - not found",
			ctx:  context.WithValue(context.Background(), "user_id", "tutor-123"),
			req:  validRequest,
			mockSetup: func() {
				mockAssignmentService.EXPECT().CreateAssignment(gomock.Any(), gomock.Any()).
					Return(nil, repository.ErrNotFound)
			},
			expectedError: status.Error(codes.NotFound, repository.ErrNotFound.Error()),
		},
		{
			name: "Service error - permission denied",
			ctx:  context.WithValue(context.Background(), "user_id", "tutor-123"),
			req:  validRequest,
			mockSetup: func() {
				mockAssignmentService.EXPECT().CreateAssignment(gomock.Any(), gomock.Any()).
					Return(nil, service.ErrPermissionDenied)
			},
			expectedError: status.Error(codes.PermissionDenied, service.ErrPermissionDenied.Error()),
		},
		{
			name: "Service error - invalid argument",
			ctx:  context.WithValue(context.Background(), "user_id", "tutor-123"),
			req:  validRequest,
			mockSetup: func() {
				mockAssignmentService.EXPECT().CreateAssignment(gomock.Any(), gomock.Any()).
					Return(nil, service.ErrInvalidArgument)
			},
			expectedError: status.Error(codes.InvalidArgument, service.ErrInvalidArgument.Error()),
		},
		{
			name: "Service error - internal",
			ctx:  context.WithValue(context.Background(), "user_id", "tutor-123"),
			req:  validRequest,
			mockSetup: func() {
				mockAssignmentService.EXPECT().CreateAssignment(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some unexpected error"))
			},
			expectedError: status.Error(codes.Internal, "internal server error"),
		},
		{
			name: "Success - optional fields omitted",
			ctx:  context.WithValue(context.Background(), "user_id", "tutor-123"),
			req: &v1.CreateAssignmentRequest{
				TutorId:     "tutor-123",
				StudentId:   "student-456",
				Title:       "Test Assignment",
				Description: "Test Description",
			},
			mockSetup: func() {
				mockAssignmentService.EXPECT().CreateAssignment(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, a *domain.Assignment) (*domain.Assignment, error) {
						assert.Empty(t, a.FileID)
						assert.Nil(t, a.DueDate)

						return &domain.Assignment{
							ID:          "assignment-123",
							TutorID:     a.TutorID,
							StudentID:   a.StudentID,
							Title:       a.Title,
							Description: a.Description,
							CreatedAt:   now,
							EditedAt:    now,
						}, nil
					})
			},
			expectedResp: &v1.Assignment{
				Id:          "assignment-123",
				TutorId:     "tutor-123",
				StudentId:   "student-456",
				Title:       "Test Assignment",
				Description: "Test Description",
				CreatedAt:   timestamppb.New(now),
				EditedAt:    timestamppb.New(now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := handler.CreateAssignment(tt.ctx, tt.req)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
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
			name:     "NotFound",
			input:    repository.ErrNotFound,
			expected: status.Error(codes.NotFound, repository.ErrNotFound.Error()),
		},
		{
			name:     "PermissionDenied",
			input:    service.ErrPermissionDenied,
			expected: status.Error(codes.PermissionDenied, service.ErrPermissionDenied.Error()),
		},
		{
			name:     "InvalidArgument",
			input:    service.ErrInvalidArgument,
			expected: status.Error(codes.InvalidArgument, service.ErrInvalidArgument.Error()),
		},
		{
			name:     "InternalError",
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
			name: "All fields",
			input: &domain.Assignment{
				ID:          "assignment-123",
				TutorID:     "tutor-123",
				StudentID:   "student-456",
				Title:       "Test Assignment",
				Description: "Test Description",
				FileID:      "file-789",
				DueDate:     &dueDate,
				CreatedAt:   now,
				EditedAt:    now,
			},
			expected: &v1.Assignment{
				Id:          "assignment-123",
				TutorId:     "tutor-123",
				StudentId:   "student-456",
				Title:       "Test Assignment",
				Description: "Test Description",
				FileId:      "file-789",
				DueDate:     timestamppb.New(dueDate),
				CreatedAt:   timestamppb.New(now),
				EditedAt:    timestamppb.New(now),
			},
		},
		{
			name: "Optional fields omitted",
			input: &domain.Assignment{
				ID:          "assignment-123",
				TutorID:     "tutor-123",
				StudentID:   "student-456",
				Title:       "Test Assignment",
				Description: "Test Description",
				CreatedAt:   now,
				EditedAt:    now,
			},
			expected: &v1.Assignment{
				Id:          "assignment-123",
				TutorId:     "tutor-123",
				StudentId:   "student-456",
				Title:       "Test Assignment",
				Description: "Test Description",
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
