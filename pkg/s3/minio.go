package s3

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"path/filepath"
	"strings"
)

// параметры подключения к minio
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
}

// обертка над клиентом minio
type Client struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

// инициализация клиента minio, создание багета при необходимости
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	c := &Client{
		client:     minioClient,
		bucketName: cfg.BucketName,
		endpoint:   cfg.Endpoint,
		useSSL:     cfg.UseSSL,
	}

	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, cfg.BucketName)

		if err := minioClient.SetBucketPolicy(ctx, cfg.BucketName, policy); err != nil {
			return nil, fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}
	return c, nil
}

// загружает файл в хранилище и возвращает url
func (c *Client) UploadImage(ctx context.Context, originalFileName string, fileData []byte, contentType string) (string, error) {
	//генерация уникального имени файла
	ext := strings.ToLower(filepath.Ext(originalFileName))
	if ext == "" {
		ext = ".jpg"
	}
	objectName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	//потоковая загрузка байтов в бакет
	reader := bytes.NewReader(fileData)
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, int64(len(fileData)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("failed to upload object to minio: %w", err)
	}

	//формирование url к файлу
	scheme := "http"
	if c.useSSL {
		scheme = "https"
	}
	fileURL := fmt.Sprintf("%s://%s/%s/%s", scheme, c.endpoint, c.bucketName, objectName)
	return fileURL, nil
}

func (c *Client) DeleteImage(ctx context.Context, objectName string) error {
	//если передался полный url, отсекается префикс и достается только имя
	if strings.Contains(objectName, "/") {
		parts := strings.Split(objectName, "/")
		objectName = parts[len(parts)-1]
	}

	err := c.client.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object from minio: %w", err)
	}
	return nil
}
