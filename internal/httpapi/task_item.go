package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	sched "final/internal/schedule"

	"github.com/jmoiron/sqlx"
)

type updateTaskRequest struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Универсальный обработчик:
//  ПОДСКАЗОЧНИК
//	GET    /api/task?id=...
//	PUT    /api/task  (JSON с id и полями)
//	DELETE /api/task?id=...
//	POST   /api/task/done?id=...

func TaskItemHandler(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		jsonError(w, http.StatusInternalServerError, "db not initialized")
		return
	}

	if r.URL.Path == "/api/task/done" {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		id, err := parseID(r)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "bad id")
			return
		}
		handleDone(w, r, db, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		id, err := parseID(r)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "bad id")
			return
		}
		handleGet(w, r, db, id)
	case http.MethodPut:
		handleUpdate(w, r, db)
	case http.MethodDelete:
		id, err := parseID(r)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "bad id")
			return
		}
		handleDelete(w, r, db, id)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE, POST")
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleGet(w http.ResponseWriter, r *http.Request, dbx *sqlx.DB, id int64) {
	var t struct {
		ID      int64  `db:"id"`
		Date    string `db:"date"`
		Title   string `db:"title"`
		Comment string `db:"comment"`
		Repeat  string `db:"repeat"`
	}
	if err := dbx.Get(&t, `SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?`, id); err != nil {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":      strconv.FormatInt(t.ID, 10),
		"date":    t.Date,
		"title":   t.Title,
		"comment": t.Comment,
		"repeat":  t.Repeat,
	})
}

func handleUpdate(w http.ResponseWriter, r *http.Request, dbx *sqlx.DB) {
	var req struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "bad json")
		return
	}
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil || id <= 0 {
		jsonError(w, http.StatusBadRequest, "bad id")
		return
	}
	if len(req.Date) != 8 {
		jsonError(w, http.StatusBadRequest, "bad date")
		return
	}
	if _, err := time.Parse("20060102", req.Date); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid date")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		jsonError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.Repeat != "" && req.Repeat != "y" {
		if _, err := sched.NextDate(req.Date, req.Repeat); err != nil {
			jsonError(w, http.StatusBadRequest, "bad repeat rule")
			return
		}
	}
	res, err := dbx.Exec(`UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`,
		req.Date, req.Title, req.Comment, req.Repeat, id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "db error")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{})
}

func handleDelete(w http.ResponseWriter, r *http.Request, dbx *sqlx.DB, id int64) {
	res, err := dbx.Exec(`DELETE FROM scheduler WHERE id=?`, id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "db error")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{})
}

func handleDone(w http.ResponseWriter, r *http.Request, dbx *sqlx.DB, id int64) {
	var row struct {
		Date   string `db:"date"`
		Repeat string `db:"repeat"`
	}
	if err := dbx.Get(&row, `SELECT date, repeat FROM scheduler WHERE id=?`, id); err != nil {
		jsonError(w, http.StatusNotFound, "not found")
		return
	}
	now := strings.TrimSpace(r.URL.Query().Get("now"))
	if strings.TrimSpace(row.Repeat) == "" {
		if _, err := dbx.Exec(`DELETE FROM scheduler WHERE id=?`, id); err != nil {
			jsonError(w, http.StatusInternalServerError, "db error")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{})
		return
	}
	next, err := sched.NextAfter(row.Date, row.Repeat, now)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "bad repeat rule")
		return
	}
	if _, err := dbx.Exec(`UPDATE scheduler SET date=? WHERE id=?`, next, id); err != nil {
		jsonError(w, http.StatusInternalServerError, "db error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{})
}
