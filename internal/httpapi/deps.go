package httpapi

import "github.com/jmoiron/sqlx"

var db *sqlx.DB

func WithDB(d *sqlx.DB) { db = d }
