package data

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jackc/pgx/v5"
	errdefs "paymentservice/internal/errors"
	"paymentservice/internal/models"
)

type AnyTime struct{}

func (a AnyTime) Match(v interface{}) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestPaymentRepo_CreateReceipt(t *testing.T) {
	// arrange
	mockPool, err := pgxmock.NewPool() // Используем пул вместо Conn
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewPaymentRepository(mockPool)
	ctx := context.Background()
	now := time.Now()
	id := uuid.New()
	lessonID := uuid.New()
	fileID := uuid.New()

	mockPool.ExpectQuery("INSERT INTO receipts").
		WithArgs(id, lessonID, fileID, true, AnyTime{}, AnyTime{}).
		WillReturnRows(pgxmock.NewRows([]string{"id", "lesson_id", "file_id", "is_verified", "created_at", "edited_at"}).
			AddRow(id, lessonID, fileID, true, now, now))

	input := &models.PaymentReceiptCreateInput{
		ID:         id,
		LessonID:   lessonID,
		FileID:     fileID,
		IsVerified: true,
	}

	// act
	res, err := repo.CreateReceipt(ctx, input)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, id, res.ID)
}

func TestPaymentRepo_GetReceiptByID_NotFound(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewPaymentRepository(mockPool)
	ctx := context.Background()
	id := uuid.New()

	mockPool.ExpectQuery("SELECT .* FROM receipts WHERE id =").
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetReceiptByID(ctx, id)
	assert.ErrorIs(t, err, errdefs.ErrNotFound)
}

func TestPaymentRepo_ExistsByID(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewPaymentRepository(mockPool)
	ctx := context.Background()
	id := uuid.New()

	mockPool.ExpectQuery("SELECT EXISTS").
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.ExistsByID(ctx, id)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestPaymentRepo_GetReceiptByLessonID_NotFound(t *testing.T) {
	mockPool, err := pgxmock.NewPool() // Используем пул вместо Conn
	require.NoError(t, err)
	defer mockPool.Close()

	repo := NewPaymentRepository(mockPool)
	ctx := context.Background()
	lessonID := uuid.New()

	mockPool.ExpectQuery("SELECT .* FROM receipts WHERE lesson_id =").
		WithArgs(lessonID).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetReceiptByLessonID(ctx, lessonID)
	assert.ErrorIs(t, err, errdefs.ErrNotFound)
}
