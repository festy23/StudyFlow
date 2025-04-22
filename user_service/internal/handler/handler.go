package handler

import (
	"common_library/logging"
	"context"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"slices"
	"userservice/internal/errdefs"
	"userservice/internal/model"
	pb "userservice/pkg/api"
)

type UserService interface {
	RegisterViaTelegram(ctx context.Context, input *model.RegisterViaTelegramInput) (*model.User, error)
	Authorize(ctx context.Context, input *model.AuthorizeInput) (*model.User, error)
	GetMe(ctx context.Context) (*model.User, error)
	GetUserPublic(ctx context.Context, id uuid.UUID) (*model.UserPublic, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input *model.UpdateUserInput) (*model.User, error)
	GetTutorProfile(ctx context.Context, userId uuid.UUID) (*model.TutorProfile, error)
	UpdateTutorProfile(ctx context.Context, userId uuid.UUID, input *model.UpdateTutorProfileInput) (*model.TutorProfile, error)
	CreateTutorStudent(ctx context.Context, input *model.CreateTutorStudentInput) (*model.TutorStudent, error)
	GetTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) (*model.TutorStudent, error)
	UpdateTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID, input *model.UpdateTutorStudentInput) (*model.TutorStudent, error)
	DeleteTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) error
	ListTutorStudents(ctx context.Context, tutorId uuid.UUID) ([]*model.TutorStudent, error)
	ListTutorStudentsForStudent(ctx context.Context, studentId uuid.UUID) ([]*model.TutorStudent, error)
	ResolveTutorStudentContext(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) (*model.TutorStudentContext, error)
	AcceptInvitationFromTutor(ctx context.Context, tutorId uuid.UUID) error
}

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	service UserService
}

func NewUserServiceServer(userService UserService) *UserServiceServer {
	return &UserServiceServer{service: userService}
}

func (h *UserServiceServer) RegisterViaTelegram(ctx context.Context, req *pb.RegisterViaTelegramRequest) (*pb.User, error) {
	input := &model.RegisterViaTelegramInput{
		TelegramId: req.TelegramId,
		Role:       model.Role(req.Role),
		Username:   req.Username,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Timezone:   req.Timezone,
	}
	if logger, ok := logging.GetFromContext(ctx); ok {
		logger.Info(ctx, "registering user", zap.Any("input", input))
	}
	user, err := h.service.RegisterViaTelegram(ctx, input)
	if err != nil {
		return nil, mapError(err, errdefs.ErrAlreadyExists, errdefs.ValidationErr)
	}

	return toPbUser(user), nil
}

func (h *UserServiceServer) AuthorizeByAuthHeader(ctx context.Context, req *pb.AuthorizeByAuthHeaderRequest) (*pb.User, error) {
	input := &model.AuthorizeInput{
		AuthorizationHeader: req.GetAuthorizationHeader(),
	}

	user, err := h.service.Authorize(ctx, input)
	if err != nil {
		return nil, mapError(err, errdefs.ValidationErr, errdefs.AuthenticationErr)
	}

	return toPbUser(user), nil
}

func (h *UserServiceServer) GetMe(ctx context.Context, _ *pb.Empty) (*pb.User, error) {
	user, err := h.service.GetMe(ctx)

	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound)
	}

	return toPbUser(user), nil
}

func (h *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserPublic, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	user, err := h.service.GetUserPublic(ctx, id)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound)
	}

	userPb := &pb.UserPublic{
		Id:        user.Id.String(),
		Role:      user.Role.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	return userPb, nil
}

func (h *UserServiceServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	input := &model.UpdateUserInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Timezone:  req.Timezone,
	}

	user, err := h.service.UpdateUser(ctx, id, input)

	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ValidationErr, errdefs.ErrPermissionDenied)
	}

	return toPbUser(user), nil
}

func (h *UserServiceServer) GetTutorProfileByUserId(ctx context.Context, req *pb.GetTutorProfileByUserIdRequest) (*pb.TutorProfile, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	profile, err := h.service.GetTutorProfile(ctx, id)

	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrPermissionDenied)
	}

	return toPbTutorProfile(profile), nil
}

