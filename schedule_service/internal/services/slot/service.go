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
		isValidPair, err := s.userService.ValidateTutorStudentPair(ctx, slot.TutorID, userID)
		if err != nil || !isValidPair {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
	}

	return &pb.Slot{
		Id:        slot.ID,
		TutorId:   slot.TutorID,
		StartsAt:  timestamppb.New(slot.StartsAt),
		EndsAt:    timestamppb.New(slot.EndsAt),
		IsBooked:  slot.IsBooked,
		CreatedAt: timestamppb.New(slot.CreatedAt),
		EditedAt:  timestamppb.New(*slot.EditedAt),
	}, nil
}

func (s *ScheduleServer) CreateSlot(ctx context.Context, req *pb.CreateSlotRequest) (*pb.Slot, error) {
	userID, ok := ctxdata.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	isTutor, err := s.userService.IsTutor(ctx, userID)
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
		isValidPair, err := s.userService.ValidateTutorStudentPair(ctx, req.TutorId, userID)
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
