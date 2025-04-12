package service_test

import (
	"context"
	"errors"
	"testing"

	"homework_service/internal/app"
	"homework_service/internal/domain"
	"homework_service/internal/repository/mocks"
	"homework_service/internal/service"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAssignmentService_CreateAssignment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAssignmentRepo := mocks.NewMockAssignmentRepository(ctrl)
	mockUserClient := service.NewMockUserClient(ctrl)
	mockFileClient := service.NewMockFileClient(ctrl)

	assignmentService := service.NewAssignmentService(mockAssignmentRepo, mockUserClient, mockFileClient)

	tests := []struct {
		name          string
		setupMocks    func()
		input         *domain.Assignment
		expectedError error
	}{
		{
			name: "Success",
			setupMocks: func() {
				mockUserClient.EXPECT().UserExists(gomock.Any(), "student-123").Return(true)
				mockUserClient.EXPECT().IsPair(gomock.Any(), "tutor-123", "student-123").Return(true)
				mockAssignmentRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			input: &domain.Assignment{
				TutorID:     "tutor-123",
				StudentID:   "student-123",
				Title:       "Test Assignment",
				Description: "Test Description",
			},
		},
		{
			name: "Invalid arguments",
			input: &domain.Assignment{
				TutorID:     "",
				StudentID:   "",
				Title:       "",
				Description: "",
			},
			expectedError: errors.New("invalid arguments"),
		},
		{
			name: "Student not found",
			setupMocks: func() {
				mockUserClient.EXPECT().UserExists(gomock.Any(), "student-123").Return(false)
			},
			input: &domain.Assignment{
				TutorID:     "tutor-123",
				StudentID:   "student-123",
				Title:       "Test Assignment",
				Description: "Test Description",
			},
			expectedError: errors.New("student not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			ctx := context.WithValue(context.Background(), "user_role", "tutor")
			_, err := assignmentService.CreateAssignment(ctx, tt.input)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFeedbackService_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockSubmissionRepo := mocks.NewMockSubmissionRepository(ctrl)
	mockAssignmentRepo := mocks.NewMockAssignmentRepository(ctrl)
	mockFileClient := service.NewMockFileClient(ctrl)

	feedbackService := service.NewFeedbackService(
		mockFeedbackRepo,
		mockSubmissionRepo,
		mockAssignmentRepo,
		mockFileClient,
	)

	assignment := &domain.Assignment{
		ID:        "assignment-123",
		TutorID:   "tutor-123",
		StudentID: "student-456",
	}
	submission := &domain.Submission{
		ID:           "submission-123",
		AssignmentID: "assignment-123",
		StudentID:    "student-456",
	}
	feedback := &domain.Feedback{
		SubmissionID: "submission-123",
		Comment:      "Good job!",
	}

	t.Run("CreateFeedback Success", func(t *testing.T) {
		mockSubmissionRepo.EXPECT().GetByID(gomock.Any(), "submission-123").Return(submission, nil)
		mockAssignmentRepo.EXPECT().GetByID(gomock.Any(), "assignment-123").Return(assignment, nil)
		mockFeedbackRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		ctx := context.WithValue(context.Background(), "user_role", "tutor")
		ctx = context.WithValue(ctx, "user_id", "tutor-123")
		_, err := feedbackService.CreateFeedback(ctx, feedback)
		assert.NoError(t, err)
	})

	t.Run("CreateFeedback Permission Denied", func(t *testing.T) {
		mockSubmissionRepo.EXPECT().GetByID(gomock.Any(), "submission-123").Return(submission, nil)
		mockAssignmentRepo.EXPECT().GetByID(gomock.Any(), "assignment-123").Return(assignment, nil)

		ctx := context.WithValue(context.Background(), "user_role", "tutor")
		ctx = context.WithValue(ctx, "user_id", "wrong-tutor") // Different tutor
		_, err := feedbackService.CreateFeedback(ctx, feedback)
		assert.Equal(t, service.ErrPermissionDenied, err)
	})
}

func TestSubmissionService_CreateSubmission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubmissionRepo := mocks.NewMockSubmissionRepository(ctrl)
	mockAssignmentRepo := mocks.NewMockAssignmentRepository(ctrl)
	mockFileClient := app.NewMockFileClient(ctrl)

	submissionService := service.NewSubmissionService(
		mockSubmissionRepo,
		mockAssignmentRepo,
		mockFileClient,
	)

	t.Run("Success", func(t *testing.T) {
		submission := &domain.Submission{
			AssignmentID: "assignment-123",
			StudentID:    "student-456",
			FileID:       "file-789",
		}

		mockFileClient.EXPECT().GetFile(gomock.Any(), "file-789").Return(nil, nil)
		mockAssignmentRepo.EXPECT().GetByID(gomock.Any(), "assignment-123").Return(&domain.Assignment{}, nil)
		mockSubmissionRepo.EXPECT().Create(gomock.Any(), submission).Return(nil)

		_, err := submissionService.CreateSubmission(context.Background(), submission)
		assert.NoError(t, err)
	})

	t.Run("Invalid submission data", func(t *testing.T) {
		_, err := submissionService.CreateSubmission(context.Background(), &domain.Submission{})
		assert.Equal(t, service.ErrInvalidSubmission, err)
	})
}

func TestServiceIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAssignmentRepo := mocks.NewMockAssignmentRepository(ctrl)
	mockSubmissionRepo := mocks.NewMockSubmissionRepository(ctrl)
	mockFeedbackRepo := mocks.NewMockFeedbackRepository(ctrl)
	mockUserClient := service.NewMockUserClient(ctrl)
	mockFileClient := app.NewMockFileClient(ctrl)

	assignmentService := service.NewAssignmentService(
		mockAssignmentRepo,
		mockUserClient,
		mockFileClient,
	)

	submissionService := service.NewSubmissionService(
		mockSubmissionRepo,
		mockAssignmentRepo,
		mockFileClient,
	)

	feedbackService := service.NewFeedbackService(
		mockFeedbackRepo,
		mockSubmissionRepo,
		mockAssignmentRepo,
		mockFileClient,
	)

	assignment := &domain.Assignment{
		ID:          "assignment-123",
		TutorID:     "tutor-123",
		StudentID:   "student-456",
		Title:       "Math Homework",
		Description: "Solve problems",
	}

	submission := &domain.Submission{
		ID:           "submission-123",
		AssignmentID: "assignment-123",
		StudentID:    "student-456",
		FileID:       "file-submission-123",
	}

	feedback := &domain.Feedback{
		ID:           "feedback-123",
		SubmissionID: "submission-123",
		Comment:      "Well done",
	}

	t.Run("Full workflow", func(t *testing.T) {
		mockUserClient.EXPECT().UserExists(gomock.Any(), "student-456").Return(true)
		mockUserClient.EXPECT().IsPair(gomock.Any(), "tutor-123", "student-456").Return(true)
		mockAssignmentRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, a *domain.Assignment) error {
				a.ID = "assignment-123"
				return nil
			})

		ctx := context.WithValue(context.Background(), "user_role", "tutor")
		createdAssignment, err := assignmentService.CreateAssignment(ctx, assignment)
		assert.NoError(t, err)
		assert.Equal(t, "assignment-123", createdAssignment.ID)

		mockFileClient.EXPECT().GetFile(gomock.Any(), "file-submission-123").Return(nil, nil)
		mockAssignmentRepo.EXPECT().GetByID(gomock.Any(), "assignment-123").Return(assignment, nil)
		mockSubmissionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, s *domain.Submission) error {
				s.ID = "submission-123"
				return nil
			})

		ctx = context.WithValue(context.Background(), "user_role", "student")
		createdSubmission, err := submissionService.CreateSubmission(ctx, submission)
		assert.NoError(t, err)
		assert.Equal(t, "submission-123", createdSubmission.ID)

		mockSubmissionRepo.EXPECT().GetByID(gomock.Any(), "submission-123").Return(createdSubmission, nil)
		mockAssignmentRepo.EXPECT().GetByID(gomock.Any(), "assignment-123").Return(assignment, nil)
		mockFeedbackRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, f *domain.Feedback) error {
				f.ID = "feedback-123"
				return nil
			})

		ctx = context.WithValue(context.Background(), "user_role", "tutor")
		ctx = context.WithValue(ctx, "user_id", "tutor-123")
		createdFeedback, err := feedbackService.CreateFeedback(ctx, feedback)
		assert.NoError(t, err)
		assert.Equal(t, "feedback-123", createdFeedback.ID)
	})
}
