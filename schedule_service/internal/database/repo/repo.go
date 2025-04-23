package repo

import (
	"context"
	"time"
)

type Slot struct {
	ID        string
	TutorID   string
	StartsAt  time.Time
	EndsAt    time.Time
	IsBooked  bool
	CreatedAt time.Time
	EditedAt  *time.Time
}

type Lesson struct {
	ID             string
	SlotID         string
	StudentID      string
	Status         string // "booked", "cancelled", "completeбd"
	IsPaid         bool
	ConnectionLink *string
	PriceRub       *int32
	PaymentInfo    *string
	CreatedAt      time.Time
	EditedAt       time.Time
}

type Repository interface {
	// Slot operations
	GetSlot(ctx context.Context, id string) (*Slot, error)
	CreateSlot(ctx context.Context, slot Slot) error
	UpdateSlot(ctx context.Context, slot Slot) error
	DeleteSlot(ctx context.Context, id string) error
	ListSlotsByTutor(ctx context.Context, tutorID string, onlyAvailable bool) ([]Slot, error)

	// Lesson operations
	GetLesson(ctx context.Context, id string) (*Lesson, error)
	CreateLessonAndBookSlot(ctx context.Context, lesson Lesson, slotID string) error
	UpdateLesson(ctx context.Context, lesson Lesson) error
	CancelLessonAndFreeSlot(ctx context.Context, lesson Lesson, slotID string) error
	ListLessonsByTutor(ctx context.Context, tutorID string, statusFilter []string) ([]Lesson, error)
	ListLessonsByStudent(ctx context.Context, studentID string, statusFilter []string) ([]Lesson, error)
	ListLessonsByPair(ctx context.Context, tutorID, studentID string, statusFilter []string) ([]Lesson, error)
	ListCompletedUnpaidLessons(ctx context.Context, after *time.Time) ([]Lesson, error)

	UpdateCompletedLessons(ctx context.Context) (int, error)

	MarkAsPaid(ctx context.Context, lessonID string) error
}
