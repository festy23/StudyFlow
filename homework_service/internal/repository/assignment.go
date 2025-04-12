package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"homework_service/internal/domain"
)

type AssignmentRepository struct {
	db *sql.DB
}

type AssignmentRepositoryInterface interface {
	Create(ctx context.Context, assignment *domain.Assignment) error
	GetByID(ctx context.Context, id string) (*domain.Assignment, error)
	FindAssignmentsDueSoon(ctx context.Context, duration time.Duration) ([]*domain.Assignment, error)
	Update(ctx context.Context, assignment *domain.Assignment) error
	ListByTutorID(ctx context.Context, tutorID string) ([]*domain.Assignment, error)
	ListByStudentID(ctx context.Context, studentID string) ([]*domain.Assignment, error)
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, status string) error
}

func NewAssignmentRepository(db *sql.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) FindAssignmentsDueSoon(ctx context.Context, duration time.Duration) ([]*domain.Assignment, error) {
	query := `
		SELECT id, tutor_id, student_id, title, description, file_id, due_date, status, 
		       created_at, edited_at
		FROM assignments
		WHERE due_date BETWEEN NOW() AND NOW() + $1::interval
		AND status NOT IN ('REVIEWED', 'OVERDUE')
	`

	rows, err := r.db.QueryContext(ctx, query, duration)
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
			&a.Status,
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
			(id, tutor_id, student_id, title, description, file_id, due_date, status, created_at, edited_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	id, err := uuid.NewRandom()
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
		assignment.Status,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	assignment.ID = id.String()
	return nil
}

func (r *AssignmentRepository) Update(ctx context.Context, assignment *domain.Assignment) error {
	query := `
		UPDATE assignments 
		SET title = $1, description = $2, file_id = $3, due_date = $4, 
		    status = $5, edited_at = $6
		WHERE id = $7
	`
	result, err := r.db.ExecContext(ctx, query,
		assignment.Title,
		assignment.Description,
		assignment.FileID,
		assignment.DueDate,
		assignment.Status,
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

func (r *AssignmentRepository) GetByID(ctx context.Context, id string) (*domain.Assignment, error) {
	query := `
		SELECT id, tutor_id, student_id, title, description, file_id, due_date, 
		       status, created_at, edited_at
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
		&assignment.Status,
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

func (r *AssignmentRepository) ListByTutorID(ctx context.Context, tutorID string) ([]*domain.Assignment, error) {
	query := `
		SELECT id, tutor_id, student_id, title, description, file_id, due_date, 
		       status, created_at, edited_at
		FROM assignments
		WHERE tutor_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tutorID)
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
			&a.Status,
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

func (r *AssignmentRepository) ListByStudentID(ctx context.Context, studentID string) ([]*domain.Assignment, error) {
	query := `
		SELECT id, tutor_id, student_id, title, description, file_id, due_date, 
		       status, created_at, edited_at
		FROM assignments
		WHERE student_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, studentID)
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
			&a.Status,
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

func (r *AssignmentRepository) Delete(ctx context.Context, id string) error {
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

func (r *AssignmentRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE assignments 
		SET status = $1, edited_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update assignment status: %w", err)
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
