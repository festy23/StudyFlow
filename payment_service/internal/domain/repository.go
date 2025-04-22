package domain

import "context"

type PaymentRepository interface {
	GetPaymentInfo(ctx context.Context, lessonID string) (*PaymentInfo, error)
	CreateReceipt(ctx context.Context, receipt *Receipt) error
	GetReceipt(ctx context.Context, receiptID string) (*Receipt, error)
	UpdateReceiptVerification(ctx context.Context, receiptID string, isVerified bool) error
}
