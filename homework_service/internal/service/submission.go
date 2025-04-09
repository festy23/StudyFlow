package service

import (
	"context"
	"errors"

	"homework_service/internal/app"
	"homework_service/internal/domain"
	"homework_service/internal/repository"
)

var (
	ErrInvalidSubmission = errors.New("invalid submission data")
)

type SubmissionService struct {
	submissionRepo *repository.SubmissionRepository
	assignmentRepo *repository.AssignmentRepository
	fileClient     *app.FileClient
}

func NewSubmissionService(
	submissionRepo *repository.SubmissionRepository,
	assignmentRepo *repository.AssignmentRepository,
	fileClient *app.FileClient,
) *SubmissionService {
	return &SubmissionService{
		submissionRepo: submissionRepo,
		assignmentRepo: assignmentRepo,
		fileClient:     fileClient,
	}
}

func (s *SubmissionService) CreateSubmission(ctx context.Context, submission *domain.Submission) (*domain.Submission, error) {
	if submission.AssignmentID == "" || submission.StudentID == "" || submission.FileID == "" {
		return nil, ErrInvalidSubmission
	}

	if _, err := s.fileClient.GetFile(ctx, submission.FileID); err != nil {
		return nil, err
	}

	if _, err := s.assignmentRepo.GetByID(ctx, submission.AssignmentID); err != nil {
		return nil, err
	}

	if err := s.submissionRepo.Create(ctx, submission); err != nil {
		return nil, err
	}

	return submission, nil
}

func (s *SubmissionService) GetSubmission(ctx context.Context, id string) (*domain.Submission, error) {
	return s.submissionRepo.GetByID(ctx, id)
}

func (s *SubmissionService) ListSubmissions(ctx context.Context, filter domain.SubmissionFilter) ([]*domain.Submission, error) {
	return s.submissionRepo.ListByFilter(ctx, filter)
}
