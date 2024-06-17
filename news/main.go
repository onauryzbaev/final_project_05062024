package main

import (
	"database/sql"
	"log"
	"net/http"

	"yourmodule/comments" // Импортируйте пакет комментариев

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

// Инициализация базы данных
func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./rss.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS rss (
        "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        "title" TEXT,
        "description" TEXT,
        "link" TEXT,
        "pubDate" DATETIME
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация базы данных комментариев
	comments.InitDB("./comments.db")
}

// Добавление API для комментариев
func setupCommentsAPI(r *mux.Router) {
	r.HandleFunc("/api/comments", comments.AddCommentHandler).Methods("POST")
	r.HandleFunc("/api/comments/{id}", comments.DeleteCommentHandler).Methods("DELETE")
}

func main() {
	// Чтение конфигурационного файла
	config, err := readConfig("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Инициализация базы данных
	initDB()
	defer db.Close()

	// Запуск периодического обхода RSS-лент
	go pollFeeds(config)

	// Настройка маршрутов HTTP
	r := mux.NewRouter()
	r.HandleFunc("/api/news/{count}", apiHandler).Methods("GET")

	// Настройка API для комментариев
	setupCommentsAPI(r)

	// Настройка статических файлов
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/").Handler(fs)

	// Запуск сервера
	log.Fatal(http.ListenAndServe(":8080", r))
}
