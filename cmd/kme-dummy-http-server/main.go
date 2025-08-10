package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type KeyResponse struct {
	Keys []Key `json:"keys"`
}

type Key struct {
	Key   string `json:"key"`
	KeyID string `json:"key_ID"`
}

func keysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := KeyResponse{
		Keys: []Key{
			{
				Key:   "NiXDkmgcAztCFzyhO8XI+COj1Y1pEMDR8H0LzxZxoFo=",
				KeyID: "3db4bb3f-0f51-49af-9531-7fb9bc08d1e0",
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
	http.HandleFunc("/api/v1/keys/CONS_TIM_UPT/dec_keys", keysHandler)
	http.HandleFunc("/api/v1/keys/CONS_TIM_UPT/enc_keys", keysHandler)
	log.Println("Server is listening on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
