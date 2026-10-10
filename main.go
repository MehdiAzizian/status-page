package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
	"encoding/json"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Component struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			http.Error(w, "db down", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok\n"))
	})

mux.HandleFunc("GET /components", func(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(r.Context(), "SELECT id, name, status FROM components ORDER BY id")
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []Component{}
	for rows.Next() {
		var c Component
		rows.Scan(&c.ID, &c.Name, &c.Status)
		list = append(list, c)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
})

mux.HandleFunc("GET /components/{id}", func(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be a number", http.StatusBadRequest)
		return
	}
	var c Component
	err = db.QueryRow(r.Context(),
		"SELECT id, name, status FROM components WHERE id = $1", id,
	).Scan(&c.ID, &c.Name, &c.Status)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
})

	log.Println("listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
