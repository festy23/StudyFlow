package domain

import (
	"github.com/google/uuid"
	"time"
)

type Assignment struct {
	ID          uuid.UUID
	TutorID     uuid.UUID
	StudentID   uuid.UUID
	Title       *string
	Description *string
	FileID      *uuid.UUID
	DueDate     *time.Time
	CreatedAt   time.Time
	EditedAt    time.Time
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
	TutorID   uuid.UUID
	StudentID uuid.UUID
	Statuses  []AssignmentStatus
}
