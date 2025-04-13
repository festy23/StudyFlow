package domain

import (
	"errors"
	"time"
)

type Feedback struct {
	ID           string
	SubmissionID string
	FileID       *string
	Comment      string
	CreatedAt    time.Time
	EditedAt     time.Time
}

type FeedbackFilter struct {
	SubmissionID string
	AssignmentID string
	TutorID      string
	StudentID    string
}

func (f *Feedback) Validate() error {
	if f.SubmissionID == "" {
		return errors.New("submission_id is required")
	}
	return nil
}
