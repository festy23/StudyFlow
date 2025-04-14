package service

import (
	"context"
	"fileservice/internal/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type FileRepository interface {
	GetFile(ctx context.Context, fileId uuid.UUID) (*model.File, error)
}

type s3Client interface {
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
}

type FileService struct {
	fileRepo FileRepository
	s3Client s3Client
}

func NewFileService(ctx context.Context, fileRepo FileRepository, client s3Client, bucketName string) (*FileService, error) {
	s := &FileService{fileRepo: fileRepo, s3Client: client}
	err := s.createBucket(ctx, bucketName)
	return s, err
}

func (s *FileService) InitUpload(ctx context.Context, input *model.InitUploadInput) (*model.InitUpload, error) {

}

func (s *FileService) GenerateDownloadURL(ctx context.Context, fileId uuid.UUID) (string, error) {

}

func (s *FileService) GetFileMeta(ctx context.Context, fileId uuid.UUID) (*model.File, error) {
	return s.fileRepo.GetFile(ctx, fileId)
}

func (s *FileService) createBucket(ctx context.Context, name string) error {
	_, err := s.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(name)})
	return err
}
