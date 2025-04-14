package model

import "github.com/google/uuid"

type InitUploadInput struct {
	UploadedBy uuid.UUID
	Filename   string
}
