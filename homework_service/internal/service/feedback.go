package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"time"

	"common_library/ctxdata"
	"homework_service/internal/domain"
	"homework_service/internal/repository"
)

var (
	ErrFeedbackNotFound    = errors.New("feedback not found")
	ErrFileNotFound        = errors.New("file not found")
	ErrInvalidFeedbackData = errors.New("invalid feedback data")
	ErrSubmissionNotFound  = errors.New("submission not found")
	ErrAssignmentNotFound  = errors.New("assignment not found")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrInvalidArgument     = errors.New("invalid argument")
)

type FeedbackServiceInterface interface {
	CreateFeedback(ctx context.Context, feedback *domain.Feedback) (*domain.Feedback, error)
	GetFeedback(ctx context.Context, id uuid.UUID) (*domain.Feedback, error)
	UpdateFeedback(ctx context.Context, feedback *domain.Feedback) (*domain.Feedback, error)
	ListFeedbacksByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]*domain.Feedback, error)
	GetFeedbackFileURL(ctx context.Context, id uuid.UUID) (string, error)
}

type feedbackService struct {
	feedbackRepo   *repository.FeedbackRepository
	submissionRepo *repository.SubmissionRepository
	assignmentRepo *repository.AssignmentRepository
	fileClient     FileClient
}

func NewFeedbackService(
	feedbackRepo *repository.FeedbackRepository,
	submissionRepo *repository.SubmissionRepository,
	assignmentRepo *repository.AssignmentRepository,
	fileClient FileClient,
) FeedbackServiceInterface {
	return &feedbackService{
		feedbackRepo:   feedbackRepo,
		submissionRepo: submissionRepo,
		assignmentRepo: assignmentRepo,
		fileClient:     fileClient,
	}
}

func (s *feedbackService) CreateFeedback(ctx context.Context, feedback *domain.Feedback) (*domain.Feedback, error) {
	userRole, ok := ctxdata.GetUserRole(ctx)
	if !ok || userRole != "tutor" {
		return nil, ErrPermissionDenied
	}

	submission, err := s.submissionRepo.GetByID(ctx, feedback.SubmissionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSubmissionNotFound
		}
		return nil, err
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAssignmentNotFound
		}
		return nil, err
	}

	userID, ok := ctxdata.GetUserID(ctx)
	if !ok || assignment.TutorID.String() != userID {
		return nil, ErrPermissionDenied
	}

	now := time.Now()
	newFeedback := &domain.Feedback{
		SubmissionID: feedback.SubmissionID,
		FileID:       feedback.FileID,
		Comment:      feedback.Comment,
		CreatedAt:    now,
		EditedAt:     now,
	}

	if err := s.feedbackRepo.Create(ctx, newFeedback); err != nil {
		return nil, err
	}

	return newFeedback, nil
}

func (s *feedbackService) UpdateFeedback(ctx context.Context, feedback *domain.Feedback) (*domain.Feedback, error) {
	userRole, ok := ctxdata.GetUserRole(ctx)
	if !ok || userRole != "tutor" {
		return nil, ErrPermissionDenied
	}

	existingFeedback, err := s.feedbackRepo.GetByID(ctx, feedback.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrFeedbackNotFound
		}
		return nil, err
	}

	submission, err := s.submissionRepo.GetByID(ctx, existingFeedback.SubmissionID)
	if err != nil {
		return nil, err
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		return nil, err
	}

	userID, ok := ctxdata.GetUserID(ctx)
	if !ok || assignment.TutorID.String() != userID {
		return nil, ErrPermissionDenied
	}

	if feedback.Comment != nil {
		existingFeedback.Comment = feedback.Comment
	}

	if feedback.FileID != nil {
		existingFeedback.FileID = feedback.FileID
	}

	existingFeedback.EditedAt = time.Now()

	if err := s.feedbackRepo.Update(ctx, existingFeedback); err != nil {
		return nil, err
	}

	return existingFeedback, nil
}

func (s *feedbackService) GetFeedback(ctx context.Context, id uuid.UUID) (*domain.Feedback, error) {
	feedback, err := s.feedbackRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	submission, err := s.submissionRepo.GetByID(ctx, feedback.SubmissionID)
	if err != nil {
		return nil, err
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		return nil, err
	}

	userRole, ok := ctxdata.GetUserRole(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	if userRole == "tutor" && assignment.TutorID.String() != userID {
		return nil, ErrPermissionDenied
	}
	if userRole == "student" && assignment.StudentID.String() != userID {
		return nil, ErrPermissionDenied
	}

	return feedback, nil
}

func (s *feedbackService) ListFeedbacksByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]*domain.Feedback, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	userRole, ok := ctxdata.GetUserRole(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	if userRole == "tutor" && assignment.TutorID.String() != userID {
		return nil, ErrPermissionDenied
	}
	if userRole == "student" && assignment.StudentID.String() != userID {
		return nil, ErrPermissionDenied
	}

	return s.feedbackRepo.ListByAssignment(ctx, assignmentID)
}

func (s *feedbackService) GetFeedbackFileURL(ctx context.Context, id uuid.UUID) (string, error) {
	feedback, err := s.feedbackRepo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	if feedback.FileID == nil {
		return "", ErrFileNotFound
	}

	submission, err := s.submissionRepo.GetByID(ctx, feedback.SubmissionID)
	if err != nil {
		return "", err
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		return "", err
	}

	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return "", ErrPermissionDenied
	}

	userRole, ok := ctxdata.GetUserRole(ctx)
	if !ok {
		return "", ErrPermissionDenied
	}

	if userRole == "tutor" && assignment.TutorID.String() != userID {
		return "", ErrPermissionDenied
	}
	if userRole == "student" && assignment.StudentID.String() != userID {
		return "", ErrPermissionDenied
	}

	url, err := s.fileClient.GetFileURL(ctx, *feedback.FileID)
	if err != nil {
		return "", err
	}

	return url, nil
}
