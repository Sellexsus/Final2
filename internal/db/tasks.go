package db

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Task struct {
	ID      int64  `db:"id"`
	Date    string `db:"date"`
	Title   string `db:"title"`
	Comment string `db:"comment"`
	Repeat  string `db:"repeat"`
}

func ListTasks(ctx context.Context, db *sqlx.DB, from string, limit int, search string) ([]Task, error) {
	sql := `
SELECT id, date, title, comment, repeat
FROM scheduler
WHERE date >= ?
`
	args := []any{from}
	if search != "" {
		sql += ` AND (title LIKE ? OR comment LIKE ?) `
		p := "%" + search + "%"
		args = append(args, p, p)
	}
	sql += ` ORDER BY date, id LIMIT ?`
	args = append(args, limit)

	var rows []Task
	if err := db.SelectContext(ctx, &rows, sql, args...); err != nil {
		return nil, err
	}
	return rows, nil
}
