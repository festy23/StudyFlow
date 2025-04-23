package service

import (
	"common_library/ctxdata"
	"context"
	"errors"
	"fmt"
	"schedule_service/internal/database/repo"
	pb "schedule_service/pkg/api"
	"time"

	"google.golang.org/grpc/metadata"

	userpb "userservice/pkg/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserClient struct {
	conn   *grpc.ClientConn
	client userpb.UserServiceClient
}

func NewUserClient(adress string) (*UserClient, error) {
	conn, err := grpc.NewClient(adress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &UserClient{
		conn:   conn,
		client: userpb.NewUserServiceClient(conn),
	}, nil

}
func (c *UserClient) Close() {
	c.conn.Close()
}

func (c *UserClient) GetTutorStudent(ctx context.Context, tutorID, studentID string) (*userpb.TutorStudent, error) {
	return c.client.GetTutorStudent(ctx, &userpb.GetTutorStudentRequest{
		TutorId:   tutorID,
		StudentId: studentID,
	})
}

func convertrepoLessonToProto(lesson *repo.Lesson) *pb.Lesson {
	protoLesson := &pb.Lesson{
		Id:        lesson.ID,
		SlotId:    lesson.SlotID,
		StudentId: lesson.StudentID,
		Status:    lesson.Status,
		IsPaid:    lesson.IsPaid,
		CreatedAt: timestamppb.New(lesson.CreatedAt),
		EditedAt:  timestamppb.New(lesson.EditedAt),
	}

	if lesson.ConnectionLink != nil {
		protoLesson.ConnectionLink = lesson.ConnectionLink
	}

	if lesson.PriceRub != nil {
		protoLesson.PriceRub = lesson.PriceRub
	}

	if lesson.PaymentInfo != nil {
		protoLesson.PaymentInfo = lesson.PaymentInfo
	}

	return protoLesson
}

func createListLessonsResponse(lessons []repo.Lesson) *pb.ListLessonsResponse {
	protoLessons := make([]*pb.Lesson, 0, len(lessons))

	for _, lesson := range lessons {
		lessonCopy := lesson
		protoLesson := convertrepoLessonToProto(&lessonCopy)
		protoLessons = append(protoLessons, protoLesson)
	}

	return &pb.ListLessonsResponse{
		Lessons: protoLessons,
	}
}

func validateTimeRange(start, end time.Time) bool {
	return start.Before(end)
}

func (s *ScheduleServer) ValidateTutorStudentPair(ctx context.Context, tutorID, studentID string) (bool, error) {
	currentUserID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return false, errors.New("user ID not found in context")
	}

	currentUserRole, ok := ctxdata.GetUserRole(ctx)
	if !ok {
		return false, errors.New("user role not found in context")
	}

	reqCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("x-user-id", currentUserID, "x-user-role", currentUserRole))
	tutorStudent, err := s.UserClient.GetTutorStudent(reqCtx, tutorID, studentID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify tutor-student pair: %w", err)
	}
	if tutorStudent.GetStatus() != "active" {
		return false, nil
	}

	switch currentUserRole {
	case "tutor":
		return currentUserID == tutorID, nil
	case "student":
		return currentUserID == studentID, nil
	default:
		return false, nil
	}

}
func IsTutor(ctx context.Context, userID string) (bool, error) {
	currentUserID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return false, errors.New("user ID not found in context")
	}

	if currentUserID != userID {
		return false, nil
	}

	role, ok := ctxdata.GetUserRole(ctx)
	if !ok {
		return false, errors.New("user role not found in context")
	}

	return role == "tutor", nil
}
