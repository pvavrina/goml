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

	// Alias the generated gRPC stubs package
	mlservice "github.com/pvavrina/goml/api/mlservice"
)

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
		mlSolidAddr = "mlsolid-service:5000" // Fallback, but should be set by K8s
	}

	// 2. Setup gRPC connection
	// Use insecure credentials because it is internal K8s communication
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
	}() // Close the connection when the handler exits
	// 3. Create gRPC client and context
	// Using the validated function name: NewMlsolidServiceClient (Action 374)
	client := mlservice.NewMlsolidServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) // 5 seconds timeout
	defer cancel()                                                          // Release resources tied to the context

	// 4. Perform the gRPC call (using the validated TaggedModelRequest structure)
	// Correction: Use TaggedModelRequest with actual fields (Name, Tag) (Action 385)
	req := &mlservice.TaggedModelRequest{
		Name: "iris-model", // Placeholder data, assuming you want a specific model
		Tag:  "v1",
	}

	// Correction: Use the validated method name TaggedModel (Action 384)
	resp, err := client.TaggedModel(ctx, req)
	if err != nil {
		log.Printf("gRPC Call Error: Failed to get prediction: %v", err)
		http.Error(w, fmt.Sprintf("gRPC Call Error: %v", err), http.StatusServiceUnavailable)
		return
	}

	// 5. Success response
	// Correction: Print the entire response object (using %+v) since GetPredictionValue is undefined (Action 385)
	log.Printf("gRPC Call SUCCESS. Received response object: %+v", resp)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseBody := fmt.Sprintf(`{"status": "SUCCESS", "response_data": "%+v"}`, resp)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(responseBody))
}

func main() {
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
