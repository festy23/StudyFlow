package service

import (
	"context"
	"errors"
	"time"

	"homework_service/internal/domain"
	"homework_service/internal/repository"
)

type AssignmentService struct {
	assignmentRepo *repository.AssignmentRepository
	userClient     UserClient
	fileClient     FileClient
}

func NewAssignmentService(assignmentRepo *repository.AssignmentRepository, userClient UserClient, fileClient FileClient) *AssignmentService {
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

	if req.FileID != nil {
		if !s.fileClient.FileExists(ctx, *req.FileID) {
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
