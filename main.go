package main

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

type Builder struct {
	db         *sql.DB
	table      *string
	createList map[string]string
}

type Column struct {
	Name         string
	Type         string
	IsPrimary    bool
	IsNullable   bool
	AutoIncrease bool
	IsUnique     bool
	Default      any
	ForeignKey   string
}

type Foreign struct {
	Table  string
	Column string
}

func NewBuilder(db *sql.DB) *Builder {
	return &Builder{
		db:         db,
		createList: make(map[string]string),
	}
}

func (b *Builder) Table(name string) *Builder {
	b.table = &name
	return b
}

func (b *Builder) Create(columns ...Column) error {
	for _, e := range columns {
		b.createList[e.Name] = b.buildColumn(e)
	}
	if b.table == nil {
		return fmt.Errorf("table name is required")
	}
	if len(b.createList) == 0 {
		return fmt.Errorf("no columns defined")
	}

	var sb strings.Builder
	sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	sb.WriteString(*b.table)
	sb.WriteString(" (")

	i := 0
	for name, def := range b.createList {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(name)
		sb.WriteString(" ")
		sb.WriteString(def)
		i++
	}

	sb.WriteString(")")

	_, err := b.db.Exec(sb.String())
	return err
}

func (b *Builder) buildColumn(c Column) string {
	var parts []string
	parts = append(parts, c.Type)

	if c.IsPrimary {
		parts = append(parts, "PRIMARY KEY")
	}

	if c.AutoIncrease {
		parts = append(parts, "AUTOINCREMENT")
	}

	if c.IsUnique {
		parts = append(parts, "UNIQUE")
	}

	if !c.IsNullable {
		parts = append(parts, "NOT NULL")
	}

	if c.Default != nil {
		parts = append(parts, fmt.Sprintf("DEFAULT %v", b.formatValue(c.Default)))
	}

	if c.ForeignKey != "" {
		parts = append(parts, fmt.Sprintf("REFERENCES %s", c.ForeignKey))
	}

	return strings.Join(parts, " ")
}

func (b *Builder) formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("'%v'", val)
	}
}
