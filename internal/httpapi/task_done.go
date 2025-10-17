package httpapi

import (
	"net/http"
	"time"

	sched "final/internal/schedule"
)

func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if db == nil {
		jsonError(w, http.StatusInternalServerError, "db not initialized")
		return
	}

	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "bad id")
		return
	}

	var row struct{ Date, Repeat string }
	if err := db.Get(&row, `SELECT date, repeat FROM scheduler WHERE id=?`, id); err != nil {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}

	// если одноразовая задачка удалится
	if row.Repeat == "" {
		if _, err := db.Exec(`DELETE FROM scheduler WHERE id=?`, id); err != nil {
			jsonError(w, http.StatusInternalServerError, "db error")
			return
		}
		writeJSON(w, map[string]bool{"deleted": true})
		return
	}

	now := r.URL.Query().Get("now")
	if now == "" {
		now = time.Now().UTC().Format(ymdFmt)
	}

	next, err := sched.NextAfter(row.Date, row.Repeat, now)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := db.Exec(`UPDATE scheduler SET date=? WHERE id=?`, next, id); err != nil {
		jsonError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, map[string]string{"next": next})
}
