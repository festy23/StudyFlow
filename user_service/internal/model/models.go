package model

import (
	"github.com/google/uuid"
	"time"
)

type Role string

const (
	RoleStudent Role = "student"
	RoleTutor   Role = "tutor"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	return r == RoleStudent || r == RoleTutor
}

type AuthProvider string

const (
	AuthProviderTelegram AuthProvider = "telegram"
)

func (a AuthProvider) String() string {
	return string(a)
}

func (a AuthProvider) IsValid() bool {
	return a == AuthProviderTelegram
}

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusDeleted UserStatus = "deleted"
)

func (u UserStatus) String() string {
	return string(u)
}

func (u UserStatus) IsValid() bool {
	return u == UserStatusActive || u == UserStatusDeleted
}

type User struct {
	Id           uuid.UUID    `db:"id"`
	Role         Role         `db:"role"`
	AuthProvider AuthProvider `db:"auth_provider"`
	Status       UserStatus   `db:"status"`
	FirstName    *string      `db:"first_name"`
	LastName     *string      `db:"last_name"`
	Timezone     *string      `db:"timezone"`
	CreatedAt    time.Time    `db:"created_at"`
	EditedAt     time.Time    `db:"edited_at"`
}

type TelegramAccount struct {
	Id         uuid.UUID `db:"id"`
	UserId     uuid.UUID `db:"user_id"`
	TelegramId int64     `db:"telegram_id"`
	Username   *string   `db:"username"`
	CreatedAt  time.Time `db:"created_at"`
}

type TutorProfile struct {
	Id                   uuid.UUID `db:"id"`
	UserId               uuid.UUID `db:"user_id"`
	PaymentInfo          *string   `db:"payment_info"`
	LessonPriceRub       *int32    `db:"lesson_price_rub"`
	LessonConnectionLink *string   `db:"lesson_connection_link"`
	CreatedAt            time.Time `db:"created_at"`
	EditedAt             time.Time `db:"edited_at"`
}

type TutorStudentStatus string

const (
	TutorStudentStatusActive  TutorStudentStatus = "active"
	TutorStudentStatusInvited TutorStudentStatus = "invited"
)

func (t TutorStudentStatus) String() string {
	return string(t)
}

func (t TutorStudentStatus) IsValid() bool {
	return t == TutorStudentStatusInvited || t == TutorStudentStatusActive
}

func TutorStudentStatusFromString(s string) (TutorStudentStatus, bool) {
	status := TutorStudentStatus(s)
	return status, status.IsValid()
}

type TutorStudent struct {
	Id                   uuid.UUID          `db:"id"`
	TutorId              uuid.UUID          `db:"tutor_id"`
	StudentId            uuid.UUID          `db:"student_id"`
	LessonPriceRub       *int32             `db:"lesson_price_rub"`
	LessonConnectionLink *string            `db:"lesson_connection_link"`
	Status               TutorStudentStatus `db:"status"`
	CreatedAt            time.Time          `db:"created_at"`
	EditedAt             time.Time          `db:"edited_at"`
}

// TutorStudentContext not from db
type TutorStudentContext struct {
	RelationshipStatus TutorStudentStatus

	LessonPriceRub       *int32
	LessonConnectionLink *string
	PaymentInfo          *string
}

type UserPublic struct {
	Id        uuid.UUID
	Role      Role
	FirstName *string
	LastName  *string
}
