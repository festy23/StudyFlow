//go:generate mockgen -source=scheduleServiceClient.go -destination=../mocks/schedule_serice_mock.go -package=mocks

package clients

import (
	"context"
	"google.golang.org/grpc"
	api3 "schedule_service/pkg/api"
)

type ScheduleServiceClient interface {
	GetLesson(ctx context.Context, req *api3.GetLessonRequest, opts ...grpc.CallOption) (*api3.Lesson, error)
	CreateLesson(ctx context.Context, req *api3.CreateLessonRequest, opts ...grpc.CallOption) (*api3.Lesson, error)
	UpdateLesson(ctx context.Context, req *api3.UpdateLessonRequest, opts ...grpc.CallOption) (*api3.Lesson, error)
	CancelLesson(ctx context.Context, req *api3.CancelLessonRequest, opts ...grpc.CallOption) (*api3.Lesson, error)
}
