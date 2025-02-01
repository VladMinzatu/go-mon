package handlers

import (
	"html/template"
	"net/http"

	"github.com/VladMinzatu/go-mon/monitor"
)

var funcMap = template.FuncMap{
	"toGB": func(bytes uint64) float64 {
		return float64(bytes) / (1024 * 1024 * 1024)
	},
}

var tmpl = template.Must(template.New("index.html").Funcs(template.FuncMap(funcMap)).ParseFiles("web/views/index.html"))

func ServeHomepage(w http.ResponseWriter, r *http.Request) {
	stats := monitor.SystemStats{
		CPUUsagePerCore: []float64{0, 0, 0, 0},
		TotalMemory:     24000000000,
		UsedMemory:      12000000000,
		FreeMemory:      12000000000,
		MemoryUsage:     0.5,
	}
	err := tmpl.Funcs(funcMap).Execute(w, stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
