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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Component struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func validStatus(s string) bool {
	switch s {
	case "operational", "degraded", "down":
		return true
	}
	return false
}


var requests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "http_requests_total",
	Help: "HTTP requests by status code and method.",
}, []string{"code", "method"})

var duration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_request_duration_seconds",
	Help: "How long HTTP requests take.",
}, []string{"code", "method"})

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

mux.HandleFunc("POST /components/{id}/status", func(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be a number", http.StatusBadRequest)
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if !validStatus(body.Status) {
		http.Error(w, "status must be operational, degraded or down", http.StatusBadRequest)
		return
	}
	tag, err := db.Exec(r.Context(),
		"UPDATE components SET status = $1 WHERE id = $2", body.Status, id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
})

	mux.Handle("GET /metrics", promhttp.Handler())

	handler := promhttp.InstrumentHandlerDuration(duration,
		promhttp.InstrumentHandlerCounter(requests, mux))

	log.Println("listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, handler))

}
