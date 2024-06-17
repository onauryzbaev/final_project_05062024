package comments

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

type Comment struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Censored  bool      `json:"censored"`
}

var db *sql.DB

// Инициализация базы данных
func InitDB(databasePath string) {
	var err error
	db, err = sql.Open("sqlite", databasePath)
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS comments (
        "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        "content" TEXT,
        "created_at" DATETIME,
        "censored" BOOLEAN
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

// Добавление комментария
func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	var comment Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	comment.CreatedAt = time.Now()
	comment.Censored = false // Изначально комментарий не цензурирован

	_, err := db.Exec("INSERT INTO comments (content, created_at, censored) VALUES (?, ?, ?)", comment.Content, comment.CreatedAt, comment.Censored)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Status: "success", Message: "Comment added"})
}

// Удаление комментария
func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := db.Exec("DELETE FROM comments WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Status: "success", Message: "Comment deleted"})
}
