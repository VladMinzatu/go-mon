package handlers

import (
	"html/template"
	"net/http"
)

var homeTmpl = template.Must(template.New("index.html").ParseFiles("web/views/index.html"))

func ServeHomepage(w http.ResponseWriter, r *http.Request) {
	err := homeTmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
