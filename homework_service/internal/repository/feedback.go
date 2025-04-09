package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
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
	if err := feedback.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO feedbacks (id, submission_id, file_id, comment, created_at, edited_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	id, err := uuid.NewRandom()
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

	feedback.ID = id.String()
	return nil
}

func (r *FeedbackRepository) Update(ctx context.Context, feedback *domain.Feedback) error {
	if feedback.ID == "" {
		return errors.New("feedback id is required")
	}

	query := `
		UPDATE feedbacks 
		SET file_id = $1, comment = $2, edited_at = $3
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query,
		feedback.FileID,
		feedback.Comment,
		time.Now(),
		feedback.ID,
	)

	return err
}

func (r *FeedbackRepository) GetByID(ctx context.Context, id string) (*domain.Feedback, error) {
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

func (r *FeedbackRepository) ListByFilter(ctx context.Context, filter domain.FeedbackFilter) ([]*domain.Feedback, error) {
	baseQuery := `
		SELECT f.id, f.submission_id, f.file_id, f.comment, f.created_at, f.edited_at
		FROM feedbacks f
		JOIN submissions s ON f.submission_id = s.id
		JOIN assignments a ON s.assignment_id = a.id
		WHERE 1=1
	`

	var args []interface{}
	var conditions []string

	if filter.SubmissionID != "" {
		conditions = append(conditions, "f.submission_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, filter.SubmissionID)
	}

	if filter.AssignmentID != "" {
		conditions = append(conditions, "s.assignment_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, filter.AssignmentID)
	}

	if filter.TutorID != "" {
		conditions = append(conditions, "a.tutor_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, filter.TutorID)
	}

	if filter.StudentID != "" {
		conditions = append(conditions, "a.student_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, filter.StudentID)
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
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
