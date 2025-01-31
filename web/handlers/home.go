package handlers

import (
	"html/template"
	"net/http"
)

var tmpl = template.Must(template.ParseFiles("index.html"))

func ServeHomepage(w http.ResponseWriter, r *http.Request) {
	type data struct {
		Heading string
	}
	tmpl.Execute(w, data{Heading: "Heading is templated"})
}
