package grpc

import (
	"common_library/ctxdata"
	"context"
	"errors"
	"go.uber.org/zap"

	"homework_service/internal/domain"
	"homework_service/internal/repository"
	"homework_service/internal/service"

	v1 "homework_service/pkg/api"
	"homework_service/pkg/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FileServiceClient interface {
	GetFileURL(ctx context.Context, fileID string) (string, error)
}

type HomeworkHandler struct {
	v1.UnimplementedHomeworkServiceServer

	assignmentService service.AssignmentService
	submissionService service.SubmissionServiceInterface
	feedbackService   service.FeedbackServiceInterface
	fileService       FileServiceClient
	logger            *logger.Logger
}

func NewHomeworkHandler(
	assignmentService service.AssignmentService,
	submissionService service.SubmissionServiceInterface,
	feedbackService service.FeedbackServiceInterface,
	fileService FileServiceClient,
	logger *logger.Logger,
) *HomeworkHandler {
	return &HomeworkHandler{
		assignmentService: assignmentService,
		submissionService: submissionService,
		feedbackService:   feedbackService,
		fileService:       fileService,
		logger:            logger,
	}
}

func (h *HomeworkHandler) CreateAssignment(ctx context.Context, req *v1.CreateAssignmentRequest) (*v1.Assignment, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		h.logger.Error("user id not found in context")
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	userRole, ok := ctxdata.GetUserRole(ctx)
	if !ok {
		h.logger.Error("user role not found in context")
		return nil, status.Error(codes.Unauthenticated, "user role not found")
	}

	if userRole != "tutor" {
		return nil, status.Error(codes.PermissionDenied, "only tutors can create assignments")
	}

	if req.TutorId != userID {
		h.logger.Warn("permission denied: user tried to create assignment for another tutor",
			zap.String("requested_tutor_id", req.TutorId),
			zap.String("actual_user_id", userID),
		)
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
		h.logger.Error("failed to create assignment",
			zap.Error(err),
			zap.Any("assignment", assignment),
		)
		return nil, toGRPCError(err)
	}

	h.logger.Info("assignment created successfully",
		zap.String("assignment_id", createdAssignment.ID),
	)

	return toProtoAssignment(createdAssignment), nil
}

func (h *HomeworkHandler) UpdateAssignment(ctx context.Context, req *v1.UpdateAssignmentRequest) (*v1.Assignment, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "can only update own assignments")
	}

	updatedAssignment := *assignment
	updatedAssignment.Title = req.Title
	updatedAssignment.Description = req.Description

	if req.FileId != "" {
		updatedAssignment.FileID = req.FileId
	} else {
		updatedAssignment.FileID = ""
	}

	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		updatedAssignment.DueDate = &dueDate
	} else {
		updatedAssignment.DueDate = nil
	}

	err = h.assignmentService.UpdateAssignment(ctx, &updatedAssignment)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toProtoAssignment(&updatedAssignment), nil
}

func (h *HomeworkHandler) DeleteAssignment(ctx context.Context, req *v1.DeleteAssignmentRequest) (*v1.Empty, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, req.AssignmentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "can only delete own assignments")
	}

	err = h.assignmentService.DeleteAssignment(ctx, req.AssignmentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.Empty{}, nil
}

