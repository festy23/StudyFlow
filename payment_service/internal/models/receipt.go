package models

import (
	"time"

	"github.com/google/uuid"
)

type PaymentReceipt struct {
	ID         uuid.UUID
	LessonID   uuid.UUID
	FileID     uuid.UUID
	IsVerified bool
	CreatedAt  time.Time
	EditedAt   time.Time
}

type PaymentReceiptCreateInput struct {
	ID         uuid.UUID
	LessonID   uuid.UUID
	FileID     uuid.UUID
	IsVerified bool
}

type PaymentReceiptUpdateInput struct {
	ID         uuid.UUID
	IsVerified bool
}

type ReceiptFileUrl struct {
	URL string
}
