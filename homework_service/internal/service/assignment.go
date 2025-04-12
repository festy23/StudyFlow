package service

import (
	"context"
	"errors"
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
	if req.TutorID == "" || req.StudentID == "" || req.Title == "" || req.Description == "" {
		return nil, errors.New("invalid arguments")
	}

	userRole, ok := ctx.Value("user_role").(string)
	if !ok || userRole != "tutor" {
		return nil, errors.New("permission denied")
	}

	if !s.userClient.UserExists(ctx, req.StudentID) {
		return nil, errors.New("student not found")
	}

	if !s.userClient.IsPair(ctx, req.TutorID, req.StudentID) {
		return nil, errors.New("not a tutor-student pair")
	}

	if req.FileID != "" {
		if !s.fileClient.FileExists(ctx, req.FileID) {
			return nil, errors.New("file not found")
		}
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

	err := s.assignmentRepo.Create(ctx, assignment)
	if err != nil {
		return nil, err
	}

	return assignment, nil
}

func (s *AssignmentService) GetAssignment(ctx context.Context, id string) (*domain.Assignment, error) {
	assignment, err := s.assignmentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, ErrPermissionDenied
	}

	if assignment.TutorID != userID && assignment.StudentID != userID {
		return nil, ErrPermissionDenied
	}

	return assignment, nil
}

func (s *AssignmentService) UpdateAssignment(ctx context.Context, assignment *domain.Assignment) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || assignment.TutorID != userID {
		return ErrPermissionDenied
	}

	existing, err := s.assignmentRepo.GetByID(ctx, assignment.ID)
	if err != nil {
		return err
	}

	if existing.TutorID != assignment.TutorID {
		return ErrPermissionDenied
	}

	return s.assignmentRepo.Update(ctx, assignment)
}

func (s *AssignmentService) DeleteAssignment(ctx context.Context, id string) error {
	assignment, err := s.assignmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	userID, ok := ctx.Value("user_id").(string)
	if !ok || assignment.TutorID != userID {
		return ErrPermissionDenied
	}

	return s.assignmentRepo.Delete(ctx, id)
}

func (s *AssignmentService) ListAssignmentsByTutor(ctx context.Context, tutorID string) ([]*domain.Assignment, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || tutorID != userID {
		return nil, ErrPermissionDenied
	}

	return s.assignmentRepo.ListByTutorID(ctx, tutorID)
}

func (s *AssignmentService) ListAssignmentsByStudent(ctx context.Context, studentID string) ([]*domain.Assignment, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || studentID != userID {
		return nil, ErrPermissionDenied
	}

	return s.assignmentRepo.ListByStudentID(ctx, studentID)
}
