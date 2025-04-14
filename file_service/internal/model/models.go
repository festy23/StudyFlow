package model

import (
	"github.com/google/uuid"
	"time"
)

type File struct {
	Id         uuid.UUID `db:"id"`
	Extension  string    `db:"extension"`
	UploadedBy uuid.UUID `db:"uploaded_by"`
	Filename   *string   `db:"filename"`
	CreatedAt  time.Time `db:"created_at"`
}

type InitUpload struct {
	FileId    uuid.UUID
	UploadURL string
	Method    string
}
