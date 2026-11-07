package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client *s3.Client
	bucket string
	region string
}

func NewS3Storage(ctx context.Context, region, bucket string) (*S3Storage, error) {
	if region == "" || bucket == "" {
		return nil, fmt.Errorf("AWS_REGION and AWS_S3_BUCKET are required")
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}

	return &S3Storage{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
		region: region,
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("upload object to S3: %w", err)
	}
	return nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object from S3: %w", err)
	}
	return nil
}

// PublicURL returns a stable S3 object URL. The bucket policy must allow
// s3:GetObject for the room-bills/* prefix.
func (s *S3Storage) ObjectKeyFromPublicURL(publicURL string) (string, error) {
	parsed, err := url.Parse(publicURL)
	if err != nil {
		return "", fmt.Errorf("parse receipt image URL: %w", err)
	}
	expectedHost := fmt.Sprintf("%s.s3.%s.amazonaws.com", s.bucket, s.region)
	if parsed.Scheme != "https" || parsed.Host != expectedHost || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("receipt image URL does not belong to configured S3 bucket")
	}
	key, err := url.PathUnescape(strings.TrimPrefix(parsed.EscapedPath(), "/"))
	if err != nil || key == "" {
		return "", fmt.Errorf("invalid S3 object key in receipt image URL")
	}
	return key, nil
}

func (s *S3Storage) PublicURL(key string) string {
	objectURL := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s.s3.%s.amazonaws.com", s.bucket, s.region),
		Path:   "/" + key,
	}
	return objectURL.String()
}
