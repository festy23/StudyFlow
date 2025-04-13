package data

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"userservice/internal/model"
)

type TutorStudentRepository struct {
	db *pgxpool.Pool
}

func NewTutorStudentRepository(db *pgxpool.Pool) *TutorStudentRepository {
	return &TutorStudentRepository{db: db}
}

func (r *TutorStudentRepository) CreateTutorStudent(ctx context.Context, input *model.RepositoryCreateTutorStudentInput) (*model.TutorStudent, error) {
	query := `
INSERT INTO tutor_students (
	id, tutor_id, student_id,
    lesson_price_rub, lesson_connection_link,
    status
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tutor_id, student_id,
    lesson_price_rub, lesson_connection_link,
    status, created_at, edited_at
`
	var ts model.TutorStudent
	err := pgxscan.Get(ctx, r.db, &ts, query,
		input.Id,
		input.TutorId,
		input.StudentId,
		input.LessonPriceRub,
		input.LessonConnectionLink,
		input.Status,
	)
	if err != nil {
		return nil, handleError(err)
	}
	return &ts, nil
}

func (r *TutorStudentRepository) UpdateTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID, input *model.UpdateTutorStudentInput) (*model.TutorStudent, error) {
	query, args := buildUpdateTutorStudentQuery(input)
	args = append(args, tutorId, studentId)

	var ts model.TutorStudent
	err := pgxscan.Get(ctx, r.db, &ts, query, args...)
	if err != nil {
		return nil, handleError(err)
	}

	return &ts, nil
}

func (r *TutorStudentRepository) GetTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) (*model.TutorStudent, error) {
	query := `
SELECT 
	id, tutor_id, student_id, 
	lesson_price_rub, lesson_connection_link,
	status, created_at, edited_at

FROM tutor_students
WHERE tutor_id = $1 AND student_id = $2
`
	var ts model.TutorStudent
	err := pgxscan.Get(ctx, r.db, &ts, query, tutorId, studentId)
	if err != nil {
		return nil, handleError(err)
	}

	return &ts, nil
}

func (r *TutorStudentRepository) DeleteTutorStudent(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) error {
	query := `
DELETE FROM tutor_students
WHERE tutor_id = $1 AND student_id = $2
`
	_, err := r.db.Exec(ctx, query, tutorId, studentId)
	if err != nil {
		return handleError(err)
	}
	return nil
}

func (r *TutorStudentRepository) ListTutorStudents(ctx context.Context, tutorId uuid.UUID, studentId uuid.UUID) ([]*model.TutorStudent, error) {
	query, args := buildListTutorStudentsQuery(tutorId, studentId)

	var rows []*model.TutorStudent
	err := pgxscan.Select(ctx, r.db, &rows, query, args...)
	if err != nil {
		return nil, handleError(err)
	}
	return rows, nil
}
