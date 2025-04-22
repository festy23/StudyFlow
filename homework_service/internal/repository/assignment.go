package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"homework_service/internal/domain"
	"strings"
	"time"

	"github.com/google/uuid"
)

const statusSubQuery = `
WITH latest_submissions AS (
    SELECT *
    FROM (
        SELECT *,
               ROW_NUMBER() OVER (PARTITION BY assignment_id ORDER BY created_at DESC) AS rn
        FROM submissions
    ) s
    WHERE rn = 1
),
latest_feedbacks AS (
    SELECT *
    FROM (
        SELECT *,
               ROW_NUMBER() OVER (PARTITION BY submission_id ORDER BY created_at DESC) AS rn
        FROM feedbacks
    ) f
    WHERE rn = 1
),
assignment_statuses AS (
    SELECT
        a.id, a.tutor_id, a.student_id, a.title, a.description,
        a.file_id, a.due_date, a.created_at, a.edited_at,
        CASE
            WHEN ls.id IS NULL AND a.due_date > NOW() THEN 'UNSENT'
            WHEN ls.id IS NULL AND a.due_date <= NOW() THEN 'OVERDUE'
            WHEN ls.id IS NOT NULL AND lf.id IS NULL THEN 'UNREVIEWED'
            WHEN lf.id IS NOT NULL THEN 'REVIEWED'
            ELSE 'UNSPECIFIED'
        END AS status
    FROM assignments a
    LEFT JOIN latest_submissions ls ON ls.assignment_id = a.id
    LEFT JOIN latest_feedbacks lf ON lf.submission_id = ls.id
)
`

type AssignmentRepository struct {
	db *sql.DB
}

type AssignmentRepositoryInterface interface {
	Create(ctx context.Context, assignment *domain.Assignment) error
	GetByID(ctx context.Context, id string) (*domain.Assignment, error)
	FindAssignmentsDueSoon(ctx context.Context, duration time.Duration) ([]*domain.Assignment, error)
	Update(ctx context.Context, assignment *domain.Assignment) error
	Delete(ctx context.Context, id string) error
	ListByFilter(ctx context.Context, filter *domain.AssignmentFilter) ([]*domain.Assignment, error)
}

func NewAssignmentRepository(db *sql.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) ListByFilter(ctx context.Context, filter domain.AssignmentFilter) ([]*domain.Assignment, error) {
	query := statusSubQuery + `
SELECT id, tutor_id, student_id, title, description, 
file_id, due_date, created_at, edited_at 
FROM assignment_statuses WHERE 1=1
`
	var args []interface{}
	argsCount := 1

	if filter.TutorID != uuid.Nil {
		query += fmt.Sprintf(" AND tutor_id = $%d", argsCount)
		args = append(args, filter.TutorID)
		argsCount++
	}

	if filter.StudentID != uuid.Nil {
		query += fmt.Sprintf(" AND student_id = $%d", argsCount)
		args = append(args, filter.StudentID)
		argsCount++
	}

	if len(filter.Statuses) > 0 {
		placeholders := make([]string, len(filter.Statuses))
		for i := range filter.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argsCount)
			args = append(args, filter.Statuses[i])
			argsCount++
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ", "))
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*domain.Assignment
	for rows.Next() {
		var a domain.Assignment
		if err := rows.Scan(
			&a.ID,
			&a.TutorID,
			&a.StudentID,
			&a.Title,
			&a.Description,
			&a.FileID,
			&a.DueDate,
			&a.CreatedAt,
			&a.EditedAt,
		); err != nil {
			return nil, err
		}
		assignments = append(assignments, &a)
	}

	return assignments, nil
}

func (r *AssignmentRepository) FindAssignmentsDueSoon(ctx context.Context, duration time.Duration) ([]*domain.Assignment, error) {
	query := statusSubQuery + `
		SELECT id, tutor_id, student_id, title, description, file_id, due_date,
		       created_at, edited_at
		FROM assignment_statuses
		WHERE due_date BETWEEN NOW() AND $1
		AND status NOT IN ('REVIEWED', 'OVERDUE')
	`

	deadline := time.Now().Add(duration)
	rows, err := r.db.QueryContext(ctx, query, deadline)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*domain.Assignment
	for rows.Next() {
		var a domain.Assignment
		err := rows.Scan(
			&a.ID,
			&a.TutorID,
			&a.StudentID,
			&a.Title,
			&a.Description,
			&a.FileID,
			&a.DueDate,
			&a.CreatedAt,
			&a.EditedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, &a)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return assignments, nil
}

func (r *AssignmentRepository) Create(ctx context.Context, assignment *domain.Assignment) error {
	query := `
		INSERT INTO assignments 
			(id, tutor_id, student_id, title, description, file_id, due_date, created_at, edited_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		id,
		assignment.TutorID,
		assignment.StudentID,
		assignment.Title,
		assignment.Description,
		assignment.FileID,
		assignment.DueDate,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	assignment.ID = id
	return nil
}

func (r *AssignmentRepository) Update(ctx context.Context, assignment *domain.Assignment) error {
	query := `
		UPDATE assignments 
		SET title = $1, description = $2, file_id = $3, due_date = $4, edited_at = $5
		WHERE id = $6
	`
	result, err := r.db.ExecContext(ctx, query,
		assignment.Title,
		assignment.Description,
		assignment.FileID,
		assignment.DueDate,
		time.Now(),
		assignment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update assignment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("assignment not found")
	}

	return nil
}

func (r *AssignmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Assignment, error) {
	query := `
		SELECT id, tutor_id, student_id, title, description, file_id, due_date, 
		       created_at, edited_at
		FROM assignments
		WHERE id = $1
	`

	var assignment domain.Assignment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&assignment.ID,
		&assignment.TutorID,
		&assignment.StudentID,
		&assignment.Title,
		&assignment.Description,
		&assignment.FileID,
		&assignment.DueDate,
		&assignment.CreatedAt,
		&assignment.EditedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	return &assignment, nil
}

func (r *AssignmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM assignments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete assignment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
