package s3store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Presigner выдаёт временные GET-URL объектов в бакете.
type Presigner struct {
	api    *s3.PresignClient
	bucket string
	ttl    time.Duration
}

// NewPresigner создаёт клиент presign (MinIO-совместимый endpoint).
func NewPresigner(ctx context.Context, endpoint, region, bucket, accessKey, secretKey string, ttl time.Duration) (*Presigner, error) {
	endpoint = strings.TrimRight(endpoint, "/")
	if region == "" {
		region = "us-east-1"
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
	return &Presigner{
		api:    s3.NewPresignClient(client),
		bucket: bucket,
		ttl:    ttl,
	}, nil
}

// PresignGetURL временная ссылка на скачивание объекта.
func (p *Presigner) PresignGetURL(ctx context.Context, key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("empty s3 key")
	}
	out, err := p.api.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(p.ttl))
	if err != nil {
		return "", fmt.Errorf("presign get %q: %w", key, err)
	}
	return out.URL, nil
}
