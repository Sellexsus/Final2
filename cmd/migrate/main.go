package main

import (
	dbpkg "final/internal/db"
	"log"
)

func main() {
	db, err := dbpkg.Open() //АВТОМАТИЧЕСКИЙ ЗАПУСК БАЗЫ ДАННЫХ
	if err != nil {
		log.Fatal(err)
	}
	_ = db.Close()
}
