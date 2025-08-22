package httpapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thom151/fif/internal/database"
)

func DbBrollToSignedBroll(broll database.Broll, s3client *s3.Client) (database.Broll, error) {
	if broll.S3Url.String == "" {
		return database.Broll{}, nil
	}
	parts := strings.Split(broll.S3Url.String, ",")
	if len(parts) < 2 {
		return database.Broll{}, nil
	}
	bucket := parts[0]
	key := parts[1]
	presigned, err := GeneratePresignedURL(s3client, bucket, key, 5*time.Minute)
	if err != nil {
		return database.Broll{}, err
	}
	broll.S3Url.String = presigned
	return broll, nil
}

func GeneratePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	presignedUrl, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}
	return presignedUrl.URL, nil
}
