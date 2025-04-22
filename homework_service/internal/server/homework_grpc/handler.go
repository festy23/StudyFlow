package homework_grpc

import (
	"context"
	"errors"
	"github.com/google/uuid"

	"go.uber.org/zap"

	"homework_service/internal/domain"
	"homework_service/internal/repository"
	"homework_service/internal/service"

	v1 "homework_service/pkg/api"
	"homework_service/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type HomeworkHandler struct {
	v1.UnimplementedHomeworkServiceServer

	assignmentService service.AssignmentService
	submissionService service.SubmissionServiceInterface
	feedbackService   service.FeedbackServiceInterface
	logger            *logger.Logger
}

func RegisterHomeworkServiceServer(grpcServer *grpc.Server, server v1.HomeworkServiceServer) {
	v1.RegisterHomeworkServiceServer(grpcServer, server)
}

func NewHomeworkHandler(
	assignmentService service.AssignmentService,
	submissionService service.SubmissionServiceInterface,
	feedbackService service.FeedbackServiceInterface,
	logger *logger.Logger,
) *HomeworkHandler {
	return &HomeworkHandler{
		assignmentService: assignmentService,
		submissionService: submissionService,
		feedbackService:   feedbackService,
		logger:            logger,
	}
}

func (h *HomeworkHandler) CreateAssignment(ctx context.Context, req *v1.CreateAssignmentRequest) (*v1.Assignment, error) {
	tutorId, err := uuid.Parse(req.TutorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	studentId, err := uuid.Parse(req.StudentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	assignment := &domain.Assignment{
		TutorID:     tutorId,
		StudentID:   studentId,
		Title:       req.Title,
		Description: req.Description,
	}

	if req.FileId != nil {
		fileId, err := uuid.Parse(*req.FileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		assignment.FileID = &fileId
	}
	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		assignment.DueDate = &dueDate
	}

	createdAssignment, err := h.assignmentService.CreateAssignment(ctx, assignment)
	if err != nil {
		h.logger.Error("failed to create assignment",
			zap.Error(err),
			zap.Any("assignment", assignment),
		)
		return nil, toGRPCError(err)
	}

	h.logger.Info("assignment created successfully",
		zap.String("assignment_id", createdAssignment.ID.String()),
	)

	return toProtoAssignment(createdAssignment), nil
}

func (h *HomeworkHandler) UpdateAssignment(ctx context.Context, req *v1.UpdateAssignmentRequest) (*v1.Assignment, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	assignment, err := h.assignmentService.GetAssignment(ctx, id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	updatedAssignment := *assignment
	updatedAssignment.Title = req.Title
	updatedAssignment.Description = req.Description

	if req.FileId != nil {
		fileId, err := uuid.Parse(*req.FileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		updatedAssignment.FileID = &fileId
	}

	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		updatedAssignment.DueDate = &dueDate
	}

	err = h.assignmentService.UpdateAssignment(ctx, &updatedAssignment)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toProtoAssignment(&updatedAssignment), nil
}

func (h *HomeworkHandler) DeleteAssignment(ctx context.Context, req *v1.DeleteAssignmentRequest) (*v1.Empty, error) {
	id, err := uuid.Parse(req.AssignmentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = h.assignmentService.DeleteAssignment(ctx, id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.Empty{}, nil
}

func (h *HomeworkHandler) ListAssignmentsByTutor(ctx context.Context, req *v1.ListAssignmentsByTutorRequest) (*v1.ListAssignmentsResponse, error) {
	statuses := make([]domain.AssignmentStatus, len(req.StatusFilter))
	for ind, s := range req.StatusFilter {
		statuses[ind] = domain.AssignmentStatus(s)
	}

	tutorId, err := uuid.Parse(req.TutorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	assignments, err := h.assignmentService.ListAssignmentsByTutor(ctx, tutorId, statuses)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.ListAssignmentsResponse{
		Assignments: toProtoAssignments(assignments),
	}, nil
}

func (h *HomeworkHandler) ListAssignmentsByStudent(ctx context.Context, req *v1.ListAssignmentsByStudentRequest) (*v1.ListAssignmentsResponse, error) {
	statuses := make([]domain.AssignmentStatus, len(req.StatusFilter))
	for ind, s := range req.StatusFilter {
		statuses[ind] = domain.AssignmentStatus(s)
	}
	studentId, err := uuid.Parse(req.StudentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	assignments, err := h.assignmentService.ListAssignmentsByStudent(ctx, studentId, statuses)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.ListAssignmentsResponse{
		Assignments: toProtoAssignments(assignments),
	}, nil
}

func (h *HomeworkHandler) ListAssignmentsByPair(ctx context.Context, req *v1.ListAssignmentsByPairRequest) (*v1.ListAssignmentsResponse, error) {
	statuses := make([]domain.AssignmentStatus, len(req.StatusFilter))
	for ind, s := range req.StatusFilter {
		statuses[ind] = domain.AssignmentStatus(s)
	}
	studentId, err := uuid.Parse(req.StudentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tutorId, err := uuid.Parse(req.TutorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	assignments, err := h.assignmentService.ListAssignmentsByPair(ctx, tutorId, studentId, statuses)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.ListAssignmentsResponse{
		Assignments: toProtoAssignments(assignments),
	}, nil
}

func (h *HomeworkHandler) CreateSubmission(ctx context.Context, req *v1.CreateSubmissionRequest) (*v1.Submission, error) {
	assignmentId, err := uuid.Parse(req.AssignmentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var fileId *uuid.UUID
	if req.FileId != nil {
		id, err := uuid.Parse(*req.FileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		fileId = &id
	}
	submission := &domain.Submission{
		AssignmentID: assignmentId,
		Comment:      req.Comment,
		FileID:       fileId,
	}

	createdSubmission, err := h.submissionService.CreateSubmission(ctx, submission)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toProtoSubmission(createdSubmission), nil
}

func (h *HomeworkHandler) ListSubmissionsByAssignment(ctx context.Context, req *v1.ListSubmissionsByAssignmentRequest) (*v1.ListSubmissionsResponse, error) {
	assignmentId, err := uuid.Parse(req.AssignmentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	submissions, err := h.submissionService.ListSubmissionsByAssignment(ctx, assignmentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.ListSubmissionsResponse{
		Submissions: toProtoSubmissions(submissions),
	}, nil
}

func (h *HomeworkHandler) CreateFeedback(ctx context.Context, req *v1.CreateFeedbackRequest) (*v1.Feedback, error) {
	submissionId, err := uuid.Parse(req.SubmissionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var fileId *uuid.UUID
	if req.FileId != nil {
		id, err := uuid.Parse(*req.FileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		fileId = &id
	}
	feedback := &domain.Feedback{
		SubmissionID: submissionId,
		Comment:      req.Comment,
		FileID:       fileId,
	}

	createdFeedback, err := h.feedbackService.CreateFeedback(ctx, feedback)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toProtoFeedback(createdFeedback), nil
}

func (h *HomeworkHandler) UpdateFeedback(ctx context.Context, req *v1.UpdateFeedbackRequest) (*v1.Feedback, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	existingFeedback, err := h.feedbackService.GetFeedback(ctx, id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	update := &domain.Feedback{
		ID:           id,
		SubmissionID: existingFeedback.SubmissionID,
	}

	if req.Comment != nil {
		update.Comment = req.Comment
	}

	if req.FileId != nil {
		fileId, err := uuid.Parse(*req.FileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		update.FileID = &fileId
	}

	updatedFeedback, err := h.feedbackService.UpdateFeedback(ctx, update)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toProtoFeedback(updatedFeedback), nil
}

func (h *HomeworkHandler) ListFeedbacksByAssignment(ctx context.Context, req *v1.ListFeedbacksByAssignmentRequest) (*v1.ListFeedbacksResponse, error) {
	id, err := uuid.Parse(req.AssignmentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	feedbacks, err := h.feedbackService.ListFeedbacksByAssignment(ctx, id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.ListFeedbacksResponse{
		Feedbacks: toProtoFeedbacks(feedbacks),
	}, nil
}

func (h *HomeworkHandler) GetAssignmentFile(ctx context.Context, req *v1.GetAssignmentFileRequest) (*v1.HomeworkFileURL, error) {
	id, err := uuid.Parse(req.AssignmentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	url, err := h.assignmentService.GetAssignmentFileURL(ctx, id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.HomeworkFileURL{Url: url}, nil
}

func (h *HomeworkHandler) GetSubmissionFile(ctx context.Context, req *v1.GetSubmissionFileRequest) (*v1.HomeworkFileURL, error) {
	id, err := uuid.Parse(req.SubmissionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	url, err := h.submissionService.GetSubmissionFileURL(ctx, id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.HomeworkFileURL{Url: url}, nil
}

func (h *HomeworkHandler) GetFeedbackFile(ctx context.Context, req *v1.GetFeedbackFileRequest) (*v1.HomeworkFileURL, error) {
	id, err := uuid.Parse(req.FeedbackId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	url, err := h.feedbackService.GetFeedbackFileURL(ctx, id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.HomeworkFileURL{Url: url}, nil
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
		Id:          a.ID.String(),
		TutorId:     a.TutorID.String(),
		StudentId:   a.StudentID.String(),
		Title:       a.Title,
		Description: a.Description,
		CreatedAt:   timestamppb.New(a.CreatedAt),
		EditedAt:    timestamppb.New(a.EditedAt),
	}

	if a.FileID != nil {
		id := a.FileID.String()
		assignment.FileId = &id
	}
	if a.DueDate != nil {
		assignment.DueDate = timestamppb.New(*a.DueDate)
	}

	return assignment
}

func toProtoAssignments(assignments []*domain.Assignment) []*v1.Assignment {
	var protoAssignments []*v1.Assignment
	for _, a := range assignments {
		protoAssignments = append(protoAssignments, toProtoAssignment(a))
	}
	return protoAssignments
}

func toProtoSubmission(s *domain.Submission) *v1.Submission {
	submission := &v1.Submission{
		Id:           s.ID.String(),
		AssignmentId: s.AssignmentID.String(),
		Comment:      s.Comment,
		CreatedAt:    timestamppb.New(s.CreatedAt),
		EditedAt:     timestamppb.New(s.EditedAt),
	}

	if s.FileID != nil {
		id := s.FileID.String()
		submission.FileId = &id
	}

	return submission
}

func toProtoSubmissions(submissions []*domain.Submission) []*v1.Submission {
	var protoSubmissions []*v1.Submission
	for _, s := range submissions {
		protoSubmissions = append(protoSubmissions, toProtoSubmission(s))
	}
	return protoSubmissions
}

func toProtoFeedback(f *domain.Feedback) *v1.Feedback {
	feedback := &v1.Feedback{
		Id:           f.ID.String(),
		SubmissionId: f.SubmissionID.String(),
		Comment:      f.Comment,
		CreatedAt:    timestamppb.New(f.CreatedAt),
		EditedAt:     timestamppb.New(f.EditedAt),
	}

	if f.FileID != nil {
		id := f.FileID.String()
		feedback.FileId = &id
	}

	return feedback
}

func toProtoFeedbacks(feedbacks []*domain.Feedback) []*v1.Feedback {
	var protoFeedbacks []*v1.Feedback
	for _, f := range feedbacks {
		protoFeedbacks = append(protoFeedbacks, toProtoFeedback(f))
	}
	return protoFeedbacks
}
