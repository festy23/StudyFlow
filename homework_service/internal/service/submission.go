package service

import (
	"common_library/ctxdata"
	"context"
	"github.com/google/uuid"

	"homework_service/internal/domain"
	"homework_service/internal/repository"
)

type SubmissionServiceInterface interface {
	CreateSubmission(ctx context.Context, submission *domain.Submission) (*domain.Submission, error)
	GetSubmission(ctx context.Context, id uuid.UUID) (*domain.Submission, error)
	ListSubmissionsByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]*domain.Submission, error)
	GetSubmissionFileURL(ctx context.Context, id uuid.UUID) (string, error)
}

type submissionService struct {
	submissionRepo *repository.SubmissionRepository
	assignmentRepo *repository.AssignmentRepository
	fileClient     FileClient
}

func NewSubmissionService(
	submissionRepo *repository.SubmissionRepository,
	assignmentRepo *repository.AssignmentRepository,
	fileClient FileClient,
) SubmissionServiceInterface {
	return &submissionService{
		submissionRepo: submissionRepo,
		assignmentRepo: assignmentRepo,
		fileClient:     fileClient,
	}
}

func (s *submissionService) CreateSubmission(ctx context.Context, submission *domain.Submission) (*domain.Submission, error) {
	assignment, err := s.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		return nil, err
	}

	userId, ok := ctxdata.GetUserID(ctx)
	if !ok || userId != assignment.StudentID.String() {
		return nil, ErrPermissionDenied
	}

	if err := s.submissionRepo.Create(ctx, submission); err != nil {
		return nil, err
	}

	return submission, nil
}

func (s *submissionService) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	submission, err := s.submissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		return nil, err
	}

	userId, ok := ctxdata.GetUserID(ctx)
	if !ok || (userId != assignment.StudentID.String() && userId != assignment.TutorID.String()) {
		return nil, ErrPermissionDenied
	}

	return submission, nil
}

func (s *submissionService) ListSubmissionsByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]*domain.Submission, error) {
	assignment, err := s.assignmentRepo.GetByID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	userId, ok := ctxdata.GetUserID(ctx)
	if !ok || (userId != assignment.StudentID.String() && userId != assignment.TutorID.String()) {
		return nil, ErrPermissionDenied
	}

	return s.submissionRepo.ListByAssignment(ctx, assignmentID)
}

func (s *submissionService) GetSubmissionFileURL(ctx context.Context, id uuid.UUID) (string, error) {
	submission, err := s.submissionRepo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, submission.AssignmentID)
	if err != nil {
		return "", err
	}

	userId, ok := ctxdata.GetUserID(ctx)
	if !ok || (userId != assignment.StudentID.String() && userId != assignment.TutorID.String()) {
		return "", ErrPermissionDenied
	}

	if submission.FileID == nil {
		return "", ErrFileNotFound
	}

	url, err := s.fileClient.GetFileURL(ctx, *submission.FileID)
	if err != nil {
		return "", err
	}
	return url, nil
}
