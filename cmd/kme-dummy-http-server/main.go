package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type KeyResponse struct {
	Keys []Key `json:"keys"`
}

type Key struct {
	Key   string `json:"key"`
	KeyID string `json:"key_ID"`
}

func keysHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s\n", r.Method, r.URL.Path)
	log.Printf("++ User-Agent: %s\n", r.Header.Get("User-Agent"))
	log.Printf("++ Accept-Encoding: %s\n", r.Header.Get("Accept-Encoding"))
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//sae := chi.URLParam(r, "sae")

	keyID := uuid.New().String()
	key := make([]byte, 32)

	if _, err := rand.Read(key); err != nil {
		http.Error(w, "Failed to generate key", http.StatusInternalServerError)
		return
	}
	b64key := base64.StdEncoding.EncodeToString(key)
	response := KeyResponse{
		Keys: []Key{
			{
				Key:   b64key,
				KeyID: keyID,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}

func main() {
	hs := chi.NewRouter()
	hs.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Dummy KME HTTP API Service v0.1")
	})
	hs.Post("/api/v1/keys/{sae}/dec_keys", keysHandler)
	hs.Post("/api/v1/keys/{sae}/enc_keys", keysHandler)
	log.Println("Server is listening on :8080...")
	if err := http.ListenAndServe(":8080", hs); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
