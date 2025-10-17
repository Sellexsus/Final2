package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	sched "final/internal/schedule"

	"github.com/jmoiron/sqlx"
)

type addTaskRequest struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
type addTaskResponse struct {
	ID int64 `json:"id"`
}

var (
	ymdRe  = regexp.MustCompile(`^\d{8}$`)
	ymdFmt = "20060102"
)

func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if db == nil {
		jsonError(w, http.StatusInternalServerError, "db not initialized")
		return
	}

	var req addTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("AddTaskHandler: decode error: %v", err)
		jsonError(w, http.StatusBadRequest, "bad json")
		return
	}

	today := time.Now().Format(ymdFmt)
	req.Date = strings.TrimSpace(req.Date)
	if req.Date == "" {
		req.Date = today
	} else if len(req.Date) == 8 {

		if d, err := time.Parse(ymdFmt, req.Date); err == nil {
			if d.Format(ymdFmt) < today {
				req.Date = today
			}
		} else {
			log.Printf("AddTaskHandler: parse date error: %v", err)
		}
	}

	if err := validateAddTask(req); err != nil {
		log.Printf("AddTaskHandler: validation error: %v", err)
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := insertTask(r.Context(), db, req)
	if err != nil {
		log.Printf("AddTaskHandler: insertTask error: %v", err)
		jsonError(w, http.StatusInternalServerError, "db error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(addTaskResponse{ID: id}); err != nil {
		log.Printf("AddTaskHandler: encode response error: %v", err)
	}
}

func validateAddTask(req addTaskRequest) error {
	if !ymdRe.MatchString(req.Date) {
		log.Printf("validateAddTask: bad date format %q", req.Date)
		return errors.New("bad date format, want YYYYMMDD")
	}
	if _, err := time.Parse(ymdFmt, req.Date); err != nil {
		log.Printf("validateAddTask: invalid date %q: %v", req.Date, err)
		return errors.New("invalid date")
	}
	if strings.TrimSpace(req.Title) == "" {
		log.Printf("validateAddTask: empty title")
		return errors.New("title is required")
	}
	if req.Repeat != "" && req.Repeat != "y" {
		if _, err := sched.NextDate(req.Date, req.Repeat); err != nil {
			log.Printf("validateAddTask: bad repeat rule %q: %v", req.Repeat, err)
			return errors.New("bad repeat rule")
		}
	}
	return nil
}

func insertTask(ctx context.Context, dbx *sqlx.DB, req addTaskRequest) (int64, error) {
	res, err := dbx.ExecContext(ctx,
		`INSERT INTO scheduler(date, title, comment, repeat) VALUES(?, ?, ?, ?)`,
		req.Date, req.Title, req.Comment, req.Repeat,
	)
	if err != nil {
		log.Printf("insertTask: DB insert error: %v", err)
		return 0, err
	}
	return res.LastInsertId()
}
