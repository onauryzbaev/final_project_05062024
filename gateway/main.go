package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Middleware func(http.Handler) http.Handler

func main() {
	r := mux.NewRouter()

	r.Use(loggingMiddleware)
	r.Use(requestIDMiddleware)

	r.HandleFunc("/api/news/{count}", newsHandler).Methods("GET")
	r.HandleFunc("/api/comments", addCommentHandler).Methods("POST")
	r.HandleFunc("/api/comments/{id}", deleteCommentHandler).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting API Gateway on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func newsHandler(w http.ResponseWriter, r *http.Request) {
	// Логика перенаправления запросов на сервис новостей
	// Пример запроса на получение новостей
}

func addCommentHandler(w http.ResponseWriter, r *http.Request) {
	forwardRequest(w, r, "http://localhost:8081/api/comments")
}

func deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	url := "http://localhost:8081/api/comments/" + id
	forwardRequest(w, r, url)
}

func forwardRequest(w http.ResponseWriter, r *http.Request, url string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest(r.Method, url, bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			r.Header.Set("X-Request-ID", requestID)
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func generateRequestID() string {
	// Генерация уникального идентификатора
	return "some-unique-id"
}
