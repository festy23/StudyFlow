package postgres

import (
	"context"
	"database/sql"
	"time"

	"payment_service/internal/domain"
)

type paymentRepo struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *paymentRepo {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) GetPaymentInfo(ctx context.Context, lessonID string) (*domain.PaymentInfo, error) {
	// Реализация запроса к БД
	return nil, nil
}

func (r *paymentRepo) CreateReceipt(ctx context.Context, receipt *domain.Receipt) error {
	query := `INSERT INTO receipts (id, lesson_id, file_id, is_verified, created_at, edited_at) 
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query,
		receipt.ID,
		receipt.LessonID,
		receipt.FileID,
		receipt.IsVerified,
		receipt.CreatedAt,
		receipt.EditedAt)
	return err
}

func (r *paymentRepo) GetReceipt(ctx context.Context, receiptID string) (*domain.Receipt, error) {
	query := `SELECT id, lesson_id, file_id, is_verified, created_at, edited_at 
              FROM receipts WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, receiptID)

	var receipt domain.Receipt
	err := row.Scan(
		&receipt.ID,
		&receipt.LessonID,
		&receipt.FileID,
		&receipt.IsVerified,
		&receipt.CreatedAt,
		&receipt.EditedAt)
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (r *paymentRepo) UpdateReceiptVerification(ctx context.Context, receiptID string, isVerified bool) error {
	query := `UPDATE receipts SET is_verified = $1, edited_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, isVerified, time.Now(), receiptID)
	return err
}