func (h *HomeworkHandler) ListAssignmentsByTutor(ctx context.Context, req *v1.ListAssignmentsByTutorRequest) (*v1.ListAssignmentsResponse, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	if req.TutorId != userID {
		return nil, status.Error(codes.PermissionDenied, "can only view own assignments")
	}

	var statuses []domain.AssignmentStatus
	for _, s := range req.StatusFilter {
		statuses = append(statuses, domain.AssignmentStatus(s))
	}

	assignments, err := h.assignmentService.ListAssignmentsByTutor(ctx, req.TutorId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.ListAssignmentsResponse{
		Assignments: toProtoAssignments(assignments),
	}, nil
}

func (h *HomeworkHandler) ListAssignmentsByStudent(ctx context.Context, req *v1.ListAssignmentsByStudentRequest) (*v1.ListAssignmentsResponse, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	if req.StudentId != userID {
		return nil, status.Error(codes.PermissionDenied, "can only view own assignments")
	}

	var statuses []domain.AssignmentStatus
	for _, s := range req.StatusFilter {
		statuses = append(statuses, domain.AssignmentStatus(s))
	}

	assignments, err := h.assignmentService.ListAssignmentsByStudent(ctx, req.StudentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.ListAssignmentsResponse{
		Assignments: toProtoAssignments(assignments),
	}, nil
}

func (h *HomeworkHandler) ListAssignmentsByPair(ctx context.Context, req *v1.ListAssignmentsByPairRequest) (*v1.ListAssignmentsResponse, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	if req.StudentId != userID && req.TutorId != userID {
		return nil, status.Error(codes.PermissionDenied, "must be part of the pair to view assignments")
	}

	var statuses []domain.AssignmentStatus
	for _, s := range req.StatusFilter {
		statuses = append(statuses, domain.AssignmentStatus(s))
	}

	assignments, err := h.assignmentService.ListAssignmentsByPair(ctx, req.TutorId, req.StudentId, statuses)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.ListAssignmentsResponse{
		Assignments: toProtoAssignments(assignments),
	}, nil
}

func (h *HomeworkHandler) CreateSubmission(ctx context.Context, req *v1.CreateSubmissionRequest) (*v1.Submission, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, req.AssignmentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.StudentID != userID {
		return nil, status.Error(codes.PermissionDenied, "can only submit for own assignments")
	}

	submission := &domain.Submission{
		AssignmentID: req.AssignmentId,
		Comment:      req.Comment,
		FileID:       req.FileId,
	}

	createdSubmission, err := h.submissionService.CreateSubmission(ctx, submission)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toProtoSubmission(createdSubmission), nil
}

func (h *HomeworkHandler) ListSubmissionsByAssignment(ctx context.Context, req *v1.ListSubmissionsByAssignmentRequest) (*v1.ListSubmissionsResponse, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, req.AssignmentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.StudentID != userID && assignment.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "must be part of the assignment to view submissions")
	}

	submissions, err := h.submissionService.ListSubmissionsByAssignment(ctx, req.AssignmentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.ListSubmissionsResponse{
		Submissions: toProtoSubmissions(submissions),
	}, nil
}

func (h *HomeworkHandler) CreateFeedback(ctx context.Context, req *v1.CreateFeedbackRequest) (*v1.Feedback, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	userRole, ok := ctxdata.GetUserRole(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user role not found")
	}

	if userRole != "tutor" {
		return nil, status.Error(codes.PermissionDenied, "only tutors can create feedback")
	}

	submission, err := h.submissionService.GetSubmission(ctx, req.SubmissionId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, submission.AssignmentID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "can only create feedback for own assignments")
	}

	feedback := &domain.Feedback{
		SubmissionID: req.SubmissionId,
		Comment:      req.Comment,
		FileID:       &req.FileId,
	}

	createdFeedback, err := h.feedbackService.CreateFeedback(ctx, feedback)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toProtoFeedback(createdFeedback), nil
}

func (h *HomeworkHandler) UpdateFeedback(ctx context.Context, req *v1.UpdateFeedbackRequest) (*v1.Feedback, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	feedback, err := h.feedbackService.GetFeedback(ctx, req.SubmissionId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	submission, err := h.submissionService.GetSubmission(ctx, feedback.SubmissionID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, submission.AssignmentID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "can only update own feedback")
	}

	update := &domain.Feedback{
		ID:           feedback.ID,
		SubmissionID: req.SubmissionId,
		Comment:      req.Comment,
		FileID:       &req.FileId,
	}

	updatedFeedback, err := h.feedbackService.UpdateFeedback(ctx, update)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toProtoFeedback(updatedFeedback), nil
}

