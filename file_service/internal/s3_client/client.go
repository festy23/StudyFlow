package s3_client

import (
	"context"
	"fileservice/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3Config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func New(ctx context.Context, cfg *config.Config) (*s3.Client, error) {
	s3Cfg, err := s3Config.LoadDefaultConfig(ctx,
		s3Config.WithRegion(cfg.S3Region),
		s3Config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.S3AccessKeyID,
				cfg.S3SecretAccessKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(s3Cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3Endpoint)
		o.UsePathStyle = true
	})
	return s3Client, nil
}
