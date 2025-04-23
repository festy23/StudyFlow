package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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

}
