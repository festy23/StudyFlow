package service

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrSlotBooked       = errors.New("slot is already booked")
	ErrSlotNotFound     = errors.New("slot not found")
	ErrLessonNotFound   = errors.New("lesson not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidTimeRange = errors.New("invalid time range")
	ErrPastTime         = errors.New("time cannot be in the past")
	ErrInvalidPair      = errors.New("tutor and student are not connected")
	ErrNotTutor         = errors.New("user is not a tutor")

	StatusUnauthenticated  = status.Error(codes.Unauthenticated, "user not authenticated")
	StatusPermissionDenied = status.Error(codes.PermissionDenied, "permission denied")
	StatusNotFound         = status.Error(codes.NotFound, "lesson not found")
	StatusInternalError    = status.Error(codes.Internal, "internal error")
)
