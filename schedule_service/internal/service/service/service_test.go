package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "schedule_service/pkg/api"
	mock_schedule "schedule_service/pkg/mocks"
)

func TestGetSlot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_schedule.NewMockScheduleServiceClient(ctrl)

	expectedReq := &pb.GetSlotRequest{Id: "slot123"}
	expectedResp := &pb.Slot{
		Id:       "slot123",
		TutorId:  "tutor456",
		StartsAt: timestamppb.New(time.Now()),
		EndsAt:   timestamppb.New(time.Now().Add(time.Hour)),
	}

	mockClient.EXPECT().
		GetSlot(gomock.Any(), expectedReq).
		Return(expectedResp, nil)

	resp, err := mockClient.GetSlot(context.Background(), expectedReq)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
}
func TestCreateSlot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_schedule.NewMockScheduleServiceClient(ctrl)

	now := time.Now()
	req := &pb.CreateSlotRequest{
		TutorId:  "tutor1",
		StartsAt: timestamppb.New(now),
		EndsAt:   timestamppb.New(now.Add(time.Hour)),
	}

	expectedSlot := &pb.Slot{
		Id:        "new_slot",
		TutorId:   req.TutorId,
		StartsAt:  req.StartsAt,
		EndsAt:    req.EndsAt,
		IsBooked:  false,
		CreatedAt: timestamppb.Now(),
	}

	mockClient.EXPECT().
		CreateSlot(gomock.Any(), req).
		Return(expectedSlot, nil)

	resp, err := mockClient.CreateSlot(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedSlot, resp)
	assert.False(t, resp.IsBooked)
}

func TestCreateLesson(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_schedule.NewMockScheduleServiceClient(ctrl)

	req := &pb.CreateLessonRequest{
		SlotId:    "slot12",
		StudentId: "student123",
	}
	expected := &pb.Lesson{
		Id:        "new_lesson",
		SlotId:    "slot12",
		StudentId: "student123",
		Status:    "booked",
		IsPaid:    false,
	}
	mockClient.EXPECT().
		CreateLesson(gomock.Any(), req).Return(expected, nil)
	resp, err := mockClient.CreateLesson(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, expected, resp)
	assert.False(t, resp.IsPaid)

}

func TestCancelLesson(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_schedule.NewMockScheduleServiceClient(ctrl)

	req := &pb.CancelLessonRequest{Id: "lesson123"}

	cancelledLesson := &pb.Lesson{
		Id:     req.Id,
		Status: "cancelled",
	}

	mockClient.EXPECT().
		CancelLesson(gomock.Any(), req).
		Return(cancelledLesson, nil)

	resp, err := mockClient.CancelLesson(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "cancelled", resp.Status)
}

func TestListSlotsByTutor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_schedule.NewMockScheduleServiceClient(ctrl)

	tutorID := "tutor123"
	req := &pb.ListSlotsByTutorRequest{TutorId: tutorID}

	expectedSlots := []*pb.Slot{
		{Id: "slot1", TutorId: tutorID},
		{Id: "slot2", TutorId: tutorID},
	}

	mockClient.EXPECT().
		ListSlotsByTutor(gomock.Any(), req).
		Return(&pb.ListSlotsResponse{Slots: expectedSlots}, nil)

	resp, err := mockClient.ListSlotsByTutor(context.Background(), req)

	assert.NoError(t, err)
	assert.Len(t, resp.Slots, 2)
	for _, slot := range resp.Slots {
		assert.Equal(t, tutorID, slot.TutorId)
	}
}

func TestListLessonsByPairWithFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_schedule.NewMockScheduleServiceClient(ctrl)

	req := &pb.ListLessonsByPairRequest{
		TutorId:      "tutor1",
		StudentId:    "student1",
		StatusFilter: []pb.LessonStatusFilter{pb.LessonStatusFilter_BOOKED},
	}

	expectedLessons := []*pb.Lesson{
		{Id: "lesson1", Status: "booked"},
		{Id: "lesson2", Status: "booked"},
	}

	mockClient.EXPECT().
		ListLessonsByPair(gomock.Any(), req).
		Return(&pb.ListLessonsResponse{Lessons: expectedLessons}, nil)

	resp, err := mockClient.ListLessonsByPair(context.Background(), req)

	assert.NoError(t, err)
	assert.Len(t, resp.Lessons, 2)
	for _, lesson := range resp.Lessons {
		assert.Equal(t, "booked", lesson.Status)
	}
}
func TestGetSlotNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_schedule.NewMockScheduleServiceClient(ctrl)

	req := &pb.GetSlotRequest{Id: "nonexistent"}
	expectedErr := status.Error(codes.NotFound, "slot not found")

	mockClient.EXPECT().
		GetSlot(gomock.Any(), req).
		Return(nil, expectedErr)

	_, err := mockClient.GetSlot(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}
