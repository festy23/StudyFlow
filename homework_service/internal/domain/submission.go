package domain

import "time"

type Submission struct {
	ID           string
	AssignmentID string
	StudentID    string
	FileID       string
	Comment      string
	CreatedAt    time.Time
	EditedAt     time.Time
}

type SubmissionStatus string

const (
	SubmissionStatusUnspecified SubmissionStatus = "UNSPECIFIED"
	SubmissionStatusSubmitted   SubmissionStatus = "SUBMITTED"
	SubmissionStatusReviewed    SubmissionStatus = "REVIEWED"
)

type SubmissionFilter struct {
	AssignmentID string
	StudentID    string
	Status       SubmissionStatus
}
