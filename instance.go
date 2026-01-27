package goSqlite

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Key      string `json:"key,omitempty"`
	Path     string `json:"path"`
	Lifetime int    `json:"lifetime,omitempty"` // sec
}

type Database struct {
	db map[string]*sql.DB
	mu sync.Mutex
}

var (
	database *Database
)

func main() {
	database, db, err := New(Config{
		Key:      "test",
		Path:     "./data.db",
		Lifetime: 30,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	slog.Info("connected to database", slog.String("driver", "sqlite3"), slog.Any("datasource", db))

	err = NewBuilder(db).
		Table("users").
		Create(
			Column{
				Name:         "id",
				Type:         "INTEGER",
				IsPrimary:    true,
				AutoIncrease: true,
			},
			Column{
				Name:       "name",
				Type:       "TEXT",
				IsNullable: false,
			},
			Column{
				Name:       "email",
				Type:       "TEXT",
				IsNullable: false,
				Default:    "",
			},
		)
	if err != nil {
		log.Fatal(err)
	}

	err = NewBuilder(db).
		Table("users").
		Insert(map[string]any{
			"name":  "test",
			"email": "dev@pardn.io",
		})
	if err != nil {
		log.Fatal(err)
	}
}

func New(c Config) (*Database, *sql.DB, error) {
	if database == nil {
		database = &Database{db: make(map[string]*sql.DB)}
	}

	database.mu.Lock()
	defer database.mu.Unlock()

	// get {dbName}.db form path
	if c.Key == "" {
		filename := filepath.Base(c.Path)
		c.Key = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	if database.db[c.Key] != nil {
		return database, database.db[c.Key], nil
	}

	db, err := sql.Open("sqlite3", c.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if c.Lifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(c.Lifetime) * time.Second)
	}

	database.db[c.Key] = db
	return database, db, nil
}

func (d *Database) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, db := range d.db {
		err := db.Close()
		if err == nil {
			continue
		}
		slog.Error("failed to close db",
			slog.String("db", key),
			slog.Any("error", err))
	}
	d.db = nil
}
