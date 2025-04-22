package domain

import "time"

type PaymentInfo struct {
	LessonID    string
	PriceRub    int32
	PaymentInfo string
}

type Receipt struct {
	ID         string
	LessonID   string
	FileID     string
	IsVerified bool
	CreatedAt  time.Time
	EditedAt   time.Time
}
