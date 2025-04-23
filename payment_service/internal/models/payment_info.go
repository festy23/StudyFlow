package models

import "github.com/google/uuid"

type PaymentInfo struct {
	LessonID       uuid.UUID
	PriceRUB       int32
	PaymentDetails string
}
