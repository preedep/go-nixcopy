package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	appconfig "github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

type S3Storage struct {
	config   *appconfig.S3Config
	s3Client *s3.Client
}

func NewS3Storage(cfg *appconfig.S3Config) repository.Storage {
	return &S3Storage{
		config: cfg,
	}
}

func (s *S3Storage) Connect(ctx context.Context) error {
	var awsCfg aws.Config
	var err error

	configOptions := []func(*config.LoadOptions) error{
		config.WithRegion(s.config.Region),
	}

	if s.config.Profile != "" {
		configOptions = append(configOptions, config.WithSharedConfigProfile(s.config.Profile))
	}

	switch s.config.AuthType {
	case appconfig.S3AuthAccessKey:
		if s.config.AccessKeyID == "" || s.config.SecretAccessKey == "" {
			return fmt.Errorf("access_key_id and secret_access_key are required for access_key auth type")
		}
		configOptions = append(configOptions, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				s.config.AccessKeyID,
				s.config.SecretAccessKey,
				s.config.SessionToken,
			),
		))
		awsCfg, err = config.LoadDefaultConfig(ctx, configOptions...)

	case appconfig.S3AuthIAMRole, appconfig.S3AuthInstanceProfile:
		awsCfg, err = config.LoadDefaultConfig(ctx, configOptions...)

	case appconfig.S3AuthAssumeRole:
		if s.config.RoleARN == "" {
			return fmt.Errorf("role_arn is required for assume_role auth type")
		}
		awsCfg, err = config.LoadDefaultConfig(ctx, configOptions...)
		if err != nil {
			return fmt.Errorf("failed to load AWS config: %w", err)
		}

	case appconfig.S3AuthWebIdentity:
		if s.config.RoleARN == "" || s.config.WebIdentityTokenFile == "" {
			return fmt.Errorf("role_arn and web_identity_token_file are required for web_identity auth type")
		}
		configOptions = append(configOptions, config.WithWebIdentityRoleCredentialOptions(func(options *config.WebIdentityRoleOptions) {
			options.RoleARN = s.config.RoleARN
		}))
		awsCfg, err = config.LoadDefaultConfig(ctx, configOptions...)

	default:
		awsCfg, err = config.LoadDefaultConfig(ctx, configOptions...)
	}

	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	options := func(o *s3.Options) {
		if s.config.Endpoint != "" {
			o.BaseEndpoint = aws.String(s.config.Endpoint)
		}
		if s.config.UsePathStyle {
			o.UsePathStyle = true
		}
	}

	s.s3Client = s3.NewFromConfig(awsCfg, options)

	return nil
}

func (s *S3Storage) Disconnect(ctx context.Context) error {
	return nil
}

func (s *S3Storage) List(ctx context.Context, path string) ([]entity.FileInfo, error) {
	if s.s3Client == nil {
		return nil, fmt.Errorf("S3 client not connected")
	}

	prefix := strings.TrimPrefix(path, "/")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.Bucket),
		Prefix: aws.String(prefix),
	}

	result, err := s.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	fileInfos := make([]entity.FileInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		fileInfos = append(fileInfos, entity.FileInfo{
			Path:         *obj.Key,
			Name:         filepath.Base(*obj.Key),
			Size:         *obj.Size,
			ModifiedTime: *obj.LastModified,
			IsDirectory:  false,
		})
	}

	return fileInfos, nil
}

func (s *S3Storage) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	if s.s3Client == nil {
		return nil, 0, fmt.Errorf("S3 client not connected")
	}

	key := strings.TrimPrefix(path, "/")

	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	headResult, err := s.s3Client.HeadObject(ctx, headInput)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get object metadata: %w", err)
	}

	getInput := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := s.s3Client.GetObject(ctx, getInput)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get object: %w", err)
	}

	return result.Body, *headResult.ContentLength, nil
}

func (s *S3Storage) Stat(ctx context.Context, path string) (*entity.FileInfo, error) {
	if s.s3Client == nil {
		return nil, fmt.Errorf("S3 client not connected")
	}

	key := strings.TrimPrefix(path, "/")

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := s.s3Client.HeadObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	return &entity.FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		Size:         *result.ContentLength,
		ModifiedTime: *result.LastModified,
		IsDirectory:  false,
	}, nil
}

func (s *S3Storage) Write(ctx context.Context, path string, reader io.Reader, size int64) error {
	if s.s3Client == nil {
		return fmt.Errorf("S3 client not connected")
	}

	key := strings.TrimPrefix(path, "/")

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
		Body:   reader,
	}

	_, err := s.s3Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	return nil
}

func (s *S3Storage) CreateDirectory(ctx context.Context, path string) error {
	return nil
}

func (s *S3Storage) Delete(ctx context.Context, path string) error {
	if s.s3Client == nil {
		return fmt.Errorf("S3 client not connected")
	}

	key := strings.TrimPrefix(path, "/")

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.s3Client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
