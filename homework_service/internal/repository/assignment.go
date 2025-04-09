package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"homework_service/internal/domain"
)

type AssignmentRepository struct {
	db *sql.DB
}

func (r *AssignmentRepository) FindAssignmentsDueSoon(ctx context.Context, duration time.Duration) ([]*domain.Assignment, error) {
	query := `
		SELECT id, tutor_id, student_id, title, description, file_id, due_date, created_at, edited_at
		FROM assignments
		WHERE due_date BETWEEN NOW() AND NOW() + $1::interval
		AND status NOT IN ('REVIEWED', 'OVERDUE')
	`

	rows, err := r.db.QueryContext(ctx, query, duration)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		assignments = append(assignments, &a)
	}

	return assignments, nil
}

func NewAssignmentRepository(db *sql.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) Create(ctx context.Context, assignment *domain.Assignment) error {
	query := `
		INSERT INTO assignments (id, tutor_id, student_id, title, description, file_id, due_date, created_at, edited_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	id, err := uuid.NewRandom()
	if err != nil {
		return err
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
		return err
	}

	assignment.ID = id.String()
	return nil
}

func (r *AssignmentRepository) GetByID(ctx context.Context, id string) (*domain.Assignment, error) {
	query := `
        SELECT id, tutor_id, student_id, title, description, file_id, due_date, status, 
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
		&assignment.Status,
		&assignment.CreatedAt,
		&assignment.EditedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("assignment not found")
		}
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	return &assignment, nil
}