func (h *UserServiceServer) UpdateTutorProfile(ctx context.Context, req *pb.UpdateTutorProfileRequest) (*pb.TutorProfile, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	input := &model.UpdateTutorProfileInput{
		PaymentInfo:          req.PaymentInfo,
		LessonPriceRub:       req.LessonPriceRub,
		LessonConnectionLink: req.LessonConnectionLink,
	}

	profile, err := h.service.UpdateTutorProfile(ctx, id, input)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ValidationErr, errdefs.ErrPermissionDenied)
	}

	return toPbTutorProfile(profile), nil
}

func (h *UserServiceServer) CreateTutorStudent(ctx context.Context, req *pb.CreateTutorStudentRequest) (*pb.TutorStudent, error) {
	tutorId, err := uuid.Parse(req.TutorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	studentId, err := uuid.Parse(req.StudentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	input := &model.CreateTutorStudentInput{
		TutorId:        tutorId,
		StudentId:      studentId,
		LessonPriceRub: req.LessonPriceRub,
	}

	tutorStudent, err := h.service.CreateTutorStudent(ctx, input)
	if err != nil {
		return nil, mapError(err, errdefs.ErrAlreadyExists, errdefs.ValidationErr, errdefs.ErrPermissionDenied)
	}

	return toPbTutorStudent(tutorStudent), nil
}

func (h *UserServiceServer) GetTutorStudent(ctx context.Context, req *pb.GetTutorStudentRequest) (*pb.TutorStudent, error) {
	tutorId, err := uuid.Parse(req.TutorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	studentId, err := uuid.Parse(req.StudentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	tutorStudent, err := h.service.GetTutorStudent(ctx, tutorId, studentId)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrPermissionDenied, errdefs.AuthenticationErr)
	}
	return toPbTutorStudent(tutorStudent), nil
}

func (h *UserServiceServer) UpdateTutorStudent(ctx context.Context, req *pb.UpdateTutorStudentRequest) (*pb.TutorStudent, error) {
	tutorId, err := uuid.Parse(req.TutorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	studentId, err := uuid.Parse(req.StudentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	input := &model.UpdateTutorStudentInput{
		LessonPriceRub:       req.LessonPriceRub,
		LessonConnectionLink: req.LessonConnectionLink,
	}

	if req.Status != nil {
		s, ok := model.TutorStudentStatusFromString(*req.Status)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "invalid status")
		}
		input.Status = &s
	}

	tutorStudent, err := h.service.UpdateTutorStudent(ctx, tutorId, studentId, input)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrPermissionDenied, errdefs.ErrPermissionDenied)
	}

	return toPbTutorStudent(tutorStudent), nil
}

func (h *UserServiceServer) DeleteTutorStudent(ctx context.Context, req *pb.DeleteTutorStudentRequest) (*pb.Empty, error) {
	tutorId, err := uuid.Parse(req.TutorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	studentId, err := uuid.Parse(req.StudentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = h.service.DeleteTutorStudent(ctx, tutorId, studentId)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrPermissionDenied)
	}

	return &pb.Empty{}, nil
}

func (h *UserServiceServer) ListTutorStudents(ctx context.Context, req *pb.ListTutorStudentsRequest) (*pb.ListTutorStudentsResponse, error) {
	tutorId, err := uuid.Parse(req.TutorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	tutorStudents, err := h.service.ListTutorStudents(ctx, tutorId)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrPermissionDenied)
	}

	resp := make([]*pb.TutorStudent, len(tutorStudents))
	for i, tutorStudent := range tutorStudents {
		resp[i] = toPbTutorStudent(tutorStudent)
	}

	return &pb.ListTutorStudentsResponse{Students: resp}, nil
}

func (h *UserServiceServer) ListTutorsForStudent(ctx context.Context, req *pb.ListTutorsForStudentRequest) (*pb.ListTutorsForStudentResponse, error) {
	studentId, err := uuid.Parse(req.StudentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	tutorStudents, err := h.service.ListTutorStudentsForStudent(ctx, studentId)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrPermissionDenied)
	}

	resp := make([]*pb.TutorStudent, len(tutorStudents))
	for i, tutorStudent := range tutorStudents {
		resp[i] = toPbTutorStudent(tutorStudent)
	}

	return &pb.ListTutorsForStudentResponse{Tutors: resp}, nil
}

