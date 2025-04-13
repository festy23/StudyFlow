package service

import (
	"context"
	"errors"
	"time"

	"homework_service/internal/domain"
	"homework_service/internal/repository"
)

var (
	ErrFeedbackNotFound    = errors.New("feedback not found")
	ErrInvalidFeedbackData = errors.New("invalid feedback data")
	ErrSubmissionNotFound  = errors.New("submission not found")
	ErrAssignmentNotFound  = errors.New("assignment not found")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrInvalidArgument     = errors.New("invalid argument")
)

type FeedbackService struct {
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
) *FeedbackService {
	return &FeedbackService{
		feedbackRepo:   feedbackRepo,
		submissionRepo: submissionRepo,
		assignmentRepo: assignmentRepo,
		fileClient:     fileClient,
	}
}

func (s *FeedbackService) CreateFeedback(ctx context.Context, feedback *domain.Feedback) (*domain.Feedback, error) {
	userRole, ok := ctx.Value("user_role").(string)
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

	userID, ok := ctx.Value("user_id").(string)
	if !ok || assignment.TutorID != userID {
		return nil, ErrPermissionDenied
	}

	if feedback.FileID != nil {
		if !s.fileClient.FileExists(ctx, *feedback.FileID) {
			return nil, errors.New("file not found")
		}
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

func (s *FeedbackService) UpdateFeedback(ctx context.Context, feedback *domain.Feedback) (*domain.Feedback, error) {
	userRole, ok := ctx.Value("user_role").(string)
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

	userID, ok := ctx.Value("user_id").(string)
	if !ok || assignment.TutorID != userID {
		return nil, ErrPermissionDenied
	}

	if feedback.FileID != nil {
		if !s.fileClient.FileExists(ctx, *feedback.FileID) {
			return nil, errors.New("file not found")
		}
		existingFeedback.FileID = feedback.FileID
	}

	if feedback.Comment != "nothing" {
		existingFeedback.Comment = feedback.Comment
	}

	existingFeedback.EditedAt = time.Now()

	if err := s.feedbackRepo.Update(ctx, existingFeedback); err != nil {
		return nil, err
	}

	return existingFeedback, nil
}

func (s *FeedbackService) ListFeedbacks(ctx context.Context, filter domain.FeedbackFilter) ([]*domain.Feedback, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, ErrPermissionDenied
	}

	userRole, ok := ctx.Value("user_role").(string)
	if !ok {
		return nil, ErrPermissionDenied
	}

	if userRole == "student" {
		filter.StudentID = userID
	}

	if userRole == "tutor" {
		filter.TutorID = userID
	}

	return s.feedbackRepo.ListByFilter(ctx, filter)
}

func (s *FeedbackService) GetFeedback(ctx context.Context, id string) (*domain.Feedback, error) {
	feedback, err := s.feedbackRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	userID, ok := ctx.Value("user_id").(string)
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

	userRole, ok := ctx.Value("user_role").(string)
	if !ok {
		return nil, ErrPermissionDenied
	}

	if userRole == "tutor" && assignment.TutorID != userID {
		return nil, ErrPermissionDenied
	}
	if userRole == "student" && assignment.StudentID != userID {
		return nil, ErrPermissionDenied
	}

	return feedback, nil
}
