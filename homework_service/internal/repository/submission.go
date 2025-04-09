package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"homework_service/internal/domain"
)

type SubmissionRepository struct {
	db *sql.DB
}

func NewSubmissionRepository(db *sql.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) Create(ctx context.Context, submission *domain.Submission) error {
	query := `
		INSERT INTO submissions (id, assignment_id, student_id, file_id, comment, created_at, edited_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query,
		id,
		submission.AssignmentID,
		submission.StudentID,
		submission.FileID,
		submission.Comment,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	submission.ID = id.String()
	return nil
}

func (r *SubmissionRepository) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	query := `
		SELECT id, assignment_id, student_id, file_id, comment, created_at, edited_at
		FROM submissions
		WHERE id = $1
	`

	var submission domain.Submission
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&submission.ID,
		&submission.AssignmentID,
		&submission.StudentID,
		&submission.FileID,
		&submission.Comment,
		&submission.CreatedAt,
		&submission.EditedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &submission, nil
}

func (r *SubmissionRepository) ListByFilter(ctx context.Context, filter domain.SubmissionFilter) ([]*domain.Submission, error) {
	query := `
		SELECT id, assignment_id, student_id, file_id, comment, created_at, edited_at
		FROM submissions
		WHERE assignment_id = $1 AND student_id = $2
	`

	rows, err := r.db.QueryContext(ctx, query, filter.AssignmentID, filter.StudentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []*domain.Submission
	for rows.Next() {
		var submission domain.Submission
		err := rows.Scan(
			&submission.ID,
			&submission.AssignmentID,
			&submission.StudentID,
			&submission.FileID,
			&submission.Comment,
			&submission.CreatedAt,
			&submission.EditedAt,
		)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, &submission)
	}

	return submissions, nil
}
