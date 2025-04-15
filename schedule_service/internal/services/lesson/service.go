package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"common_library/ctxdata"
	"schedule_service/internal/database/repo"
	pb "schedule_service/pkg/api"
)

type ScheduleServer struct {
	pb.UnimplementedScheduleServiceServer
	db          repo.Repository
	userService UserService
}

type UserService interface {
	ValidateTutorStudentPair(ctx context.Context, tutorID, studentID string) (bool, error)
	IsTutor(ctx context.Context, userID string) (bool, error)
}

func NewScheduleServer(db repo.Repository, userService UserService) *ScheduleServer {
	return &ScheduleServer{
		db:          db,
		userService: userService,
	}
}

func (s *ScheduleServer) GetLesson(ctx context.Context, req *pb.GetLessonRequest) (*pb.Lesson, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, StatusUnauthenticated
	}

	lesson, err := s.db.GetLesson(ctx, req.Id)
	if err != nil {
		if errors.Is(err, ErrLessonNotFound) {
			return nil, StatusNotFound
		}
		return nil, StatusInternalError
	}

	slot, err := s.db.GetSlot(ctx, lesson.SlotID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get slot information")
	}

	if userID != slot.TutorID && userID != lesson.StudentID {
		return nil, StatusPermissionDenied
	}

	return convertrepoLessonToProto(lesson), nil
}

func (s *ScheduleServer) CreateLesson(ctx context.Context, req *pb.CreateLessonRequest) (*pb.Lesson, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, StatusUnauthenticated
	}

	slot, err := s.db.GetSlot(ctx, req.SlotId)
	if err != nil {
		if errors.Is(err, ErrSlotNotFound) {
			return nil, status.Error(codes.NotFound, "slot not found")
		}
		return nil, StatusInternalError
	}

	if slot.IsBooked {
		return nil, status.Error(codes.AlreadyExists, "slot is already booked")
	}

	var tutorID, studentID string

	if userID == slot.TutorID {
		tutorID = userID
		studentID = req.StudentId
	} else if userID == req.StudentId {
		tutorID = slot.TutorID
		studentID = userID
	} else {
		return nil, StatusPermissionDenied
	}

	isValidPair, err := s.userService.ValidateTutorStudentPair(ctx, tutorID, studentID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to validate tutor-student relationship")
	}
	if !isValidPair {
		return nil, status.Error(codes.FailedPrecondition, "tutor and student are not connected")
	}

	lessonID := uuid.New().String()
	now := time.Now()

	lesson := repo.Lesson{
		ID:        lessonID,
		SlotID:    req.SlotId,
		StudentID: studentID,
		Status:    "booked",
		IsPaid:    false,
		CreatedAt: now,
		EditedAt:  now,
	}

	if err := s.db.CreateLessonAndBookSlot(ctx, lesson, req.SlotId); err != nil {
		return nil, status.Error(codes.Internal, "failed to create lesson")
	}

	return &pb.Lesson{
		Id:        lessonID,
		SlotId:    req.SlotId,
		StudentId: studentID,
		Status:    "booked",
		IsPaid:    false,
		CreatedAt: timestamppb.New(now),
		EditedAt:  timestamppb.New(now),
	}, nil
}

func (s *ScheduleServer) UpdateLesson(ctx context.Context, req *pb.UpdateLessonRequest) (*pb.Lesson, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, StatusUnauthenticated
	}

	lesson, err := s.db.GetLesson(ctx, req.Id)
	if err != nil {
		if errors.Is(err, ErrLessonNotFound) {
			return nil, StatusNotFound
		}
		return nil, StatusInternalError
	}

	slot, err := s.db.GetSlot(ctx, lesson.SlotID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get slot information")
	}

	if userID != slot.TutorID {
		return nil, status.Error(codes.PermissionDenied, "only tutors can update lesson details")
	}

	now := time.Now()
	isUpdated := false

	if req.ConnectionLink != nil {
		lesson.ConnectionLink = req.ConnectionLink
		isUpdated = true
	}

	if req.PriceRub != nil {
		lesson.PriceRub = req.PriceRub
		isUpdated = true
	}

	if req.PaymentInfo != nil {
		lesson.PaymentInfo = req.PaymentInfo
		isUpdated = true
	}

	if isUpdated {
		lesson.EditedAt = now
		if err := s.db.UpdateLesson(ctx, *lesson); err != nil {
			return nil, status.Error(codes.Internal, "failed to update lesson")
		}
	}

	return convertrepoLessonToProto(lesson), nil
}

