package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"stratyx/backend/internal/domain/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *ItemService) UploadAttachment(ctx context.Context, ownerID, itemID string, fileHeader *multipart.FileHeader) (*models.FileAsset, error) {
	if fileHeader == nil {
		return nil, errors.New("file is required")
	}
	if fileHeader.Size > 10*1024*1024 {
		return nil, errors.New("file size exceeds 10MB")
	}
	contentType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") && !strings.HasPrefix(contentType, "application/pdf") {
		return nil, errors.New("unsupported file type")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	raw, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	storageKind := strings.ToLower(strings.TrimSpace(os.Getenv("UPLOAD_STORAGE")))
	if storageKind == "" {
		storageKind = "local"
	}

	key := fmt.Sprintf("%s/%s-%s", ownerID, uuid.NewString(), sanitizeFileName(fileHeader.Filename))
	url := ""
	switch storageKind {
	case "s3":
		url, err = uploadToS3(ctx, key, contentType, raw)
	default:
		url, err = uploadLocalPrivate(key, raw)
		storageKind = "local"
	}
	if err != nil {
		return nil, err
	}

	asset := &models.FileAsset{
		ID:          uuid.NewString(),
		OwnerID:     ownerID,
		ItemID:      itemID,
		FileName:    fileHeader.Filename,
		ContentType: contentType,
		SizeBytes:   fileHeader.Size,
		URL:         url,
		StorageKind: storageKind,
		CreatedAt:   time.Now(),
	}
	_, err = s.repos.Files.InsertOne(ctx, asset)
	return asset, err
}

func (s *AuthService) UploadAvatar(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", errors.New("file is required")
	}
	if fileHeader.Size > 5*1024*1024 {
		return "", errors.New("avatar size exceeds 5MB")
	}
	contentType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", errors.New("avatar must be an image")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()
	raw, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf("avatars/%s-%s", userID, sanitizeFileName(fileHeader.Filename))
	url, err := uploadLocal(key, raw)
	if err != nil {
		return "", err
	}
	_, err = s.repos.Users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{
		"avatar_url": url,
		"updated_at": time.Now(),
	}})
	if err != nil {
		return "", err
	}
	return url, nil
}

func uploadLocal(key string, raw []byte) (string, error) {
	basePath := os.Getenv("UPLOAD_LOCAL_PATH")
	if basePath == "" {
		basePath = "./uploads"
	}
	fullPath := filepath.Join(basePath, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(fullPath, raw, 0o644); err != nil {
		return "", err
	}
	return "/uploads/" + filepath.ToSlash(key), nil
}

func uploadLocalPrivate(key string, raw []byte) (string, error) {
	basePath := os.Getenv("UPLOAD_LOCAL_PATH")
	if basePath == "" {
		basePath = "./uploads"
	}
	fullPath := filepath.Join(basePath, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(fullPath, raw, 0o644); err != nil {
		return "", err
	}
	return key, nil
}

func uploadToS3(ctx context.Context, key, contentType string, raw []byte) (string, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	region := os.Getenv("S3_REGION")
	bucket := os.Getenv("S3_BUCKET")
	ak := os.Getenv("S3_ACCESS_KEY_ID")
	sk := os.Getenv("S3_SECRET_ACCESS_KEY")
	if endpoint == "" || region == "" || bucket == "" || ak == "" || sk == "" {
		return "", errors.New("s3 storage is not fully configured")
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(ak, sk, "")),
	)
	if err != nil {
		return "", err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = strings.EqualFold(os.Getenv("S3_USE_PATH_STYLE"), "true")
		}
	})
	uploader := manager.NewUploader(client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(raw),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}
	if endpoint != "" {
		return strings.TrimRight(endpoint, "/") + "/" + bucket + "/" + key, nil
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key), nil
}

func sanitizeFileName(name string) string {
	clean := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	if clean == "" {
		return "file.bin"
	}
	return clean
}
