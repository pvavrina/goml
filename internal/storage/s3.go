package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	S3Client *s3.Client
	Bucket   string
}

// NewClient initializes a new S3 client configured for Garage
func NewClient(ctx context.Context) (*Client, error) {
	accessKey := os.Getenv("SA_ACCESS_KEY")
	secretKey := os.Getenv("SA_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	endpoint := os.Getenv("S3_ENDPOINT")
	region := "garage"

	if bucket == "" {
		bucket = "goml-data"
	}
	if endpoint == "" {
		endpoint = "http://garage-s3.garage.svc.cluster.local:3900"
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,
			SigningRegion:     region,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	return &Client{
		S3Client: s3.NewFromConfig(cfg),
		Bucket:   bucket,
	}, nil
}