func (s *ScheduleServer) CancelLesson(ctx context.Context, req *pb.CancelLessonRequest) (*pb.Lesson, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, StatusUnauthenticated
	}

	lesson, err := s.db.GetLesson(ctx, req.Id)
	if err != nil {
		if errors.Is(err, ErrLessonNotFound) {
			return nil, StatusNotFound
		}
		return nil, StatusInternalError
	}

	slot, err := s.db.GetSlot(ctx, lesson.SlotID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get slot information")
	}

	if userID != slot.TutorID && userID != lesson.StudentID {
		return nil, StatusPermissionDenied
	}

	now := time.Now()
	lesson.Status = "cancelled"
	lesson.EditedAt = now

	if err := s.db.CancelLessonAndFreeSlot(ctx, *lesson, lesson.SlotID); err != nil {
		return nil, status.Error(codes.Internal, "failed to cancel lesson")
	}

	return convertrepoLessonToProto(lesson), nil
}

func (s *ScheduleServer) ListLessonsByTutor(ctx context.Context, req *pb.ListLessonsByTutorRequest) (*pb.ListLessonsResponse, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, StatusUnauthenticated
	}

	if req.TutorId != userID {
		return nil, StatusPermissionDenied
	}

	statusFilters := make([]string, 0, len(req.StatusFilter))
	for _, status := range req.StatusFilter {
		switch status {
		case pb.LessonStatusFilter_BOOKED:
			statusFilters = append(statusFilters, "booked")
		case pb.LessonStatusFilter_CANCELLED:
			statusFilters = append(statusFilters, "cancelled")
		case pb.LessonStatusFilter_COMPLETED:
			statusFilters = append(statusFilters, "completed")
		}
	}

	lessons, err := s.db.ListLessonsByTutor(ctx, req.TutorId, statusFilters)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list lessons")
	}

	return createListLessonsResponse(lessons), nil
}

func (s *ScheduleServer) ListLessonsByStudent(ctx context.Context, req *pb.ListLessonsByStudentRequest) (*pb.ListLessonsResponse, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, StatusUnauthenticated
	}

	if req.StudentId != userID {
		return nil, StatusPermissionDenied
	}

	statusFilters := make([]string, 0, len(req.StatusFilter))
	for _, status := range req.StatusFilter {
		switch status {
		case pb.LessonStatusFilter_BOOKED:
			statusFilters = append(statusFilters, "booked")
		case pb.LessonStatusFilter_CANCELLED:
			statusFilters = append(statusFilters, "cancelled")
		case pb.LessonStatusFilter_COMPLETED:
			statusFilters = append(statusFilters, "completed")
		}
	}

	lessons, err := s.db.ListLessonsByStudent(ctx, req.StudentId, statusFilters)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list lessons")
	}

	return createListLessonsResponse(lessons), nil
}

func (s *ScheduleServer) ListLessonsByPair(ctx context.Context, req *pb.ListLessonsByPairRequest) (*pb.ListLessonsResponse, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, StatusUnauthenticated
	}

	if req.TutorId != userID && req.StudentId != userID {
		return nil, StatusPermissionDenied
	}

	isValidPair, err := s.userService.ValidateTutorStudentPair(ctx, req.TutorId, req.StudentId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to validate tutor-student relationship")
	}
	if !isValidPair {
		return nil, status.Error(codes.PermissionDenied, "tutor and student are not connected")
	}

	statusFilters := make([]string, 0, len(req.StatusFilter))
	for _, status := range req.StatusFilter {
		switch status {
		case pb.LessonStatusFilter_BOOKED:
			statusFilters = append(statusFilters, "booked")
		case pb.LessonStatusFilter_CANCELLED:
			statusFilters = append(statusFilters, "cancelled")
		case pb.LessonStatusFilter_COMPLETED:
			statusFilters = append(statusFilters, "completed")
		}
	}

	lessons, err := s.db.ListLessonsByPair(ctx, req.TutorId, req.StudentId, statusFilters)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list lessons")
	}

	return createListLessonsResponse(lessons), nil
}

func (s *ScheduleServer) ListCompletedUnpaidLessons(ctx context.Context, req *pb.ListCompletedUnpaidLessonsRequest) (*pb.ListLessonsResponse, error) {
	var after *time.Time
	if req.After != nil {
		t := req.After.AsTime()
		after = &t
	}

	lessons, err := s.db.ListCompletedUnpaidLessons(ctx, after)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list completed unpaid lessons")
	}

	return createListLessonsResponse(lessons), nil
}
