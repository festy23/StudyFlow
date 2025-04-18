package service

import (
	"common_library/ctxdata"
	"context"
	"errors"
	"github.com/google/uuid"
	"time"

	"homework_service/internal/domain"
	"homework_service/internal/repository"
)

type AssignmentService struct {
	assignmentRepo repository.AssignmentRepository
	userClient     UserClient
	fileClient     FileClient
}

func NewAssignmentService(
	assignmentRepo repository.AssignmentRepository,
	userClient UserClient,
	fileClient FileClient,
) *AssignmentService {
	return &AssignmentService{
		assignmentRepo: assignmentRepo,
		userClient:     userClient,
		fileClient:     fileClient,
	}
}

func (s *AssignmentService) CreateAssignment(ctx context.Context, req *domain.Assignment) (*domain.Assignment, error) {

	userRole, ok := ctxdata.GetUserRole(ctx)
	if !ok || userRole != "tutor" {
		return nil, errors.New("permission denied")
	}

	isPair, err := s.userClient.IsPair(ctx, req.TutorID, req.StudentID)
	if err != nil {
		return nil, err
	}
	if !isPair {
		return nil, errors.New("not a tutor-student pair")
	}

	now := time.Now()
	assignment := &domain.Assignment{
		TutorID:     req.TutorID,
		StudentID:   req.StudentID,
		Title:       req.Title,
		Description: req.Description,
		FileID:      req.FileID,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		EditedAt:    now,
	}

	err = s.assignmentRepo.Create(ctx, assignment)
	if err != nil {
		return nil, err
	}

	return assignment, nil
}

func (s *AssignmentService) GetAssignment(ctx context.Context, id uuid.UUID) (*domain.Assignment, error) {
	assignment, err := s.assignmentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	if assignment.TutorID.String() != userID && assignment.StudentID.String() != userID {
		return nil, ErrPermissionDenied
	}

	return assignment, nil
}

func (s *AssignmentService) UpdateAssignment(ctx context.Context, assignment *domain.Assignment) error {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok || assignment.TutorID.String() != userID {
		return ErrPermissionDenied
	}

	return s.assignmentRepo.Update(ctx, assignment)
}

func (s *AssignmentService) DeleteAssignment(ctx context.Context, id uuid.UUID) error {
	assignment, err := s.assignmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	userID, ok := ctxdata.GetUserID(ctx)
	if !ok || assignment.TutorID.String() != userID {
		return ErrPermissionDenied
	}

	return s.assignmentRepo.Delete(ctx, id)
}

func (s *AssignmentService) ListAssignmentsByTutor(ctx context.Context, tutorID uuid.UUID, statuses []domain.AssignmentStatus) ([]*domain.Assignment, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok || tutorID.String() != userID {
		return nil, ErrPermissionDenied
	}

	return s.assignmentRepo.ListByFilter(ctx, domain.AssignmentFilter{TutorID: tutorID, Statuses: statuses})
}

func (s *AssignmentService) ListAssignmentsByStudent(ctx context.Context, studentID uuid.UUID, statuses []domain.AssignmentStatus) ([]*domain.Assignment, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok || studentID.String() != userID {
		return nil, ErrPermissionDenied
	}

	return s.assignmentRepo.ListByFilter(ctx, domain.AssignmentFilter{StudentID: studentID, Statuses: statuses})
}

func (s *AssignmentService) ListAssignmentsByPair(ctx context.Context, tutorID uuid.UUID, studentID uuid.UUID, statuses []domain.AssignmentStatus) ([]*domain.Assignment, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok || (tutorID.String() != userID && studentID.String() != userID) {
		return nil, ErrPermissionDenied
	}

	return s.assignmentRepo.ListByFilter(ctx, domain.AssignmentFilter{TutorID: tutorID, StudentID: studentID, Statuses: statuses})
}

func (s *AssignmentService) GetAssignmentFileURL(ctx context.Context, id uuid.UUID) (string, error) {
	assignment, err := s.assignmentRepo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if assignment.FileID == nil {
		return "", ErrFileNotFound
	}
	url, err := s.fileClient.GetFileURL(ctx, *assignment.FileID)
	if err != nil {
		return "", err
	}
	return url, nil
}
