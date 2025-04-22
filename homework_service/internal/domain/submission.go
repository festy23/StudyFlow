package domain

import (
	"github.com/google/uuid"
	"time"
)

type Submission struct {
	ID           uuid.UUID
	AssignmentID uuid.UUID
	FileID       *uuid.UUID
	Comment      *string
	CreatedAt    time.Time
	EditedAt     time.Time
}
