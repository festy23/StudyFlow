package data

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/pgxpool"

	errdefs "paymentservice/internal/errors"
	"paymentservice/internal/models"
)

// Querier defines pgxpool.Pool + pgxscan-compatible interface.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// PaymentRepo stores receipts.
type PaymentRepo struct {
	db Querier
}

// NewPaymentRepository creates a new PaymentRepo.
func NewPaymentRepository(db Querier) *PaymentRepo {
	return &PaymentRepo{db: db}
}

// CreateReceipt inserts a new receipt and returns it.
func (r *PaymentRepo) CreateReceipt(ctx context.Context, input *models.PaymentReceiptCreateInput) (*models.PaymentReceipt, error) {
	query := `
		INSERT INTO receipts (id, lesson_id, file_id, is_verified, created_at, edited_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, lesson_id, file_id, is_verified, created_at, edited_at
	`
	now := time.Now()
	pr := &models.PaymentReceipt{}
	err := pgxscan.Get(ctx, r.db, pr, query,
		input.ID,
		input.LessonID,
		input.FileID,
		input.IsVerified,
		now,
		now,
	)
	if err != nil {
		return nil, handleError(err)
	}
	return pr, nil
}

// GetReceiptByID retrieves a receipt by ID.
func (r *PaymentRepo) GetReceiptByID(ctx context.Context, id uuid.UUID) (*models.PaymentReceipt, error) {
	query := `SELECT id, lesson_id, file_id, is_verified, created_at, edited_at FROM receipts WHERE id = $1`
	pr := &models.PaymentReceipt{}
	err := pgxscan.Get(ctx, r.db, pr, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.ErrNotFound
		}
		return nil, handleError(err)
	}
	return pr, nil
}

// UpdateReceipt updates verification and returns receipt.
func (r *PaymentRepo) UpdateReceipt(ctx context.Context, id uuid.UUID, isVerified bool) (*models.PaymentReceipt, error) {
	query := `UPDATE receipts SET is_verified = $1, edited_at = $2 WHERE id = $3`
	now := time.Now()
	_, err := r.db.Exec(ctx, query, isVerified, now, id)
	if err != nil {
		return nil, handleError(err)
	}
	return r.GetReceiptByID(ctx, id)
}

// ExistsByID checks existence by ID.
func (r *PaymentRepo) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM receipts WHERE id = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, handleError(err)
	}
	return exists, nil
}

// GetReceiptByLessonID retrieves a receipt by lesson ID.
func (r *PaymentRepo) GetReceiptByLessonID(ctx context.Context, lessonID uuid.UUID) (*models.PaymentReceipt, error) {
	query := `SELECT id, lesson_id, file_id, is_verified, created_at, edited_at FROM receipts WHERE lesson_id = $1`
	pr := &models.PaymentReceipt{}
	err := pgxscan.Get(ctx, r.db, pr, query, lessonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errdefs.ErrNotFound
		}
		return nil, handleError(err)
	}
	return pr, nil
}
