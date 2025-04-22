package homework_grpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"homework_service/internal/domain"
	"homework_service/internal/repository"
	"homework_service/internal/service"
	v1 "homework_service/pkg/api"
	"homework_service/pkg/logger"

	handler "homework_service/internal/server/homework_grpc"
)

type MockAssignmentService struct {
	mock.Mock
}

func (m *MockAssignmentService) CreateAssignment(ctx context.Context, assignment *domain.Assignment) (*domain.Assignment, error) {
	args := m.Called(ctx, assignment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Assignment), args.Error(1)
}

func (m *MockAssignmentService) GetAssignment(ctx context.Context, id uuid.UUID) (*domain.Assignment, error) {
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

func (m *MockAssignmentService) DeleteAssignment(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAssignmentService) ListAssignmentsByTutor(ctx context.Context, tutorID uuid.UUID, statuses []domain.AssignmentStatus) ([]*domain.Assignment, error) {
	args := m.Called(ctx, tutorID, statuses)
	return args.Get(0).([]*domain.Assignment), args.Error(1)
}

func (m *MockAssignmentService) ListAssignmentsByStudent(ctx context.Context, studentID uuid.UUID, statuses []domain.AssignmentStatus) ([]*domain.Assignment, error) {
	args := m.Called(ctx, studentID, statuses)
	return args.Get(0).([]*domain.Assignment), args.Error(1)
}

func (m *MockAssignmentService) ListAssignmentsByPair(ctx context.Context, tutorID, studentID uuid.UUID, statuses []domain.AssignmentStatus) ([]*domain.Assignment, error) {
	args := m.Called(ctx, tutorID, studentID, statuses)
	return args.Get(0).([]*domain.Assignment), args.Error(1)
}

func (m *MockAssignmentService) GetAssignmentFileURL(ctx context.Context, id uuid.UUID) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

type MockSubmissionService struct {
	mock.Mock
}

func (m *MockSubmissionService) CreateSubmission(ctx context.Context, submission *domain.Submission) (*domain.Submission, error) {
	args := m.Called(ctx, submission)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Submission), args.Error(1)
}

func (m *MockSubmissionService) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Submission), args.Error(1)
}

func (m *MockSubmissionService) ListSubmissionsByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]*domain.Submission, error) {
	args := m.Called(ctx, assignmentID)
	return args.Get(0).([]*domain.Submission), args.Error(1)
}

func (m *MockSubmissionService) GetSubmissionFileURL(ctx context.Context, id uuid.UUID) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

type MockFeedbackService struct {
	mock.Mock
}

func (m *MockFeedbackService) CreateFeedback(ctx context.Context, feedback *domain.Feedback) (*domain.Feedback, error) {
	args := m.Called(ctx, feedback)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Feedback), args.Error(1)
}

func (m *MockFeedbackService) GetFeedback(ctx context.Context, id uuid.UUID) (*domain.Feedback, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Feedback), args.Error(1)
}

func (m *MockFeedbackService) UpdateFeedback(ctx context.Context, feedback *domain.Feedback) (*domain.Feedback, error) {
	args := m.Called(ctx, feedback)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Feedback), args.Error(1)
}

func (m *MockFeedbackService) ListFeedbacksByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]*domain.Feedback, error) {
	args := m.Called(ctx, assignmentID)
	return args.Get(0).([]*domain.Feedback), args.Error(1)
}

