package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"time"
)

type Task struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("request method=%s path=%s", r.Method, r.URL.Path)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatal("Failed to create database pool: ", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	log.Println("Connected to PostgreSQL")
	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
    		log.Printf("request method=%s path=%s", r.Method, r.URL.Path)

    		if r.Method != http.MethodGet {
       		 w.Header().Set("Allow", http.MethodGet)
       		 http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        	return
    }

    rows, err := dbPool.Query(
        r.Context(),
        "SELECT id, title, description, status, created_at FROM tasks ORDER BY id",
    )
    if err != nil {
        log.Printf("failed to query tasks: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    tasks := make([]Task, 0)

    for rows.Next() {
        var task Task

        if err := rows.Scan(
            &task.ID,
            &task.Title,
            &task.Description,
            &task.Status,
            &task.CreatedAt,
        ); err != nil {
            log.Printf("failed to scan task: %v", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        tasks = append(tasks, task)
    }

    if err := rows.Err(); err != nil {
        log.Printf("failed to iterate tasks: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")

    if err := json.NewEncoder(w).Encode(tasks); err != nil {
        log.Printf("failed to encode tasks: %v", err)
    }
})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