func (h *UserServiceServer) ResolveTutorStudentContext(ctx context.Context, req *pb.ResolveTutorStudentContextRequest) (*pb.ResolvedTutorStudentContext, error) {
	tutorId, err := uuid.Parse(req.TutorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	studentId, err := uuid.Parse(req.StudentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	result, err := h.service.ResolveTutorStudentContext(ctx, tutorId, studentId)
	if err != nil {
		return nil, mapError(err, errdefs.ErrNotFound, errdefs.ErrPermissionDenied)
	}

	resp := &pb.ResolvedTutorStudentContext{
		RelationshipStatus:   result.RelationshipStatus.String(),
		LessonPriceRub:       result.LessonPriceRub,
		LessonConnectionLink: result.LessonConnectionLink,
		PaymentInfo:          result.PaymentInfo,
	}

	return resp, nil
}

func (h *UserServiceServer) AcceptInvitationFromTutor(ctx context.Context, req *pb.AcceptInvitationFromTutorRequest) (*pb.Empty, error) {
	tutorId, err := uuid.Parse(req.TutorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := h.service.AcceptInvitationFromTutor(ctx, tutorId); err != nil {
		return nil, mapError(err, errdefs.ErrPermissionDenied, errdefs.ErrNotFound)
	}

	return &pb.Empty{}, nil
}

func toPbUser(user *model.User) *pb.User {
	userPb := pb.User{
		Id:           user.Id.String(),
		Role:         user.Role.String(),
		AuthProvider: user.AuthProvider.String(),
		Status:       user.Status.String(),
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Timezone:     user.Timezone,
		CreatedAt:    timestamppb.New(user.CreatedAt),
		EditedAt:     timestamppb.New(user.EditedAt),
	}

	return &userPb
}

func toPbTutorProfile(profile *model.TutorProfile) *pb.TutorProfile {
	return &pb.TutorProfile{
		Id:                   profile.Id.String(),
		UserId:               profile.UserId.String(),
		PaymentInfo:          profile.PaymentInfo,
		LessonPriceRub:       profile.LessonPriceRub,
		LessonConnectionLink: profile.LessonConnectionLink,
		CreatedAt:            timestamppb.New(profile.CreatedAt),
		EditedAt:             timestamppb.New(profile.EditedAt),
	}
}

func toPbTutorStudent(userStudent *model.TutorStudent) *pb.TutorStudent {
	return &pb.TutorStudent{
		Id:                   userStudent.Id.String(),
		TutorId:              userStudent.TutorId.String(),
		StudentId:            userStudent.StudentId.String(),
		Status:               userStudent.Status.String(),
		LessonPriceRub:       userStudent.LessonPriceRub,
		LessonConnectionLink: userStudent.LessonConnectionLink,
		CreatedAt:            timestamppb.New(userStudent.CreatedAt),
		EditedAt:             timestamppb.New(userStudent.EditedAt),
	}
}

func mapError(err error, possibleErrors ...error) error {
	switch {
	case err == nil:
		return nil

	case errors.Is(err, errdefs.ErrAlreadyExists) && slices.Contains(possibleErrors, errdefs.ErrAlreadyExists):
		return status.Errorf(codes.AlreadyExists, err.Error())

	case errors.Is(err, errdefs.ValidationErr) && slices.Contains(possibleErrors, errdefs.ValidationErr):
		return status.Errorf(codes.InvalidArgument, err.Error())

	case errors.Is(err, errdefs.AuthenticationErr) && slices.Contains(possibleErrors, errdefs.AuthenticationErr):
		return status.Errorf(codes.Unauthenticated, err.Error())

	case errors.Is(err, errdefs.ErrNotFound) && slices.Contains(possibleErrors, errdefs.ErrNotFound):
		return status.Errorf(codes.NotFound, err.Error())

	case errors.Is(err, errdefs.ErrPermissionDenied) && slices.Contains(possibleErrors, errdefs.ErrPermissionDenied):
		return status.Errorf(codes.PermissionDenied, err.Error())

	default:
		return status.Errorf(codes.Internal, err.Error())
	}
}
