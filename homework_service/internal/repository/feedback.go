package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"homework_service/internal/domain"
)

var ErrNotFound = errors.New("not found")

type FeedbackRepository struct {
	db *sql.DB
}

func NewFeedbackRepository(db *sql.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

func (r *FeedbackRepository) Create(ctx context.Context, feedback *domain.Feedback) error {
	query := `
		INSERT INTO feedbacks (id, submission_id, file_id, comment, created_at, edited_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query,
		id,
		feedback.SubmissionID,
		feedback.FileID,
		feedback.Comment,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	feedback.ID = id
	return nil
}

func (r *FeedbackRepository) Update(ctx context.Context, feedback *domain.Feedback) error {
	query := `
		UPDATE feedbacks 
		SET file_id = $1, comment = $2, edited_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		feedback.FileID,
		feedback.Comment,
		time.Now(),
		feedback.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return err
}

func (r *FeedbackRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Feedback, error) {
	query := `
		SELECT id, submission_id, file_id, comment, created_at, edited_at
		FROM feedbacks
		WHERE id = $1
	`

	var feedback domain.Feedback
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&feedback.ID,
		&feedback.SubmissionID,
		&feedback.FileID,
		&feedback.Comment,
		&feedback.CreatedAt,
		&feedback.EditedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &feedback, nil
}

func (r *FeedbackRepository) ListByAssignment(ctx context.Context, assignmentId uuid.UUID) ([]*domain.Feedback, error) {
	baseQuery := `
		SELECT f.id, f.submission_id, f.file_id, f.comment, f.created_at, f.edited_at
		FROM feedbacks f
		JOIN submissions s
		ON s.id = f.submission_id
		WHERE s.assignment_id = $1
	`

	rows, err := r.db.QueryContext(ctx, baseQuery, assignmentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []*domain.Feedback
	for rows.Next() {
		var feedback domain.Feedback
		err := rows.Scan(
			&feedback.ID,
			&feedback.SubmissionID,
			&feedback.FileID,
			&feedback.Comment,
			&feedback.CreatedAt,
			&feedback.EditedAt,
		)
		if err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, &feedback)
	}

	return feedbacks, nil
}
