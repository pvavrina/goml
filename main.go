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

// Handler for the root endpoint (Used for health checks)
func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received on path: %s", r.URL.Path)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "ML Service Go (SLES) is running. Status: OK!")
}

// Handler for the /predict endpoint. Triggers gRPC and saves result to S3.
func predictHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: /predict. Attempting gRPC call.")
	
	// 1. Setup gRPC connection
	mlSolidAddr := os.Getenv("MLSOLID_SERVICE_ADDR")
	if mlSolidAddr == "" {
		mlSolidAddr = "mlsolid-service:5000"
	}

	conn, err := grpc.NewClient(mlSolidAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("FATAL: Failed to connect to MLSolid: %v", err)
		http.Error(w, "gRPC Connection Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := mlservice.NewMlsolidServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// 2. Perform the gRPC call
	req := &mlservice.ExperimentsRequest{}	
	resp, err := client.Experiments(ctx, req)
	if err != nil {
		log.Printf("gRPC Call Error: %v", err)
		http.Error(w, "gRPC Call Failed", http.StatusServiceUnavailable)
		return
	}

	// 3. PERSISTENCE: Save the result to Garage (S3)
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("prediction-%s.json", timestamp)
	
	// Prepare the data to save
	payload := []byte(fmt.Sprintf("%+v", resp))

	err = storageClient.SaveObject(ctx, filename, payload)
	if err != nil {
		log.Printf("Warning: Failed to persist data to Garage: %v", err)
		// We continue to serve the request even if storage fails
	} else {
		log.Printf("✅ Result successfully persisted to Garage as: %s", filename)
	}

	// 4. Success response to client
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status": "SUCCESS", "stored_file": "%s", "data": "%+v"}`, filename, resp)
}

func main() {
	// Initialize S3 Storage Client
	ctx := context.Background()
	var err error
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Critical: Could not initialize storage client: %v", err)
	}
	log.Println("✅ Storage client (Garage) initialized successfully")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ROUTING
	http.HandleFunc("/", handler)
	http.HandleFunc("/predict", predictHandler)

	log.Printf("Starting Go service on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
