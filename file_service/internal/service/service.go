package service

import (
	"common_library/logging"
	"context"
	"database/sql"
	"errors"
	"fileservice/internal/errdefs"
	"fileservice/internal/model"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"path"
	"strings"
	"time"
)

type FileRepository interface {
	CreateFile(ctx context.Context, input *model.RepositoryCreateFileInput) (*model.File, error)
	GetFile(ctx context.Context, fileId uuid.UUID) (*model.File, error)
}

type FileService struct {
	fileRepo         FileRepository
	s3Client         *s3.Client
	bucket           *string
	gatewayPublicUrl string
	minioURL         string
}

func NewFileService(ctx context.Context, fileRepo FileRepository, client *s3.Client, bucketName string, gatewayPublicUrl string, minioUrl string) (*FileService, error) {
	s := &FileService{fileRepo: fileRepo, s3Client: client, bucket: aws.String(bucketName), gatewayPublicUrl: gatewayPublicUrl, minioURL: minioUrl}
	err := s.createBucket(ctx, bucketName)
	return s, err
}

func (s *FileService) InitUpload(ctx context.Context, input *model.InitUploadInput) (*model.InitUpload, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	extension := path.Ext(input.Filename)
	if extension == "" {
		return nil, fmt.Errorf("invalid file extension: %w", errdefs.ValidationErr)
	}
	fileInput := &model.RepositoryCreateFileInput{
		Id:         id,
		Extension:  extension,
		UploadedBy: input.UploadedBy,
		Filename:   &input.Filename,
	}

	file, err := s.fileRepo.CreateFile(ctx, fileInput)
	if err != nil {
		return nil, err
	}

	key := file.Id.String() + file.Extension
	uploadRequest, err := s.generateUploadURL(ctx, key)
	if err != nil {
		return nil, err
	}

	res := &model.InitUpload{
		FileId:    file.Id,
		UploadURL: strings.Replace(uploadRequest.URL, s.minioURL, s.gatewayPublicUrl+"/files/upload", 1),
		Method:    uploadRequest.Method,
	}

	return res, nil
}

func (s *FileService) GenerateDownloadURL(ctx context.Context, fileId uuid.UUID) (string, error) {
	file, err := s.fileRepo.GetFile(ctx, fileId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("file not found: %w", errdefs.ErrNotFound)
		}
	}

	key := file.Id.String() + file.Extension
	downloadRequest, err := s.generateDownloadURL(ctx, key)
	if err != nil {
		return "", err
	}

	return strings.Replace(downloadRequest.URL, s.minioURL, s.gatewayPublicUrl+"/files/download", 1), nil
}

func (s *FileService) GetFileMeta(ctx context.Context, fileId uuid.UUID) (*model.File, error) {
	file, err := s.fileRepo.GetFile(ctx, fileId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("file not found: %w", errdefs.ErrNotFound)
		}
	}
	return file, nil
}

func (s *FileService) createBucket(ctx context.Context, name string) error {
	_, err := s.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(name)})
	if err != nil {
		var opErr *http.ResponseError
		if errors.As(err, &opErr) && opErr.HTTPStatusCode() == 409 {
			if logger, ok := logging.GetFromContext(ctx); ok {
				logger.Info(ctx, "Bucket already exists", zap.String("bucket", name))
			}
			return nil
		}
	}
	return err
}

func (s *FileService) generateUploadURL(ctx context.Context, key string) (*v4.PresignedHTTPRequest, error) {
	presigner := s3.NewPresignClient(s.s3Client)

	req, err := presigner.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket: s.bucket,
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(5*time.Minute),
	)
	return req, err
}

func (s *FileService) generateDownloadURL(ctx context.Context, key string) (*v4.PresignedHTTPRequest, error) {
	presigner := s3.NewPresignClient(s.s3Client)

	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(key),
	},
		s3.WithPresignExpires(5*time.Minute),
	)

	return req, err
}
