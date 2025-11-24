package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Handler for the root endpoint
func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received on path: %s", r.URL.Path)

	// Simulates a Health Check response
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "ML Service Go (SLES) est en cours d'exécution. Statut: OK!")
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

// Go service code (main.go)
