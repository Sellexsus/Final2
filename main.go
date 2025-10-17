package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	dbpkg "final/internal/db"
	api "final/internal/httpapi"
)

func getPort() int {
	const def = 7540
	if v := os.Getenv("TODO_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 && p <= 65535 {
			return p
		}
	}
	return def
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	clean := path.Clean("/" + r.URL.Path)
	if clean == "/" {
		clean = "/index.html"
	}
	rel := strings.TrimPrefix(clean, "/")
	file := filepath.Join("web", rel)

	baseAbs, _ := filepath.Abs("web")
	fileAbs, _ := filepath.Abs(file)
	sep := string(os.PathSeparator)
	if fileAbs != baseAbs && !strings.HasPrefix(fileAbs, baseAbs+sep) {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(file)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if info.IsDir() {
		file = filepath.Join(file, "index.html")
		if info2, err2 := os.Stat(file); err2 != nil || info2.IsDir() {
			http.NotFound(w, r)
			return
		}
	}
	http.ServeFile(w, r, file)
}

func main() {
	db, err := dbpkg.Open()
	if err != nil {
		log.Fatal(err)
	}
	api.WithDB(db)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/nextdate", api.NextDateHandler)

	mux.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.AddTaskHandler(w, r) //
			return
		}
		api.TaskItemHandler(w, r)
	})

	mux.HandleFunc("/api/task/done", api.TaskItemHandler)
	mux.HandleFunc("/api/tasks", api.ListTasksHandler)
	mux.HandleFunc("/", staticHandler)

	addr := fmt.Sprintf(":%d", getPort())
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
