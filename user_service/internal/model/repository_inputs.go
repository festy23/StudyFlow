package model

import (
	"github.com/google/uuid"
)

type RepositoryCreateUserInput struct {
	Id           uuid.UUID    `db:"id"`
	Role         Role         `db:"role"`
	AuthProvider AuthProvider `db:"auth_provider"`
	Status       UserStatus   `db:"status"`
	FirstName    *string      `db:"first_name"`
	LastName     *string      `db:"last_name"`
	Timezone     *string      `db:"timezone"`
}

type RepositoryCreateTutorProfileInput struct {
	Id                   uuid.UUID `db:"id"`
	UserId               uuid.UUID `db:"user_id"`
	PaymentInfo          *string   `db:"payment_info"`
	LessonPriceRub       *int32    `db:"lesson_price_rub"`
	LessonConnectionLink *string   `db:"lesson_connection_link"`
}

type RepositoryCreateTelegramAccountInput struct {
	Id         uuid.UUID `db:"id"`
	UserId     uuid.UUID `db:"user_id"`
	TelegramId int64     `db:"telegram_id"`
	Username   *string   `db:"user_name"`
}

type RepositoryCreateTutorStudentInput struct {
	Id                   uuid.UUID          `db:"id"`
	TutorId              uuid.UUID          `db:"tutor_id"`
	StudentId            uuid.UUID          `db:"student_id"`
	LessonPriceRub       *int32             `db:"lesson_price_rub"`
	LessonConnectionLink *string            `db:"lesson_connection_link"`
	Status               TutorStudentStatus `db:"status"`
}
