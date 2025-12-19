package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	// Read credentials from environment variables
	accessKey := os.Getenv("SA_ACCESS_KEY")
	secretKey := os.Getenv("SA_SECRET_KEY")

	const (
		endpoint = "http://localhost:3900"
		bucket   = "goml-data"
		region   = "garage"
	)

	// Validate environment variables
	if accessKey == "" || secretKey == "" {
		log.Fatal("Error: SA_ACCESS_KEY or SA_SECRET_KEY environment variables are not set.")
	}

	// Configure a custom resolver to map S3 calls to your local Garage instance
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,   // Essential for Garage path-style addressing
			SigningRegion:     region, // Forces the signature to use 'garage'
		}, nil
	})

	// Load the AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize S3 client
	client := s3.NewFromConfig(cfg)

	// Test upload to your 'goml-data' bucket
	fmt.Println("⏳ Attempting to upload to Garage...")
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("test-go-ml.txt"),
		Body:   strings.NewReader("Hello from Go! This is a test for the GoML project."),
	})

	if err != nil {
		log.Fatalf("❌ Upload failed: %v", err)
	}

	fmt.Println("🚀 Success! File 'test-go-ml.txt' is now stored in Garage.")
}
