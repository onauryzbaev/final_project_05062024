package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

var bannedWords = []string{"badword1", "badword2", "badword3"}
var comments = make(map[string]string) // In-memory storage for comments
var mu sync.Mutex

type Comment struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func censorComment(w http.ResponseWriter, r *http.Request) {
	var comment Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, word := range bannedWords {
		if strings.Contains(comment.Content, word) {
			http.Error(w, "Comment contains banned words", http.StatusBadRequest)
			return
		}
	}

	mu.Lock()
	comments[comment.ID] = comment.Content
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/censorship", censorComment).Methods("POST")

	log.Println("Censorship service running on port 8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
