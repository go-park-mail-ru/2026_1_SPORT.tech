package minio

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/config"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type AvatarStorage struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

func NewAvatarStorage(cfg config.StorageConfig) (*AvatarStorage, error) {
	client, err := minio.New(cfg.Endpoint(), &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &AvatarStorage{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

func (storage *AvatarStorage) UploadAvatar(
	ctx context.Context,
	userID int64,
	fileName string,
	contentType string,
	file io.Reader,
	size int64,
) (string, error) {
	if err := storage.ensureBucket(ctx); err != nil {
		return "", err
	}

	objectName := fmt.Sprintf("users/%d/%s%s", userID, randomObjectID(), avatarExtension(contentType, fileName))
	_, err := storage.client.PutObject(ctx, storage.bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return storage.publicBaseURL + "/" + objectName, nil
}

func (storage *AvatarStorage) DeleteAvatar(ctx context.Context, avatarURL string) error {
	objectName := storage.avatarObjectName(avatarURL)
	if objectName == "" {
		return nil
	}

	return storage.client.RemoveObject(ctx, storage.bucket, objectName, minio.RemoveObjectOptions{})
}

func (storage *AvatarStorage) ensureBucket(ctx context.Context) error {
	exists, err := storage.client.BucketExists(ctx, storage.bucket)
	if err != nil {
		return err
	}

	if !exists {
		if err := storage.client.MakeBucket(ctx, storage.bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	policy := fmt.Sprintf(`{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{"AWS":["*"]},
      "Action":["s3:GetObject"],
      "Resource":["arn:aws:s3:::%s/*"]
    }
  ]
}`, storage.bucket)

	return storage.client.SetBucketPolicy(ctx, storage.bucket, policy)
}

func avatarExtension(contentType string, _ string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func (storage *AvatarStorage) avatarObjectName(avatarURL string) string {
	parsedURL, err := url.Parse(avatarURL)
	if err != nil {
		return ""
	}

	path := strings.Trim(parsedURL.Path, "/")
	bucketPrefix := strings.Trim(storage.bucket, "/") + "/"
	if strings.HasPrefix(path, bucketPrefix) {
		return strings.TrimPrefix(path, bucketPrefix)
	}

	return ""
}

func randomObjectID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "avatar"
	}

	return hex.EncodeToString(buffer)
}
