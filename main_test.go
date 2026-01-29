package goSqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pardnchiu/go-sqlite/core"
)

func TestNew(t *testing.T) {
	t.Run("create new database", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		conn, err := New(core.Config{
			Path:     dbPath,
			Lifetime: 30,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if conn == nil {
			t.Fatal("expected connector instance, got nil")
		}
		defer conn.Close()

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Fatalf("database file not created at %s", dbPath)
		}

		if conn.Read == nil {
			t.Error("expected read connection, got nil")
		}
		if conn.Write == nil {
			t.Error("expected write connection, got nil")
		}
	})

	t.Run("with custom connection settings", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		conn, err := New(core.Config{
			Path:         dbPath,
			MaxOpenConns: 10,
			MaxIdleConns: 5,
			Lifetime:     60,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer conn.Close()
	})

	t.Run("default values applied", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		conn, err := New(core.Config{
			Path: dbPath,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer conn.Close()
	})

	t.Run("invalid path", func(t *testing.T) {
		conn, err := New(core.Config{
			Path: "/nonexistent/path/to/db.db",
		})
		if err == nil {
			if conn != nil {
				conn.Close()
			}
			t.Fatal("expected error for invalid path")
		}
	})

	t.Run("with zero lifetime uses default", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test.db")
		conn, err := New(core.Config{
			Path:     dbPath,
			Lifetime: 0,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer conn.Close()
	})
}

func TestNewIntegration(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "integration.db")
	conn, err := New(core.Config{
		Path:     dbPath,
		Lifetime: 30,
	})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer conn.Close()

	t.Run("create table via Write", func(t *testing.T) {
		err := conn.Write.Table("users").Create(
			core.Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
			core.Column{Name: "name", Type: "TEXT"},
		)
		if err != nil {
			t.Fatalf("failed to create table: %v", err)
		}
	})

	t.Run("insert via Write", func(t *testing.T) {
		id, err := conn.Write.Table("users").Insert(map[string]any{"name": "Alice"})
		if err != nil {
			t.Fatalf("insert failed: %v", err)
		}
		if id < 1 {
			t.Errorf("expected valid id, got %d", id)
		}
	})

	t.Run("read via Read", func(t *testing.T) {
		count, err := conn.Read.Table("users").Count()
		if err != nil {
			t.Fatalf("count failed: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1, got %d", count)
		}
	})

	t.Run("Query method", func(t *testing.T) {
		rows, err := conn.Query("", "SELECT * FROM users")
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("expected at least one row")
		}
	})

	t.Run("Exec method", func(t *testing.T) {
		_, err := conn.Exec("", "UPDATE users SET name = 'Bob' WHERE id = 1")
		if err != nil {
			t.Fatalf("exec failed: %v", err)
		}
	})

	t.Run("Read builder access", func(t *testing.T) {
		builder := conn.Read
		if builder == nil {
			t.Fatal("expected builder")
		}
		if builder.Raw() != conn.Read.DB {
			t.Error("expected read connection")
		}
	})
}
