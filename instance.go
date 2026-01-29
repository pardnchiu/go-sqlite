package goSqlite

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pardnchiu/go-sqlite/core"
)

func New(c core.Config) (*core.Connector, error) {
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 50
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 25
	}

	write, err := sql.Open("sqlite3",
		fmt.Sprintf("file:%s?"+
			"cache=shared"+
			"&mode=rwc"+
			"&_journal_mode=WAL"+
			"&_busy_timeout=15000"+
			"&_synchronous=NORMAL"+
			"&_txlock=immediate", c.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to open write db: %w", err)
	}

	write.SetMaxOpenConns(1)
	write.SetMaxIdleConns(1)
	write.SetConnMaxLifetime(0)

	for _, e := range []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -131072",
		"PRAGMA mmap_size = 1073741824",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA wal_autocheckpoint = 0",
		"PRAGMA journal_size_limit = 268435456",
		"PRAGMA busy_timeout = 15000",
		"PRAGMA foreign_keys = ON",
	} {
		if _, err := write.Exec(e); err != nil {
			write.Close()
			return nil, fmt.Errorf("failed to setup pragma: %w", err)
		}
	}

	if err := write.Ping(); err != nil {
		write.Close()
		return nil, fmt.Errorf("failed to ping write db: %w", err)
	}

	read, err := sql.Open("sqlite3",
		fmt.Sprintf("file:%s?"+
			"cache=shared"+
			"&mode=ro"+ // read-only
			"&_query_only=1"+
			"&_journal_mode=WAL"+
			"&_busy_timeout=5000", c.Path))
	if err != nil {
		write.Close()
		return nil, fmt.Errorf("failed to open read db: %w", err)
	}

	read.SetMaxOpenConns(c.MaxOpenConns)
	read.SetMaxIdleConns(c.MaxIdleConns)

	if c.Lifetime > 0 {
		read.SetConnMaxLifetime(time.Duration(c.Lifetime) * time.Second)
	} else {
		read.SetConnMaxLifetime(2 * time.Minute)
	}

	if err := read.Ping(); err != nil {
		write.Close()
		read.Close()
		return nil, fmt.Errorf("failed to ping read db: %w", err)
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := write.Exec("PRAGMA wal_checkpoint(PASSIVE)"); err != nil {
				fmt.Printf("checkpoint failed: %v\n", err)
			}
		}
	}()

	return &core.Connector{
		Read:  core.NewBuilder(read),
		Write: core.NewBuilder(write),
	}, nil
}
