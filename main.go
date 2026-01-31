package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pvavrina/goml/internal/storage"
	pb "github.com/pvavrina/goml/proto" 
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Define a context for the overall operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to the gRPC MlsolidService (Python service)
	// Address should be the service name in your Kubernetes cluster
	conn, err := grpc.Dial("mlsolid-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC service: %v", err)
	}
	defer conn.Close()

	client := pb.NewMlsolidServiceClient(conn)

	// Execute the data processing and storage logic
	runDataPipeline(ctx, client)
}

func runDataPipeline(ctx context.Context, client pb.MlsolidServiceClient) {
	// Initialize the Garage S3 storage client
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Garage storage: %v", err)
	}

	// Request experiment IDs from the gRPC service
	resp, err := client.Experiments(ctx, &pb.ExperimentsRequest{})
	if err != nil {
		log.Fatalf("Error calling gRPC Experiments: %v", err)
	}

	// Map response data to a JSON-friendly structure
	payload := struct {
		ExpIDs    []string `json:"exp_ids"`
		Count     int      `json:"count"`
		Timestamp int64    `json:"timestamp"`
	}{
		ExpIDs:    resp.GetExpIds(),
		Count:     len(resp.GetExpIds()),
		Timestamp: time.Now().Unix(),
	}

	// Serialize the payload to JSON format
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Failed to marshal experiment data: %v", err)
	}

	// Create a unique key using the current timestamp
	objectKey := fmt.Sprintf("exports/experiments_%d.json", time.Now().Unix())

	// Store the JSON file in the Garage bucket
	err = storageClient.SaveObject(ctx, objectKey, jsonData)
	if err != nil {
		log.Fatalf("Failed to upload object to Garage: %v", err)
	}

	fmt.Printf("Pipeline successful: Saved %d IDs to Garage at %s\n", payload.Count, objectKey)
}
