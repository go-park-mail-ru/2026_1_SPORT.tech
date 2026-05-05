package minio

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/config"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type PostMediaStorage struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

func NewPostMediaStorage(cfg config.StorageConfig) (*PostMediaStorage, error) {
	client, err := minio.New(cfg.Endpoint(), &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &PostMediaStorage{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

func (storage *PostMediaStorage) UploadPostMedia(
	ctx context.Context,
	authorUserID int64,
	fileName string,
	contentType string,
	file io.Reader,
	size int64,
) (string, error) {
	if err := storage.ensureBucket(ctx); err != nil {
		return "", err
	}

	objectName := fmt.Sprintf("posts/%d/%s%s", authorUserID, randomObjectID(), mediaExtension(contentType, fileName))
	_, err := storage.client.PutObject(ctx, storage.bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return storage.publicBaseURL + "/" + objectName, nil
}

func (storage *PostMediaStorage) ensureBucket(ctx context.Context) error {
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

func mediaExtension(contentType string, _ string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}

func randomObjectID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "media"
	}

	return hex.EncodeToString(buffer)
}
