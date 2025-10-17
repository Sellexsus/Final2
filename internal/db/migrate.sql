CREATE TABLE IF NOT EXISTS scheduler (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,
  date    TEXT    NOT NULL,
  title   TEXT    NOT NULL,
  comment TEXT    NOT NULL DEFAULT '',
  repeat  TEXT    NOT NULL DEFAULT ''
);
