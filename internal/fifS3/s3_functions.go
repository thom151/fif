package fifS3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func DownloadAssetFromS3(ctx context.Context, url, bucket, fullPath string) (string, error) {
	key, err := GetKey(url)
	if err != nil {
		return "", err
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("create parent dir: %w", err)
	}

	// Create/truncate destination file
	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", fmt.Errorf("open dest file: %w", err)
	}
	defer f.Close()

	// Stream copy
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	// Optional: fsync to be extra safe
	if err := f.Sync(); err != nil {
		return "", fmt.Errorf("fsync: %w", err)
	}

	return fullPath, nil

}
