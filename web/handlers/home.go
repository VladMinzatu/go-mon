package handlers

import (
	"html/template"
	"net/http"
)

type HomepageHandler struct {
	template *template.Template
}

func NewHomepageHandler(tmpl *template.Template) *HomepageHandler {
	return &HomepageHandler{
		template: tmpl,
	}
}

func (h *HomepageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.template.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
