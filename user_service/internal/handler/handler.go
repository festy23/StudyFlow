package handler

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"userservice/internal/errdefs"
	"userservice/internal/model"
	pb "userservice/pkg/api"
)

type UserService interface {
	RegisterViaTelegram(ctx context.Context, input model.RegisterViaTelegramInput) (model.User, error)
	Authorize(ctx context.Context, input model.AuthorizeInput) (model.User, error)
	GetMe(ctx context.Context) (model.User, error)
	GetUserPublic(ctx context.Context, id uuid.UUID) (model.UserPublic, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input model.User) (model.User, error)
	GetTutorProfile(ctx context.Context, userId uuid.UUID) (model.TutorProfile, error)
	UpdateTutorProfile(ctx context.Context, userId uuid.UUID, input model.TutorProfile) (model.TutorProfile, error)
	CreateTutorStudent(ctx context.Context, input model.CreateTutorStudentInput) (model.TutorStudent, error)
	GetTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) (model.TutorStudent, error)
	UpdateTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID, input model.UpdateTutorStudentInput) (model.TutorStudent, error)
	DeleteTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) error
	ListTutorStudents(ctx context.Context, tutorId uuid.UUID) ([]model.TutorStudent, error)
	ListTutorStudentsForStudent(ctx context.Context, studentId uuid.UUID) ([]model.TutorStudent, error)
	ResolveTutorStudentContext(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) (model.TutorStudentContext, error)
}

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	service UserService
}

func NewUserServiceServer(userService UserService) *UserServiceServer {
	return &UserServiceServer{service: userService}
}

func (h *UserServiceServer) RegisterViaTelegram(ctx context.Context, req *pb.RegisterViaTelegramRequest) (*pb.User, error) {
	input := model.RegisterViaTelegramInput{
		TelegramId: req.GetTelegramId(),
		Username:   req.GetUsername(),
		FirstName:  req.GetFirstName(),
		LastName:   req.GetLastName(),
		Timezone:   req.GetTimezone(),
	}
	user, err := h.service.RegisterViaTelegram(ctx, input)
	if err != nil {
		if errors.Is(err, errdefs.ErrUserAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "User already exists")
		}

		return nil, status.Errorf(codes.Internal, "Internal Error")
	}

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

	return &userPb, nil
}
