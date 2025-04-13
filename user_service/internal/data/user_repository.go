package data

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"userservice/internal/model"
	"userservice/internal/service"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) NewUserCreationRepositoryTx(ctx context.Context) (service.UserCreationRepositoryTx, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &UserCreationRepository{tx: tx}, nil
}

func (r *UserRepository) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
SELECT 
	id, role, auth_provider, status,
	first_name, last_name, timezone,
	created_at, edited_at

FROM users
WHERE id = $1
`
	var user model.User
	err := pgxscan.Get(ctx, r.db, &user, query, id)
	if err != nil {
		return nil, handleError(err)
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, id uuid.UUID, input *model.UpdateUserInput) (*model.User, error) {
	query, args := buildUserUpdateQuery(input)
	args = append(args, id)
	var user model.User
	err := pgxscan.Get(ctx, r.db, &user, query, args...)
	if err != nil {
		return nil, handleError(err)
	}
	return &user, nil
}

func (r *UserRepository) GetTutorProfile(ctx context.Context, userId uuid.UUID) (*model.TutorProfile, error) {
	query := `
SELECT 
	id, user_id, payment_info, 
	lesson_price_rub, lesson_connection_link, 
	created_at, edited_at 

FROM tutor_profiles
WHERE user_id = $1
`

	var user model.TutorProfile
	err := pgxscan.Get(ctx, r.db, &user, query, userId)
	if err != nil {
		return nil, handleError(err)
	}
	return &user, nil
}

func (r *UserRepository) UpdateTutorProfile(ctx context.Context, userId uuid.UUID, input *model.UpdateTutorProfileInput) (*model.TutorProfile, error) {
	query, args := buildTutorProfileUpdateQuery(input)
	args = append(args, userId)

	var user model.TutorProfile
	err := pgxscan.Get(ctx, r.db, &user, query, args...)
	if err != nil {
		return nil, handleError(err)
	}

	return &user, nil
}

func (r *UserRepository) GetTelegramAccount(ctx context.Context, userId uuid.UUID) (*model.TelegramAccount, error) {
	query := `
SELECT id, user_id, telegram_id, username, created_at
FROM telegram_accounts
WHERE user_id = $1
`
	var telegramAccount model.TelegramAccount
	err := pgxscan.Get(ctx, r.db, &telegramAccount, query, userId)
	if err != nil {
		return nil, handleError(err)
	}
	return &telegramAccount, nil
}

func (r *UserRepository) GetTelegramAccountByTelegramId(ctx context.Context, telegramId int64) (*model.TelegramAccount, error) {
	query := `
SELECT id, user_id, telegram_id, username, created_at
FROM telegram_accounts
WHERE telegram_id = $1
`
	var telegramAccount model.TelegramAccount
	err := pgxscan.Get(ctx, r.db, &telegramAccount, query, telegramId)
	if err != nil {
		return nil, handleError(err)
	}
	return &telegramAccount, nil
}

type UserCreationRepositoryTxInterface interface {
	CreateUser(ctx context.Context, input *model.RepositoryCreateUserInput) (*model.User, error)
	CreateTutorProfile(ctx context.Context, input *model.RepositoryCreateTutorProfileInput) (*model.TutorProfile, error)
	CreateTelegramAccount(ctx context.Context, input *model.RepositoryCreateTelegramAccountInput) (*model.TelegramAccount, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type UserCreationRepository struct {
	tx pgx.Tx
}

func (r *UserCreationRepository) CreateUser(ctx context.Context, input *model.RepositoryCreateUserInput) (*model.User, error) {
	query := `
INSERT INTO users (
	id, role, auth_provider, status,
	first_name, last_name, timezone
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING 
	id, role, auth_provider, status,
	first_name, last_name, timezone,
	created_at, edited_at
`

	var user model.User
	err := pgxscan.Get(ctx, r.tx, &user, query,
		input.Id,
		input.Role,
		input.AuthProvider,
		input.Status,
		input.FirstName,
		input.LastName,
		input.Timezone,
	)
	if err != nil {
		return nil, handleError(err)
	}

	return &user, nil
}

func (r *UserCreationRepository) CreateTutorProfile(ctx context.Context, input *model.RepositoryCreateTutorProfileInput) (*model.TutorProfile, error) {
	query := `
INSERT INTO tutor_profiles (
	id, user_id, payment_info, 
	lesson_price_rub, lesson_connection_link
)
VALUES ($1, $2, $3, $4, $5)
RETURNING 
	id, user_id, payment_info, 
	lesson_price_rub, lesson_connection_link,
	created_at, edited_at

`
	var profile model.TutorProfile
	err := pgxscan.Get(ctx, r.tx, &profile, query,
		input.Id,
		input.UserId,
		input.PaymentInfo,
		input.LessonPriceRub,
		input.LessonConnectionLink,
	)
	if err != nil {
		return nil, handleError(err)
	}

	return &profile, nil
}

func (r *UserCreationRepository) CreateTelegramAccount(ctx context.Context, input *model.RepositoryCreateTelegramAccountInput) (*model.TelegramAccount, error) {
	query := `
INSERT INTO telegram_accounts (id, user_id, telegram_id, username)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, telegram_id, username, created_at
`
	var account model.TelegramAccount
	err := pgxscan.Get(ctx, r.tx, &account, query,
		input.Id,
		input.UserId,
		input.TelegramId,
		input.Username,
	)
	if err != nil {
		return nil, handleError(err)
	}

	return &account, nil
}

func (r *UserCreationRepository) Commit(ctx context.Context) error {
	err := r.tx.Commit(ctx)
	return err
}

func (r *UserCreationRepository) Rollback(ctx context.Context) error {
	err := r.tx.Rollback(ctx)
	return err
}
