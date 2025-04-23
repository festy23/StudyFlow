//go:generate mockgen -source=userServiceClient.go -destination=../mocks/user_service_mock.go
//.go -package=mocks

package clients

import (
	"context"
	"google.golang.org/grpc"
	api4 "userservice/pkg/api"
)

type UserServiceClient interface {
	RegisterViaTelegram(ctx context.Context, req *api4.RegisterViaTelegramRequest, opts ...grpc.CallOption) (*api4.User, error)

	AuthorizeByAuthHeader(ctx context.Context, req *api4.AuthorizeByAuthHeaderRequest, opts ...grpc.CallOption) (*api4.User, error)

	GetMe(ctx context.Context, req *api4.Empty, opts ...grpc.CallOption) (*api4.User, error)

	GetUser(ctx context.Context, req *api4.GetUserRequest, opts ...grpc.CallOption) (*api4.UserPublic, error)

	UpdateUser(ctx context.Context, req *api4.UpdateUserRequest, opts ...grpc.CallOption) (*api4.User, error)

	UpdateTutorProfile(ctx context.Context, req *api4.UpdateTutorProfileRequest, opts ...grpc.CallOption) (*api4.TutorProfile, error)

	GetTutorProfileByUserId(ctx context.Context, req *api4.GetTutorProfileByUserIdRequest, opts ...grpc.CallOption) (*api4.TutorProfile, error)

	GetTutorStudent(ctx context.Context, req *api4.GetTutorStudentRequest, opts ...grpc.CallOption) (*api4.TutorStudent, error)

	CreateTutorStudent(ctx context.Context, req *api4.CreateTutorStudentRequest, opts ...grpc.CallOption) (*api4.TutorStudent, error)

	UpdateTutorStudent(ctx context.Context, req *api4.UpdateTutorStudentRequest, opts ...grpc.CallOption) (*api4.TutorStudent, error)

	DeleteTutorStudent(ctx context.Context, req *api4.DeleteTutorStudentRequest, opts ...grpc.CallOption) (*api4.Empty, error)

	ListTutorStudents(ctx context.Context, req *api4.ListTutorStudentsRequest, opts ...grpc.CallOption) (*api4.ListTutorStudentsResponse, error)

	ListTutorsForStudent(ctx context.Context, req *api4.ListTutorsForStudentRequest, opts ...grpc.CallOption) (*api4.ListTutorsForStudentResponse, error)

	ResolveTutorStudentContext(ctx context.Context, req *api4.ResolveTutorStudentContextRequest, opts ...grpc.CallOption) (*api4.ResolvedTutorStudentContext, error)

	AcceptInvitationFromTutor(ctx context.Context, req *api4.AcceptInvitationFromTutorRequest, opts ...grpc.CallOption) (*api4.Empty, error)
}
