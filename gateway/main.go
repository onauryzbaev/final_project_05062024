package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var censorshipServiceURL = "http://localhost:8081/censorship"

type Comment struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func createComment(w http.ResponseWriter, r *http.Request) {
	var comment Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Forward comment to censorship service
	censorshipReqBody, _ := json.Marshal(comment)
	resp, err := http.Post(censorshipServiceURL, "application/json", bytes.NewBuffer(censorshipReqBody))
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Comment failed censorship", http.StatusBadRequest)
		return
	}

	// Comment passed censorship, process further
	// Here you can add logic to store the comment in your comment storage service

	w.WriteHeader(http.StatusOK)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/comments", createComment).Methods("POST")

	log.Println("API Gateway running on port 8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}
