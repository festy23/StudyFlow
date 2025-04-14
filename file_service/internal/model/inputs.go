package model

import (
	"github.com/google/uuid"
)

type InitUploadInput struct {
	UploadedBy uuid.UUID
	Filename   string
}

type RepositoryCreateFileInput struct {
	Id         uuid.UUID
	Extension  string
	UploadedBy uuid.UUID
	Filename   *string
}
