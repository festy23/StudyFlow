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
	db repo.Repository
}

func NewScheduleServer(db repo.Repository) *ScheduleServer {
	return &ScheduleServer{
		db: db,
	}
}

func (s *ScheduleServer) GetSlot(ctx context.Context, req *pb.GetSlotRequest) (*pb.Slot, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	slot, err := s.db.GetSlot(ctx, req.Id)
	if err != nil {
		if errors.Is(err, ErrSlotNotFound) {
			return nil, status.Error(codes.NotFound, "slot not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	if slot.TutorID != userID {
		isValidPair, err := ValidateTutorStudentPair(ctx, slot.TutorID, userID)
		if err != nil || !isValidPair {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
	}
	Pbslot := &pb.Slot{
		Id:        slot.ID,
		TutorId:   slot.TutorID,
		StartsAt:  timestamppb.New(slot.StartsAt),
		EndsAt:    timestamppb.New(slot.EndsAt),
		IsBooked:  slot.IsBooked,
		CreatedAt: timestamppb.New(slot.CreatedAt),
		EditedAt:  timestamppb.New(*slot.EditedAt),
	}
	if Pbslot.EditedAt != nil {
		Pbslot.EditedAt = timestamppb.New(*slot.EditedAt)
	} else {
		Pbslot.EditedAt = Pbslot.CreatedAt
	}

	return Pbslot, nil

}

func (s *ScheduleServer) CreateSlot(ctx context.Context, req *pb.CreateSlotRequest) (*pb.Slot, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	isTutor, err := IsTutor(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to verify tutor status")
	}
	if !isTutor {
		return nil, status.Error(codes.PermissionDenied, "only tutors can create slots")
	}

	if req.TutorId != userID {
		return nil, status.Error(codes.PermissionDenied, "cannot create slots for another tutor")
	}

	startsAt := req.StartsAt.AsTime()
	endsAt := req.EndsAt.AsTime()

	if !validateTimeRange(startsAt, endsAt) {
		return nil, status.Error(codes.InvalidArgument, "invalid time range")
	}

	if time.Now().After(startsAt) {
		return nil, status.Error(codes.InvalidArgument, "slot must be scheduled in the future")
	}

	slotID := uuid.New().String()
	now := time.Now()

	slot := repo.Slot{
		ID:        slotID,
		TutorID:   req.TutorId,
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		IsBooked:  false,
		CreatedAt: now,
	}

	if err := s.db.CreateSlot(ctx, slot); err != nil {
		return nil, status.Error(codes.Internal, "failed to create slot")
	}

	return &pb.Slot{
		Id:        slotID,
		TutorId:   req.TutorId,
		StartsAt:  timestamppb.New(startsAt),
		EndsAt:    timestamppb.New(endsAt),
		IsBooked:  false,
		CreatedAt: timestamppb.New(now),
	}, nil
}

func (s *ScheduleServer) UpdateSlot(ctx context.Context, req *pb.UpdateSlotRequest) (*pb.Slot, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	existingSlot, err := s.db.GetSlot(ctx, req.Id)
	if err != nil {
		if errors.Is(err, ErrSlotNotFound) {
			return nil, status.Error(codes.NotFound, "slot not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	if existingSlot.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	if existingSlot.IsBooked {
		return nil, status.Error(codes.FailedPrecondition, "cannot update a booked slot")
	}

	startsAt := req.StartsAt.AsTime()
	endsAt := req.EndsAt.AsTime()

	if !validateTimeRange(startsAt, endsAt) {
		return nil, status.Error(codes.InvalidArgument, "invalid time range")
	}

	if time.Now().After(startsAt) {
		return nil, status.Error(codes.InvalidArgument, "slot must be scheduled in the future")
	}

	now := time.Now()
	existingSlot.StartsAt = startsAt
	existingSlot.EndsAt = endsAt
	existingSlot.EditedAt = &now

	if err := s.db.UpdateSlot(ctx, *existingSlot); err != nil {
		return nil, status.Error(codes.Internal, "failed to update slot")
	}

	return &pb.Slot{
		Id:        existingSlot.ID,
		TutorId:   existingSlot.TutorID,
		StartsAt:  timestamppb.New(existingSlot.StartsAt),
		EndsAt:    timestamppb.New(existingSlot.EndsAt),
		IsBooked:  existingSlot.IsBooked,
		CreatedAt: timestamppb.New(existingSlot.CreatedAt),
		EditedAt:  timestamppb.New(*existingSlot.EditedAt),
	}, nil
}

func (s *ScheduleServer) DeleteSlot(ctx context.Context, req *pb.DeleteSlotRequest) (*pb.Empty, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	existingSlot, err := s.db.GetSlot(ctx, req.Id)
	if err != nil {
		if errors.Is(err, ErrSlotNotFound) {
			return nil, status.Error(codes.NotFound, "slot not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	if existingSlot.TutorID != userID {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	if existingSlot.IsBooked {
		return nil, status.Error(codes.FailedPrecondition, "cannot delete a booked slot")
	}

	if err := s.db.DeleteSlot(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete slot")
	}

	return &pb.Empty{}, nil
}

func (s *ScheduleServer) ListSlotsByTutor(ctx context.Context, req *pb.ListSlotsByTutorRequest) (*pb.ListSlotsResponse, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if req.TutorId != userID {
		isValidPair, err := ValidateTutorStudentPair(ctx, req.TutorId, userID)
		if err != nil || !isValidPair {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
	}

	var onlyAvailable bool
	if req.OnlyAvailable != nil {
		onlyAvailable = *req.OnlyAvailable
	}

	slots, err := s.db.ListSlotsByTutor(ctx, req.TutorId, onlyAvailable)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list slots")
	}

	protoSlots := make([]*pb.Slot, 0, len(slots))
	for _, slot := range slots {
		protoSlot := &pb.Slot{
			Id:        slot.ID,
			TutorId:   slot.TutorID,
			StartsAt:  timestamppb.New(slot.StartsAt),
			EndsAt:    timestamppb.New(slot.EndsAt),
			IsBooked:  slot.IsBooked,
			CreatedAt: timestamppb.New(slot.CreatedAt),
		}

		if slot.EditedAt != nil {
			protoSlot.EditedAt = timestamppb.New(*slot.EditedAt)
		}

		protoSlots = append(protoSlots, protoSlot)
	}

	return &pb.ListSlotsResponse{
		Slots: protoSlots,
	}, nil
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

	isValidPair, err := ValidateTutorStudentPair(ctx, tutorID, studentID)
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

	isValidPair, err := ValidateTutorStudentPair(ctx, req.TutorId, req.StudentId)
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
