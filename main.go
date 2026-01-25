package main

import (
	"database/sql"
	"log"
	"log/slog"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(30 * time.Minute)

	slog.Info("connected to database", slog.String("driver", "sqlite3"), slog.String("datasource", "./data.db"))
}
