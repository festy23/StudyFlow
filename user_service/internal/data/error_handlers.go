package data

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"userservice/internal/errdefs"
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func handleError(err error) error {
	if isUniqueViolation(err) {
		return errdefs.ErrAlreadyExists
	}
	if isNotFound(err) {
		return errdefs.ErrNotFound
	}
	return fmt.Errorf("repository error: %w", err)
}
