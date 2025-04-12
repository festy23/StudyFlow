package grpc

import (
	"context"
	"errors"

	"homework_service/internal/domain"
	"homework_service/internal/repository"
	"homework_service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "homework_service/pkg/api"
)

type HomeworkHandler struct {
	v1.UnimplementedHomeworkServiceServer

	assignmentService service.AssignmentServiceInt
	submissionService *service.SubmissionService
	feedbackService   *service.FeedbackService
}

func NewHomeworkHandler(
	assignmentService service.AssignmentServiceInt,
	submissionService *service.SubmissionService,
	feedbackService *service.FeedbackService,
) *HomeworkHandler {
	return &HomeworkHandler{
		assignmentService: assignmentService,
		submissionService: submissionService,
		feedbackService:   feedbackService,
	}
}

func (h *HomeworkHandler) CreateAssignment(ctx context.Context, req *v1.CreateAssignmentRequest) (*v1.Assignment, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	if req.TutorId != userID {
		return nil, status.Error(codes.PermissionDenied, "can only create assignments for yourself")
	}

	assignment := &domain.Assignment{
		TutorID:     req.TutorId,
		StudentID:   req.StudentId,
		Title:       req.Title,
		Description: req.Description,
	}

	if req.FileId != "" {
		assignment.FileID = req.FileId
	}
	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		assignment.DueDate = &dueDate
	}

	createdAssignment, err := h.assignmentService.CreateAssignment(ctx, assignment)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toProtoAssignment(createdAssignment), nil
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

func toProtoAssignment(a *domain.Assignment) *v1.Assignment {
	assignment := &v1.Assignment{
		Id:          a.ID,
		TutorId:     a.TutorID,
		StudentId:   a.StudentID,
		Title:       a.Title,
		Description: a.Description,
		CreatedAt:   timestamppb.New(a.CreatedAt),
		EditedAt:    timestamppb.New(a.EditedAt),
	}

	if a.FileID != "" {
		assignment.FileId = a.FileID
	}
	if a.DueDate != nil {
		assignment.DueDate = timestamppb.New(*a.DueDate)
	}

	return assignment
}
