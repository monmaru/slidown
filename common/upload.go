package common

import (
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

var bucketName string

func init() {
	bucketName = os.Getenv("BUCKETNAME")
}

// Upload2GCS ...
func Upload2GCS(ctx context.Context, rc io.ReadCloser, fileName string) (string, error) {
	defer rc.Close()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	w := client.Bucket(bucketName).Object(fileName).NewWriter(ctx)
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.CacheControl = "no-cache"

	_, err = io.Copy(w, rc)
	if err != nil {
		return "", err
	}

	if err := w.Close(); err != nil {
		return "", err
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, fileName), nil
}
