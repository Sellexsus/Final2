package httpapi

import (
	"encoding/json"
	"net/http"

	sched "final/internal/schedule"
)

type nextDateRequest struct {
	Date   string `json:"date"`
	Repeat string `json:"repeat"`
	Now    string `json:"now,omitempty"`
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		date := r.URL.Query().Get("date")
		rule := r.URL.Query().Get("repeat")
		now := r.URL.Query().Get("now")
		handleNextDatePlain(w, date, rule, now)
	case http.MethodPost:
		var req nextDateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		handleNextDatePlain(w, req.Date, req.Repeat, req.Now)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleNextDatePlain(w http.ResponseWriter, date, rule, now string) {
	next, err := sched.NextAfter(date, rule, now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(next))
}
