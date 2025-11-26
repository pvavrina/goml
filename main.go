package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Handler for the root endpoint
func handler(w http.ResponseWriter, r *http.Request) {
	// 1. Log the request
	log.Printf("Request received on path: %s", r.URL.Path)

	// 2. Set Content-Type and write the response
	// Force the Content-Type header to UTF-8
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Check the error return value from Fprintf (The linting fix)
	_, err := fmt.Fprintf(w, "ML Service Go (SLES) is running. Status: OK!")
	if err != nil {
		// Log the error if writing fails (optional, but good practice)
		log.Printf("Error writing response: %v", err)
	}
}

func main() {
	// Uses the PORT environment variable, defaults to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handler)

	log.Printf("Starting Go service on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
