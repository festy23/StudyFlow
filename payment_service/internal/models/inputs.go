package models

import "github.com/google/uuid"

type GetPaymentInfoInput struct {
	LessonId uuid.UUID
}

type SubmitPaymentReceiptInput struct {
	LessonId uuid.UUID
	FileId   uuid.UUID
}

type GetReceiptInput struct {
	ReceiptId uuid.UUID
}

type VerifyReceiptInput struct {
	ReceiptId uuid.UUID
}

type GetReceiptFileInput struct {
	ReceiptId uuid.UUID
}
