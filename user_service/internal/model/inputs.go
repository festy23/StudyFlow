package model

import "github.com/google/uuid"

type RegisterViaTelegramInput struct {
	TelegramId int64
	Role       Role
	Username   *string
	FirstName  *string
	LastName   *string
	Timezone   *string
}

type AuthorizeInput struct {
	AuthorizationHeader string
}

type UpdateUserInput struct {
	FirstName *string
	LastName  *string
	Timezone  *string
}

type CreateTutorStudentInput struct {
	TutorId              uuid.UUID
	StudentId            uuid.UUID
	LessonPriceRub       *int32
	LessonConnectionLink *string
}

type UpdateTutorStudentInput struct {
	LessonPriceRub       *int32
	LessonConnectionLink *string
	Status               *TutorStudentStatus
}

type UpdateTutorProfileInput struct {
	PaymentInfo          *string
	LessonPriceRub       *int32
	LessonConnectionLink *string
}
