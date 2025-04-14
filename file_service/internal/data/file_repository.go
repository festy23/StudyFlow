package data

import (
	"context"
	"fileservice/internal/model"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) CreateFile(ctx context.Context, input *model.RepositoryCreateFileInput) (*model.File, error) {
	query := `
INSERT INTO files (
 id, extension, uploaded_by, filename
)
VALUES ($1, $2, $3, $4)
RETURNING id, extension, uploaded_by, filename, created_at
`
	var file model.File
	err := pgxscan.Get(ctx, r.db, &file, query,
		input.Id,
		input.Extension,
		input.UploadedBy,
		input.Filename,
	)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) GetFile(ctx context.Context, fileId uuid.UUID) (*model.File, error) {
	query := `
SELECT id, extension, uploaded_by, filename, created_at
FROM files
WHERE id = $1
`
	var file model.File
	err := pgxscan.Get(ctx, r.db, &file, query, fileId)
	if err != nil {
		return nil, err
	}

	return &file, nil
}
