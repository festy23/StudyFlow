package domain

import (
	"github.com/google/uuid"
	"time"
)

type Feedback struct {
	ID           uuid.UUID
	SubmissionID uuid.UUID
	FileID       *uuid.UUID
	Comment      *string
	CreatedAt    time.Time
	EditedAt     time.Time
}
