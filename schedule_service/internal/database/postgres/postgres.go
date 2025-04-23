package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	repo "schedule_service/internal/database/repo"
	service "schedule_service/internal/service/service"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func (r *PostgresRepository) GetSlot(ctx context.Context, id string) (*repo.Slot, error) {
	query := `
		SELECT id, tutor_id, starts_at, ends_at, is_booked, created_at, edited_at
		FROM slots
		WHERE id = $1
	`

	var slot repo.Slot
	var editedAt pgtype.Timestamp

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&slot.ID,
		&slot.TutorID,
		&slot.StartsAt,
		&slot.EndsAt,
		&slot.IsBooked,
		&slot.CreatedAt,
		&editedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrSlotNotFound
		}
		return nil, fmt.Errorf("failed to get slot: %w", err)
	}

	if editedAt.Valid {
		slot.EditedAt = &editedAt.Time
	} else {
		slot.EditedAt = &slot.CreatedAt
	}

	return &slot, nil
}

func (r *PostgresRepository) CreateSlot(ctx context.Context, slot repo.Slot) error {
	query := `
		INSERT INTO slots (id, tutor_id, starts_at, ends_at, is_booked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		slot.ID,
		slot.TutorID,
		slot.StartsAt,
		slot.EndsAt,
		slot.IsBooked,
		slot.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create slot: %w", err)
	}

	return nil
}

func (r *PostgresRepository) UpdateSlot(ctx context.Context, slot repo.Slot) error {
	query := `
		UPDATE slots
		SET starts_at = $1, ends_at = $2, edited_at = $3
		WHERE id = $4
	`

	res, err := r.pool.Exec(ctx, query,
		slot.StartsAt,
		slot.EndsAt,
		slot.EditedAt,
		slot.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update slot: %w", err)
	}

	if res.RowsAffected() == 0 {
		return service.ErrSlotNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteSlot(ctx context.Context, id string) error {
	query := `DELETE FROM slots WHERE id = $1`

	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete slot: %w", err)
	}

	if res.RowsAffected() == 0 {
		return service.ErrSlotNotFound
	}

	return nil
}

func (r *PostgresRepository) ListSlotsByTutor(ctx context.Context, tutorID string, onlyAvailable bool) ([]repo.Slot, error) {
	var query string
	var args []interface{}

	if onlyAvailable {
		query = `
			SELECT id, tutor_id, starts_at, ends_at, is_booked, created_at, edited_at
			FROM slots
			WHERE tutor_id = $1 AND is_booked = false
			ORDER BY starts_at ASC
		`
		args = []interface{}{tutorID}
	} else {
		query = `
			SELECT id, tutor_id, starts_at, ends_at, is_booked, created_at, edited_at
			FROM slots
			WHERE tutor_id = $1
			ORDER BY starts_at ASC
		`
		args = []interface{}{tutorID}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list slots: %w", err)
	}
	defer rows.Close()

	var slots []repo.Slot
	for rows.Next() {
		var slot repo.Slot
		var editedAt pgtype.Timestamp

		err := rows.Scan(
			&slot.ID,
			&slot.TutorID,
			&slot.StartsAt,
			&slot.EndsAt,
			&slot.IsBooked,
			&slot.CreatedAt,
			&editedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan slot row: %w", err)
		}

		if editedAt.Valid {
			slot.EditedAt = &editedAt.Time
		}

		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating slot rows: %w", err)
	}

	return slots, nil
}

func (r *PostgresRepository) GetLesson(ctx context.Context, id string) (*repo.Lesson, error) {
	query := `
		SELECT id, slot_id, student_id, status, is_paid, connection_link, price_rub, payment_info, created_at, edited_at
		FROM lessons
		WHERE id = $1
	`

	var lesson repo.Lesson
	var connectionLink, paymentInfo pgtype.Text
	var priceRub pgtype.Int4

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&lesson.ID,
		&lesson.SlotID,
		&lesson.StudentID,
		&lesson.Status,
		&lesson.IsPaid,
		&connectionLink,
		&priceRub,
		&paymentInfo,
		&lesson.CreatedAt,
		&lesson.EditedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrLessonNotFound
		}
		return nil, fmt.Errorf("failed to get lesson: %w", err)
	}

	if connectionLink.Valid {
		lesson.ConnectionLink = &connectionLink.String
	}

	if priceRub.Valid {
		val := int32(priceRub.Int32)
		lesson.PriceRub = &val
	}

	if paymentInfo.Valid {
		lesson.PaymentInfo = &paymentInfo.String
	}

	return &lesson, nil
}

func (r *PostgresRepository) CreateLessonAndBookSlot(ctx context.Context, lesson repo.Lesson, slotID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var isBooked bool
	err = tx.QueryRow(ctx, "SELECT is_booked FROM slots WHERE id = $1 FOR UPDATE", slotID).Scan(&isBooked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.ErrSlotNotFound
		}
		return fmt.Errorf("failed to check slot availability: %w", err)
	}

	if isBooked {
		return service.ErrSlotBooked
	}

	_, err = tx.Exec(ctx, "UPDATE slots SET is_booked = true WHERE id = $1", slotID)
	if err != nil {
		return fmt.Errorf("failed to mark slot as booked: %w", err)
	}

	query := `
		INSERT INTO lessons (id, slot_id, student_id, status, is_paid, created_at, edited_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.Exec(ctx, query,
		lesson.ID,
		lesson.SlotID,
		lesson.StudentID,
		lesson.Status,
		lesson.IsPaid,
		lesson.CreatedAt,
		lesson.EditedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create lesson: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) UpdateLesson(ctx context.Context, lesson repo.Lesson) error {
	query := `
		UPDATE lessons
		SET status = $1, is_paid = $2, connection_link = $3, price_rub = $4, payment_info = $5, edited_at = $6
		WHERE id = $7
	`

	res, err := r.pool.Exec(ctx, query,
		lesson.Status,
		lesson.IsPaid,
		lesson.ConnectionLink,
		lesson.PriceRub,
		lesson.PaymentInfo,
		lesson.EditedAt,
		lesson.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update lesson: %w", err)
	}

	if res.RowsAffected() == 0 {
		return service.ErrLessonNotFound
	}

	return nil
}

func (r *PostgresRepository) CancelLessonAndFreeSlot(ctx context.Context, lesson repo.Lesson, slotID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		"UPDATE lessons SET status = $1, edited_at = $2 WHERE id = $3",
		lesson.Status,
		lesson.EditedAt,
		lesson.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update lesson status: %w", err)
	}

	_, err = tx.Exec(ctx, "UPDATE slots SET is_booked = false WHERE id = $1", slotID)
	if err != nil {
		return fmt.Errorf("failed to mark slot as available: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) ListLessonsByTutor(ctx context.Context, tutorID string, statusFilter []string) ([]repo.Lesson, error) {
	query := `
		SELECT l.id, l.slot_id, l.student_id, l.status, l.is_paid, l.connection_link, l.price_rub, l.payment_info, l.created_at, l.edited_at
		FROM lessons l
		JOIN slots s ON l.slot_id = s.id
		WHERE s.tutor_id = $1
	`

	args := []interface{}{tutorID}
	if len(statusFilter) > 0 {
		placeholders := make([]string, len(statusFilter))
		for i := range statusFilter {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			args = append(args, statusFilter[i])
		}
		query += " AND l.status IN (" + strings.Join(placeholders, ", ") + ")"
	}

	query += " ORDER BY s.starts_at ASC"

	return r.queryLessons(ctx, query, args...)
}

func (r *PostgresRepository) ListLessonsByStudent(ctx context.Context, studentID string, statusFilter []string) ([]repo.Lesson, error) {
	query := `
		SELECT l.id, l.slot_id, l.student_id, l.status, l.is_paid, l.connection_link, l.price_rub, l.payment_info, l.created_at, l.edited_at
		FROM lessons l
		JOIN slots s ON l.slot_id = s.id
		WHERE l.student_id = $1
	`

	args := []interface{}{studentID}
	if len(statusFilter) > 0 {
		placeholders := make([]string, len(statusFilter))
		for i := range statusFilter {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			args = append(args, statusFilter[i])
		}
		query += " AND l.status IN (" + strings.Join(placeholders, ", ") + ")"
	}

	query += " ORDER BY s.starts_at ASC"

	return r.queryLessons(ctx, query, args...)
}

func (r *PostgresRepository) ListLessonsByPair(ctx context.Context, tutorID, studentID string, statusFilter []string) ([]repo.Lesson, error) {
	query := `
		SELECT l.id, l.slot_id, l.student_id, l.status, l.is_paid, l.connection_link, l.price_rub, l.payment_info, l.created_at, l.edited_at
		FROM lessons l
		JOIN slots s ON l.slot_id = s.id
		WHERE s.tutor_id = $1 AND l.student_id = $2
	`

	args := []interface{}{tutorID, studentID}
	if len(statusFilter) > 0 {
		placeholders := make([]string, len(statusFilter))
		for i := range statusFilter {
			placeholders[i] = fmt.Sprintf("$%d", i+3)
			args = append(args, statusFilter[i])
		}
		query += " AND l.status IN (" + strings.Join(placeholders, ", ") + ")"
	}

	query += " ORDER BY s.starts_at ASC"

	return r.queryLessons(ctx, query, args...)
}

func (r *PostgresRepository) ListCompletedUnpaidLessons(ctx context.Context, after *time.Time) ([]repo.Lesson, error) {
	var query string
	var args []interface{}

	if after != nil {
		query = `
			SELECT l.id, l.slot_id, l.student_id, l.status, l.is_paid, l.connection_link, l.price_rub, l.payment_info, l.created_at, l.edited_at
			FROM lessons l
			JOIN slots s ON l.slot_id = s.id
			WHERE l.status = 'completed' AND l.is_paid = false AND s.ends_at > $1
			ORDER BY s.ends_at ASC
		`
		args = []interface{}{after}
	} else {
		query = `
			SELECT l.id, l.slot_id, l.student_id, l.status, l.is_paid, l.connection_link, l.price_rub, l.payment_info, l.created_at, l.edited_at
			FROM lessons l
			JOIN slots s ON l.slot_id = s.id
			WHERE l.status = 'completed' AND l.is_paid = false
			ORDER BY s.ends_at ASC
		`
		args = []interface{}{}
	}

	return r.queryLessons(ctx, query, args...)
}

func (r *PostgresRepository) UpdateCompletedLessons(ctx context.Context) (int, error) {
	query := `
		UPDATE lessons
		SET status = 'completed', edited_at = NOW()
		FROM slots
		WHERE lessons.slot_id = slots.id
		AND lessons.status = 'booked'
		AND slots.ends_at < NOW()
	`

	res, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to update completed lessons: %w", err)
	}

	return int(res.RowsAffected()), nil
}

func (r *PostgresRepository) queryLessons(ctx context.Context, query string, args ...interface{}) ([]repo.Lesson, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query lessons: %w", err)
	}
	defer rows.Close()

	var lessons []repo.Lesson
	for rows.Next() {
		var lesson repo.Lesson
		var connectionLink, paymentInfo pgtype.Text
		var priceRub pgtype.Int4

		err := rows.Scan(
			&lesson.ID,
			&lesson.SlotID,
			&lesson.StudentID,
			&lesson.Status,
			&lesson.IsPaid,
			&connectionLink,
			&priceRub,
			&paymentInfo,
			&lesson.CreatedAt,
			&lesson.EditedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lesson row: %w", err)
		}

		if connectionLink.Valid {
			lesson.ConnectionLink = &connectionLink.String
		}

		if priceRub.Valid {
			val := int32(priceRub.Int32)
			lesson.PriceRub = &val
		}

		if paymentInfo.Valid {
			lesson.PaymentInfo = &paymentInfo.String
		}

		lessons = append(lessons, lesson)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lesson rows: %w", err)
	}

	return lessons, nil
}

func (r *PostgresRepository) MarkAsPaid(ctx context.Context, lessonID string) error {
	query := `UPDATE lessons SET is_paid = TRUE WHERE id = $1`

	res, err := r.pool.Exec(ctx, query, lessonID)

	if err != nil {
		return fmt.Errorf("failed to mark as paid: %w", err)
	}

	if res.RowsAffected() == 0 {
		return service.ErrLessonNotFound
	}
	return nil

}
