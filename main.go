package main

import (
	"html/template"
	"log"
	"log/slog"
	"net/http"

	"github.com/VladMinzatu/go-mon/web/handlers"
)

func main() {
	tmpl := template.Must(template.ParseFiles("index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		type data struct {
			Heading string
		}
		tmpl.Execute(w, data{Heading: "Heading is templated"})
	})
	http.HandleFunc("/ws", handlers.ServeWs) // test with: websocat ws://localhost:8080/ws
	slog.Info("Starting server on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
