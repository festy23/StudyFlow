package model

type RegisterViaTelegramInput struct {
	TelegramId string
	Username   string
	FirstName  string
	LastName   string
	Timezone   string
}

type AuthorizeInput struct {
	AuthorizationHeader string
}

type UpdateUserInput struct {
	FirstName string
	LastName  string
	Timezone  string
}

type CreateTutorStudentInput struct {
	TutorId              string
	StudentId            string
	LessonPriceRub       int
	LessonConnectionLink string
}

type UpdateTutorStudentInput struct {
	LessonPriceRub       int
	LessonConnectionLink string
	Status               TutorStudentStatus
}

type UpdateTutorProfileInput struct {
	PaymentInfo          string
	LessonPriceRub       int
	LessonConnectionLink string
}
