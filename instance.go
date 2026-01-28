package goSqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pardnchiu/go-sqlite/core"
)

var (
	conn *core.Connector
	once sync.Once
)

func New(c core.Config) (*core.Connector, error) {
	once.Do(func() {
		conn = &core.Connector{Map: make(map[string]*sql.DB)}
	})

	conn.Mu.Lock()
	defer conn.Mu.Unlock()

	if conn.Map == nil {
		conn.Map = make(map[string]*sql.DB)
	}

	// get {dbName}.db form path
	if c.Key == "" {
		filename := filepath.Base(c.Path)
		c.Key = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	if conn.Map[c.Key] != nil {
		return conn, nil
	}

	db, err := sql.Open("sqlite3", c.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	db.Exec("PRAGMA journal_mode=WAL")
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(2)

	if c.Lifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(c.Lifetime) * time.Second)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	conn.Map[c.Key] = db
	return conn, nil
}