func (m *MockFeedbackService) GetFeedbackFileURL(ctx context.Context, id uuid.UUID) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func TestHomeworkHandler(t *testing.T) {
	log := logger.New()
	ctx := context.Background()

	str := func(s string) *string {
		return &s
	}

	t.Run("CreateAssignment - success", func(t *testing.T) {
		assignmentService := &MockAssignmentService{}
		submissionService := &MockSubmissionService{}
		feedbackService := &MockFeedbackService{}

		h := handler.NewHomeworkHandler(
			service.AssignmentService{},
			submissionService,
			feedbackService,
			log,
		)

		tutorID := uuid.New()
		studentID := uuid.New()
		fileID := uuid.New()
		dueDate := time.Now().Add(24 * time.Hour)
		title := "Test Assignment"
		description := "Test Description"
		fileIDStr := fileID.String()

		expectedAssignment := &domain.Assignment{
			ID:          uuid.New(),
			TutorID:     tutorID,
			StudentID:   studentID,
			Title:       str(title),
			Description: str(description),
			FileID:      &fileID,
			DueDate:     &dueDate,
			CreatedAt:   time.Now(),
			EditedAt:    time.Now(),
		}

		assignmentService.On("CreateAssignment", ctx, mock.AnythingOfType("*domain.Assignment")).
			Return(expectedAssignment, nil)

		str := func(s string) *string {
			return &s
		}

		resp, err := h.CreateAssignment(ctx, &v1.CreateAssignmentRequest{
			TutorId:     tutorID.String(),
			StudentId:   studentID.String(),
			Title:       str(title),
			Description: str(description),
			FileId:      &fileIDStr,
			DueDate:     timestamppb.New(dueDate),
		})

		assert.NoError(t, err)
		assert.Equal(t, expectedAssignment.ID.String(), resp.Id)
		assert.Equal(t, title, resp.Title)
		assert.Equal(t, description, resp.Description)
		assert.Equal(t, fileID.String(), *resp.FileId)
		assert.Equal(t, dueDate.Unix(), resp.DueDate.AsTime().Unix())
	})

	t.Run("CreateAssignment - invalid tutor ID", func(t *testing.T) {
		submissionService := &MockSubmissionService{}
		feedbackService := &MockFeedbackService{}

		h := handler.NewHomeworkHandler(
			service.AssignmentService{},
			submissionService,
			feedbackService,
			log,
		)

		str := func(s string) *string {
			return &s
		}

		title := "Test"

		_, err := h.CreateAssignment(ctx, &v1.CreateAssignmentRequest{
			TutorId:   "invalid-uuid",
			StudentId: uuid.New().String(),
			Title:     str(title),
		})

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("GetAssignment - not found", func(t *testing.T) {
		assignmentService := &MockAssignmentService{}
		submissionService := &MockSubmissionService{}
		feedbackService := &MockFeedbackService{}

		str := func(s string) *string {
			return &s
		}

		title := "Test"

		h := handler.NewHomeworkHandler(
			service.AssignmentService{},
			submissionService,
			feedbackService,
			log,
		)

		assignmentID := uuid.New()
		assignmentService.On("GetAssignment", ctx, assignmentID).
			Return((*domain.Assignment)(nil), repository.ErrNotFound)

		_, err := h.UpdateAssignment(ctx, &v1.UpdateAssignmentRequest{
			Id:    assignmentID.String(),
			Title: str(title),
		})

		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("CreateSubmission - success", func(t *testing.T) {
		submissionService := &MockSubmissionService{}
		feedbackService := &MockFeedbackService{}

		h := handler.NewHomeworkHandler(
			service.AssignmentService{},
			submissionService,
			feedbackService,
			log,
		)

		str := func(s string) *string {
			return &s
		}

		assignmentID := uuid.New()
		fileID := uuid.New()
		comment := "Test submission"
		fileIDStr := fileID.String()
		expectedSubmission := &domain.Submission{
			ID:           uuid.New(),
			AssignmentID: assignmentID,
			Comment:      str(comment),
			FileID:       &fileID,
			CreatedAt:    time.Now(),
			EditedAt:     time.Now(),
		}

		submissionService.On("CreateSubmission", ctx, mock.AnythingOfType("*domain.Submission")).
			Return(expectedSubmission, nil)

		resp, err := h.CreateSubmission(ctx, &v1.CreateSubmissionRequest{
			AssignmentId: assignmentID.String(),
			Comment:      str(comment),
			FileId:       &fileIDStr,
		})

		assert.NoError(t, err)
		assert.Equal(t, expectedSubmission.ID.String(), resp.Id)
		assert.Equal(t, comment, resp.Comment)
		assert.Equal(t, fileID.String(), *resp.FileId)
	})

	t.Run("CreateFeedback - success", func(t *testing.T) {
		submissionService := &MockSubmissionService{}
		feedbackService := &MockFeedbackService{}

		h := handler.NewHomeworkHandler(
			service.AssignmentService{},
			submissionService,
			feedbackService,
			log,
		)

		submissionID := uuid.New()
		fileID := uuid.New()
		comment := "Test feedback"
		fileIDStr := fileID.String()
		expectedFeedback := &domain.Feedback{
			ID:           uuid.New(),
			SubmissionID: submissionID,
			Comment:      str(comment),
			FileID:       &fileID,
			CreatedAt:    time.Now(),
			EditedAt:     time.Now(),
		}

		feedbackService.On("CreateFeedback", ctx, mock.AnythingOfType("*domain.Feedback")).
			Return(expectedFeedback, nil)

		resp, err := h.CreateFeedback(ctx, &v1.CreateFeedbackRequest{
			SubmissionId: submissionID.String(),
			Comment:      str(comment),
			FileId:       &fileIDStr,
		})

		assert.NoError(t, err)
		assert.Equal(t, expectedFeedback.ID.String(), resp.Id)
		assert.Equal(t, comment, resp.Comment)
		assert.Equal(t, fileID.String(), *resp.FileId)
	})

	t.Run("GetAssignmentFileURL - success", func(t *testing.T) {
		assignmentService := &MockAssignmentService{}
		submissionService := &MockSubmissionService{}
		feedbackService := &MockFeedbackService{}

		h := handler.NewHomeworkHandler(
			service.AssignmentService{},
			submissionService,
			feedbackService,
			log,
		)

		assignmentID := uuid.New()
		fileURL := "http://example.com/file.pdf"
		assignmentService.On("GetAssignmentFileURL", ctx, assignmentID).
			Return(fileURL, nil)

		resp, err := h.GetAssignmentFile(ctx, &v1.GetAssignmentFileRequest{
			AssignmentId: assignmentID.String(),
		})

		assert.NoError(t, err)
		assert.Equal(t, fileURL, resp.Url)
	})
}
