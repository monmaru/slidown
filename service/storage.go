package service

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type Storage interface {
	Upload(ctx context.Context, rc io.ReadCloser, fileName string) (string, error)
}

type StorageService struct {
	bucket string
}

func NewStorage(bucket string) Storage {
	return &StorageService{bucket: bucket}
}

func (s *StorageService) Upload(ctx context.Context, rc io.ReadCloser, fileName string) (string, error) {
	defer rc.Close()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	w := client.Bucket(s.bucket).Object(fileName).NewWriter(ctx)
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.CacheControl = "no-cache"

	_, err = io.Copy(w, rc)
	if err != nil {
		return "", err
	}

	if err := w.Close(); err != nil {
		return "", err
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucket, fileName), nil
}
