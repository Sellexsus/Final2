package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "final/internal/db"
)

func ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if db == nil {
		jsonError(w, http.StatusInternalServerError, "db not initialized")
		return
	}

	q := r.URL.Query()
	from := strings.TrimSpace(q.Get("from"))
	if from == "" {
		from = time.Now().UTC().Format("20060102")
	} else if _, err := time.Parse("20060102", from); err != nil {
		jsonError(w, http.StatusBadRequest, "bad 'from' date")
		return
	}

	limit := 20
	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		} else {
			jsonError(w, http.StatusBadRequest, "bad 'limit'")
			return
		}
	}
	search := strings.TrimSpace(q.Get("search"))

	rows, err := dbpkg.ListTasks(r.Context(), db, from, limit, search)
	if err != nil {
		log.Printf("ListTasksHandler: db error: %v", err)
		jsonError(w, http.StatusInternalServerError, "db error")
		return
	}

	out := map[string][]map[string]string{"tasks": {}}
	for _, t := range rows {
		out["tasks"] = append(out["tasks"], map[string]string{
			"id":      strconv.FormatInt(t.ID, 10),
			"date":    t.Date,
			"title":   t.Title,
			"comment": t.Comment,
			"repeat":  t.Repeat,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Printf("ListTasksHandler: encode response error: %v", err)
	}
}
