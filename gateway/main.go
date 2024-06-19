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
	url := "http://localhost:8080/api/news/" + mux.Vars(r)["count"]
	forwardRequest(w, r, url)
}

func addCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Создание запроса на цензурирование
	censorshipReq, err := http.NewRequest("POST", "http://localhost:8081/censorship", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Failed to create censorship request", http.StatusInternalServerError)
		return
	}
	censorshipReq.Header = r.Header

	// Отправка запроса на цензурирование
	censorshipResp, err := http.DefaultClient.Do(censorshipReq)
	if err != nil {
		http.Error(w, "Failed to forward censorship request", http.StatusInternalServerError)
		return
	}
	defer censorshipResp.Body.Close()

	// Проверка ответа от цензуры
	if censorshipResp.StatusCode != http.StatusOK {
		http.Error(w, "Comment contains banned words", censorshipResp.StatusCode)
		return
	}

	// Если прошло цензурирование, отправляем запрос на добавление комментария
	commentsReq, err := http.NewRequest("POST", "http://localhost:8081/api/comments", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Failed to create comments request", http.StatusInternalServerError)
		return
	}
	commentsReq.Header = r.Header

	// Отправка запроса на добавление комментария
	commentsResp, err := http.DefaultClient.Do(commentsReq)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}
	defer commentsResp.Body.Close()

	// Копирование ответа от сервиса комментариев в ответ клиенту
	w.WriteHeader(commentsResp.StatusCode)
	io.Copy(w, commentsResp.Body)
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
