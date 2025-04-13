package domain

import (
	"context"
	"time"
)

type Assignment struct {
	ID          string
	TutorID     string
	StudentID   string
	Title       string
	Description string
	FileID      string
	DueDate     *time.Time
	CreatedAt   time.Time
	EditedAt    time.Time
	Status      string
}

type AssignmentStatus string

const (
	AssignmentStatusUnspecified AssignmentStatus = "UNSPECIFIED"
	AssignmentStatusUnsent      AssignmentStatus = "UNSENT"
	AssignmentStatusUnreviewed  AssignmentStatus = "UNREVIEWED"
	AssignmentStatusReviewed    AssignmentStatus = "REVIEWED"
	AssignmentStatusOverdue     AssignmentStatus = "OVERDUE"
)

type AssignmentFilter struct {
	TutorID     string
	StudentID   string
	Statuses    []AssignmentStatus
	OnlyActive  bool
	OnlyOverdue bool
}

type AssignmentRepository interface {
	Create(ctx context.Context, assignment *Assignment) error
	GetByID(ctx context.Context, id string) (*Assignment, error)
	ListByTutorID(ctx context.Context, tutorID string) ([]*Assignment, error)
}
