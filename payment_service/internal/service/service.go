package service

import (
	"context"

	"payment_service/internal/domain"
)

type ScheduleServiceClient interface {
	GetLesson(ctx context.Context, lessonID string) (*domain.Lesson, error)
	ListCompletedUnpaidLessons(ctx context.Context) ([]*domain.Lesson, error)
}

type UserServiceClient interface {
	ResolveTutorStudentContext(ctx context.Context, tutorID, studentID string) (*domain.TutorStudentContext, error)
}

type FileServiceClient interface {
	GetTemporaryURL(ctx context.Context, fileID string) (string, error)
}

type NotificationService interface {
	SendPaymentNotification(ctx context.Context, lessonID, userID string) error
	SendReminderNotification(ctx context.Context, lessonID, userID string) error
}
