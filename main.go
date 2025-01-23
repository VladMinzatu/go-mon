package main

import (
	"log/slog"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	slog.Info("Starting server on port :8080")
	http.ListenAndServe(":8080", nil)
}
