package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Import your internal storage package
	"github.com/pvavrina/goml/internal/storage"
	
	// Alias the generated gRPC stubs package
	mlservice "github.com/pvavrina/goml/api/mlservice"
)

var storageClient *storage.Client

// Handler for the root endpoint (Used by / and /ping for health checks)
func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received on path: %s", r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err := fmt.Fprintf(w, "ML Service Go (SLES) is running. Status: OK!")
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handler for the /predict endpoint. Triggers the gRPC call to MLSolid.
func predictHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received on path: %s. Attempting gRPC call to MLSolid.", r.URL.Path)
	
	// 1. Get the gRPC address from environment variable
	mlSolidAddr := os.Getenv("MLSOLID_SERVICE_ADDR")
	if mlSolidAddr == "" {
		mlSolidAddr = "mlsolid-service:5000"
	}

	// 2. Setup gRPC connection
	conn, err := grpc.NewClient(mlSolidAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("FATAL: Failed to connect to MLSolid gRPC service (%s): %v", mlSolidAddr, err)
		http.Error(w, fmt.Sprintf("gRPC Connection Error: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() {
		if cErr := conn.Close(); cErr != nil {
			log.Printf("Warning: Failed to close gRPC connection: %v", cErr)
		}
	}()

	// 3. Create gRPC client and context
	client := mlservice.NewMlsolidServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// 4. Perform the gRPC call
	req := &mlservice.ExperimentsRequest{}	
	resp, err := client.Experiments(ctx, req)
	
	if err != nil {
		log.Printf("gRPC Call Error: Failed to list experiments: %v", err)
		http.Error(w, fmt.Sprintf("gRPC Call Error: %v", err), http.StatusServiceUnavailable)
		return
	}

	// 5. Success response
	log.Printf("gRPC Call SUCCESS. Received response object: %+v", resp)
	
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseBody := fmt.Sprintf(`{"status": "SUCCESS", "response_data": "%+v"}`, resp)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(responseBody))
}

func main() {
	// Initialize S3 Storage Client
	ctx := context.Background()
	var err error
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize storage client: %v", err)
	}
	log.Println("✅ Storage client (Garage) initialized successfully")

	// Uses the PORT environment variable, defaults to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ROUTING: Map handlers to endpoints
	http.HandleFunc("/", handler)
	http.HandleFunc("/predict", predictHandler)

	log.Printf("Starting Go service on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