func (h *HomeworkHandler) ListFeedbacksByAssignment(ctx context.Context, req *v1.ListFeedbacksByAssignmentRequest) (*v1.ListFeedbacksResponse, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, req.AssignmentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.StudentID != userID && assignment.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "must be part of the assignment to view feedbacks")
	}

	feedbacks, err := h.feedbackService.ListFeedbacksByAssignment(ctx, req.AssignmentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.ListFeedbacksResponse{
		Feedbacks: toProtoFeedbacks(feedbacks),
	}, nil
}

func (h *HomeworkHandler) GetAssignmentFile(ctx context.Context, req *v1.GetAssignmentFileRequest) (*v1.HomeworkFileURL, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, req.AssignmentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.StudentID != userID && assignment.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "must be part of the assignment to view files")
	}

	if assignment.FileID == "" {
		return nil, status.Error(codes.NotFound, "assignment has no file")
	}

	url, err := h.fileService.GetFileURL(ctx, assignment.FileID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.HomeworkFileURL{Url: url}, nil
}

func (h *HomeworkHandler) GetSubmission(ctx context.Context, req *v1.GetSubmissionFileRequest) (*v1.Submission, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	submission, err := h.submissionService.GetSubmission(ctx, req.SubmissionId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, submission.AssignmentID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.StudentID != userID && assignment.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "must be part of the assignment to view submission")
	}

	return toProtoSubmission(submission), nil
}

func (h *HomeworkHandler) GetSubmissionFile(ctx context.Context, req *v1.GetSubmissionFileRequest) (*v1.HomeworkFileURL, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	submission, err := h.submissionService.GetSubmission(ctx, req.SubmissionId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, submission.AssignmentID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.StudentID != userID && assignment.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "must be part of the assignment to view files")
	}

	if submission.FileID == "" {
		return nil, status.Error(codes.NotFound, "submission has no file")
	}

	url, err := h.fileService.GetFileURL(ctx, submission.FileID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &v1.HomeworkFileURL{Url: url}, nil
}

func (h *HomeworkHandler) GetFeedbackFile(ctx context.Context, req *v1.GetFeedbackFileRequest) (*v1.HomeworkFileURL, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user id not found")
	}

	feedback, err := h.feedbackService.GetFeedback(ctx, req.FeedbackId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	submission, err := h.submissionService.GetSubmission(ctx, feedback.SubmissionID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	assignment, err := h.assignmentService.GetAssignment(ctx, submission.AssignmentID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	if assignment.StudentID != userID && assignment.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "must be part of the assignment to view files")
	}

	if feedback.FileID == nil {
		return nil, status.Error(codes.NotFound, "feedback has no file")
	}

	url, err := h.fileService.GetFileURL(ctx, *feedback.FileID)
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

func toProtoAssignments(assignments []*domain.Assignment) []*v1.Assignment {
	var protoAssignments []*v1.Assignment
	for _, a := range assignments {
		protoAssignments = append(protoAssignments, toProtoAssignment(a))
	}
	return protoAssignments
}

func toProtoSubmission(s *domain.Submission) *v1.Submission {
	submission := &v1.Submission{
		Id:           s.ID,
		AssignmentId: s.AssignmentID,
		Comment:      s.Comment,
		CreatedAt:    timestamppb.New(s.CreatedAt),
		EditedAt:     timestamppb.New(s.EditedAt),
	}

	if s.FileID != "" {
		submission.FileId = s.FileID
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
		Id:           f.ID,
		SubmissionId: f.SubmissionID,
		Comment:      f.Comment,
		CreatedAt:    timestamppb.New(f.CreatedAt),
		EditedAt:     timestamppb.New(f.EditedAt),
	}

	if f.FileID != nil {
		feedback.FileId = *f.FileID
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
