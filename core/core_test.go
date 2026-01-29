package core

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	return db
}

func TestNewBuilder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	builder := NewBuilder(db)
	if builder == nil {
		t.Fatal("expected builder, got nil")
	}
	if builder.DB != db {
		t.Error("expected db to match")
	}
}

func TestBuilderRaw(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	builder := NewBuilder(db)
	if builder.Raw() != db {
		t.Error("Raw() should return underlying db")
	}
}

func TestBuilderTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	builder := NewBuilder(db).Table("users")
	if builder.TableName == nil || *builder.TableName != "users" {
		t.Error("expected table name 'users'")
	}
}

func TestBuilderCreate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	t.Run("Create table with columns", func(t *testing.T) {
		err := NewBuilder(db).Table("users").Create(
			Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
			Column{Name: "name", Type: "TEXT", IsNullable: false},
			Column{Name: "email", Type: "TEXT", IsUnique: true},
			Column{Name: "age", Type: "INTEGER", Default: 0},
			Column{Name: "parent_id", Type: "INTEGER", ForeignKey: &Foreign{Table: "users", Column: "id"}},
		)
		if err != nil {
			t.Fatalf("failed to create table: %v", err)
		}
	})

	t.Run("Create without table name", func(t *testing.T) {
		err := NewBuilder(db).Create(Column{Name: "id", Type: "INTEGER"})
		if err == nil {
			t.Error("expected error for missing table name")
		}
	})

	t.Run("Create without columns", func(t *testing.T) {
		err := NewBuilder(db).Table("empty").Create()
		if err == nil {
			t.Error("expected error for no columns")
		}
	})

	t.Run("Create with invalid column name", func(t *testing.T) {
		err := NewBuilder(db).Table("test").Create(Column{Name: "invalid-col", Type: "TEXT"})
		if err == nil {
			t.Error("expected error for invalid column name")
		}
	})
}

func TestBuilderInsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("products").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
		Column{Name: "price", Type: "REAL"},
	)

	t.Run("Insert single row", func(t *testing.T) {
		id, err := NewBuilder(db).Table("products").Insert(map[string]any{
			"name":  "Widget",
			"price": 9.99,
		})
		if err != nil {
			t.Fatalf("insert failed: %v", err)
		}
		if id < 1 {
			t.Errorf("expected valid id, got %d", id)
		}
	})

	t.Run("Insert without table name", func(t *testing.T) {
		_, err := NewBuilder(db).Insert(map[string]any{"name": "test"})
		if err == nil {
			t.Error("expected error for missing table name")
		}
	})

	t.Run("Insert without data", func(t *testing.T) {
		_, err := NewBuilder(db).Table("products").Insert()
		if err == nil {
			t.Error("expected error for no data")
		}
	})

	t.Run("Insert with invalid column", func(t *testing.T) {
		_, err := NewBuilder(db).Table("products").Insert(map[string]any{"invalid-col": "x"})
		if err == nil {
			t.Error("expected error for invalid column")
		}
	})

	t.Run("Insert with invalid table", func(t *testing.T) {
		_, err := NewBuilder(db).Table("invalid-table").Insert(map[string]any{"name": "x"})
		if err == nil {
			t.Error("expected error for invalid table")
		}
	})
}

func TestBuilderInsertBatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("items").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
	)

	t.Run("InsertBatch multiple rows", func(t *testing.T) {
		affected, err := NewBuilder(db).Table("items").InsertBatch([]map[string]any{
			{"name": "item1"},
			{"name": "item2"},
			{"name": "item3"},
		})
		if err != nil {
			t.Fatalf("batch insert failed: %v", err)
		}
		if affected != 3 {
			t.Errorf("expected 3 affected rows, got %d", affected)
		}
	})

	t.Run("InsertBatch empty data", func(t *testing.T) {
		_, err := NewBuilder(db).Table("items").InsertBatch([]map[string]any{})
		if err == nil {
			t.Error("expected error for empty batch")
		}
	})

	t.Run("InsertBatch without table", func(t *testing.T) {
		_, err := NewBuilder(db).InsertBatch([]map[string]any{{"name": "x"}})
		if err == nil {
			t.Error("expected error for missing table")
		}
	})

	t.Run("InsertBatch with invalid column", func(t *testing.T) {
		_, err := NewBuilder(db).Table("items").InsertBatch([]map[string]any{{"invalid-col": "x"}})
		if err == nil {
			t.Error("expected error for invalid column")
		}
	})

	t.Run("InsertBatch with invalid table", func(t *testing.T) {
		_, err := NewBuilder(db).Table("invalid-table").InsertBatch([]map[string]any{{"name": "x"}})
		if err == nil {
			t.Error("expected error for invalid table")
		}
	})
}

func TestBuilderConflict(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("uniq").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true},
		Column{Name: "value", Type: "TEXT"},
	)

	NewBuilder(db).Table("uniq").Insert(map[string]any{"id": 1, "value": "first"})

	modes := []conflict{Ignore, Replace, Abort, Fail, Rollback}
	for i, mode := range modes {
		t.Run("Conflict mode", func(t *testing.T) {
			_, err := NewBuilder(db).Table("uniq").Conflict(mode).Insert(map[string]any{
				"id":    i + 10,
				"value": "test",
			})
			if err != nil {
				t.Errorf("insert with conflict mode %d failed: %v", mode, err)
			}
		})
	}

	t.Run("ON CONFLICT DO UPDATE", func(t *testing.T) {
		_, err := NewBuilder(db).Table("uniq").Insert(
			map[string]any{"id": 1, "value": "new"},
			map[string]any{"value": "updated"},
		)
		if err != nil {
			t.Fatalf("upsert failed: %v", err)
		}
	})
}

func TestBuilderSelect(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("sel_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
		Column{Name: "value", Type: "INTEGER"},
	)

	NewBuilder(db).Table("sel_test").InsertBatch([]map[string]any{
		{"name": "a", "value": 10},
		{"name": "b", "value": 20},
		{"name": "c", "value": 30},
	})

	t.Run("Select all columns", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("sel_test").Get()
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 3 {
			t.Errorf("expected 3 rows, got %d", count)
		}
	})

	t.Run("Select specific columns", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("sel_test").Select("name", "value").Get()
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		if len(cols) != 2 {
			t.Errorf("expected 2 columns, got %d", len(cols))
		}
	})

	t.Run("Select with asterisk", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("sel_test").Select("*").Get()
		if err != nil {
			t.Fatalf("select * failed: %v", err)
		}
		rows.Close()
	})

	t.Run("Select without table", func(t *testing.T) {
		_, err := NewBuilder(db).Get()
		if err == nil {
			t.Error("expected error for missing table")
		}
	})

	t.Run("Select with invalid column", func(t *testing.T) {
		_, err := NewBuilder(db).Table("sel_test").Select("invalid-col").Get()
		if err == nil {
			t.Error("expected error for invalid column")
		}
	})

	t.Run("Select with invalid table", func(t *testing.T) {
		_, err := NewBuilder(db).Table("invalid-table").Get()
		if err == nil {
			t.Error("expected error for invalid table")
		}
	})
}

func TestBuilderWhere(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("where_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "status", Type: "TEXT"},
		Column{Name: "count", Type: "INTEGER"},
	)

	NewBuilder(db).Table("where_test").InsertBatch([]map[string]any{
		{"status": "active", "count": 5},
		{"status": "inactive", "count": 10},
		{"status": "active", "count": 15},
	})

	tests := []struct {
		name     string
		builder  func() *Builder
		expected int64
	}{
		{"WhereEq", func() *Builder { return NewBuilder(db).Table("where_test").WhereEq("status", "active") }, 2},
		{"WhereNotEq", func() *Builder { return NewBuilder(db).Table("where_test").WhereNotEq("status", "active") }, 1},
		{"WhereGt", func() *Builder { return NewBuilder(db).Table("where_test").WhereGt("count", 5) }, 2},
		{"WhereLt", func() *Builder { return NewBuilder(db).Table("where_test").WhereLt("count", 15) }, 2},
		{"WhereGe", func() *Builder { return NewBuilder(db).Table("where_test").WhereGe("count", 10) }, 2},
		{"WhereLe", func() *Builder { return NewBuilder(db).Table("where_test").WhereLe("count", 10) }, 2},
		{"WhereBetween", func() *Builder { return NewBuilder(db).Table("where_test").WhereBetween("count", 5, 10) }, 2},
		{"WhereIn", func() *Builder { return NewBuilder(db).Table("where_test").WhereIn("count", []any{5, 15}) }, 2},
		{"WhereNotIn", func() *Builder { return NewBuilder(db).Table("where_test").WhereNotIn("count", []any{5, 15}) }, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tt.builder().Count()
			if err != nil {
				t.Fatalf("count failed: %v", err)
			}
			if count != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, count)
			}
		})
	}

	t.Run("Where with raw condition", func(t *testing.T) {
		count, err := NewBuilder(db).Table("where_test").Where("count > ?", 5).Count()
		if err != nil {
			t.Fatalf("count failed: %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2, got %d", count)
		}
	})
}

func TestBuilderWhereNull(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("null_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
	)

	NewBuilder(db).Table("null_test").Insert(map[string]any{"name": "notNull"})

	t.Run("WhereNull", func(t *testing.T) {
		builder := NewBuilder(db).Table("null_test").WhereNull("name")
		if len(builder.Error) > 0 {
			t.Errorf("unexpected error: %v", builder.Error)
		}
	})

	t.Run("WhereNotNull", func(t *testing.T) {
		count, err := NewBuilder(db).Table("null_test").WhereNotNull("name").Count()
		if err != nil {
			t.Fatalf("count failed: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1, got %d", count)
		}
	})
}

func TestBuilderOrWhere(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("or_where").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "val", Type: "INTEGER"},
	)

	NewBuilder(db).Table("or_where").InsertBatch([]map[string]any{
		{"val": 1}, {"val": 2}, {"val": 3},
	})

	tests := []struct {
		name    string
		builder func() *Builder
	}{
		{"OrWhere", func() *Builder { return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhere("val = ?", 1) }},
		{"OrWhereEq", func() *Builder { return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereEq("val", 1) }},
		{"OrWhereNotEq", func() *Builder { return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereNotEq("val", 2) }},
		{"OrWhereGt", func() *Builder { return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereGt("val", 2) }},
		{"OrWhereLt", func() *Builder { return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereLt("val", 2) }},
		{"OrWhereGe", func() *Builder { return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereGe("val", 3) }},
		{"OrWhereLe", func() *Builder { return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereLe("val", 1) }},
		{"OrWhereIn", func() *Builder {
			return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereIn("val", []any{1, 2})
		}},
		{"OrWhereNotIn", func() *Builder {
			return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereNotIn("val", []any{2, 3})
		}},
		{"OrWhereNull", func() *Builder { return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereNull("val") }},
		{"OrWhereNotNull", func() *Builder { return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereNotNull("val") }},
		{"OrWhereBetween", func() *Builder {
			return NewBuilder(db).Table("or_where").WhereEq("val", 999).OrWhereBetween("val", 1, 2)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := tt.builder().Get()
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
			rows.Close()
		})
	}
}

func TestBuilderUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("upd_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
		Column{Name: "counter", Type: "INTEGER", Default: 0},
		Column{Name: "active", Type: "INTEGER", Default: 1},
	)

	NewBuilder(db).Table("upd_test").Insert(map[string]any{"name": "test", "counter": 0, "active": 1})

	t.Run("Update with map", func(t *testing.T) {
		affected, err := NewBuilder(db).Table("upd_test").WhereEq("name", "test").Update(map[string]any{
			"name": "updated",
		})
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}
		if affected != 1 {
			t.Errorf("expected 1 affected, got %d", affected)
		}
	})

	t.Run("Increase", func(t *testing.T) {
		affected, err := NewBuilder(db).Table("upd_test").WhereEq("name", "updated").Increase("counter", 5).Update()
		if err != nil {
			t.Fatalf("increase failed: %v", err)
		}
		if affected != 1 {
			t.Errorf("expected 1 affected, got %d", affected)
		}
	})

	t.Run("Increase default", func(t *testing.T) {
		affected, err := NewBuilder(db).Table("upd_test").WhereEq("name", "updated").Increase("counter").Update()
		if err != nil {
			t.Fatalf("increase failed: %v", err)
		}
		if affected != 1 {
			t.Errorf("expected 1 affected, got %d", affected)
		}
	})

	t.Run("Decrease", func(t *testing.T) {
		affected, err := NewBuilder(db).Table("upd_test").WhereEq("name", "updated").Decrease("counter", 2).Update()
		if err != nil {
			t.Fatalf("decrease failed: %v", err)
		}
		if affected != 1 {
			t.Errorf("expected 1 affected, got %d", affected)
		}
	})

	t.Run("Decrease default", func(t *testing.T) {
		affected, err := NewBuilder(db).Table("upd_test").WhereEq("name", "updated").Decrease("counter").Update()
		if err != nil {
			t.Fatalf("decrease failed: %v", err)
		}
		if affected != 1 {
			t.Errorf("expected 1 affected, got %d", affected)
		}
	})

	t.Run("Toggle", func(t *testing.T) {
		affected, err := NewBuilder(db).Table("upd_test").WhereEq("name", "updated").Toggle("active").Update()
		if err != nil {
			t.Fatalf("toggle failed: %v", err)
		}
		if affected != 1 {
			t.Errorf("expected 1 affected, got %d", affected)
		}
	})

	t.Run("Update without data", func(t *testing.T) {
		_, err := NewBuilder(db).Table("upd_test").Update()
		if err == nil {
			t.Error("expected error for no update data")
		}
	})

	t.Run("Update without table", func(t *testing.T) {
		_, err := NewBuilder(db).Update(map[string]any{"name": "x"})
		if err == nil {
			t.Error("expected error for missing table")
		}
	})

	t.Run("Update with invalid table", func(t *testing.T) {
		_, err := NewBuilder(db).Table("invalid-table").Update(map[string]any{"name": "x"})
		if err == nil {
			t.Error("expected error for invalid table")
		}
	})

	t.Run("Update with invalid column", func(t *testing.T) {
		_, err := NewBuilder(db).Table("upd_test").Update(map[string]any{"invalid-col": "x"})
		if err == nil {
			t.Error("expected error for invalid column")
		}
	})
}

func TestBuilderDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("del_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
	)

	NewBuilder(db).Table("del_test").InsertBatch([]map[string]any{
		{"name": "delete_me"},
		{"name": "keep_me"},
	})

	t.Run("Delete with where", func(t *testing.T) {
		affected, err := NewBuilder(db).Table("del_test").WhereEq("name", "delete_me").Delete()
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}
		if affected != 1 {
			t.Errorf("expected 1 affected, got %d", affected)
		}
	})

	t.Run("Delete without where requires force", func(t *testing.T) {
		_, err := NewBuilder(db).Table("del_test").Delete()
		if err == nil {
			t.Error("expected error for delete without where")
		}
	})

	t.Run("Delete with force", func(t *testing.T) {
		affected, err := NewBuilder(db).Table("del_test").Delete(true)
		if err != nil {
			t.Fatalf("force delete failed: %v", err)
		}
		if affected != 1 {
			t.Errorf("expected 1 affected, got %d", affected)
		}
	})

	t.Run("Delete without table", func(t *testing.T) {
		_, err := NewBuilder(db).WhereEq("id", 1).Delete()
		if err == nil {
			t.Error("expected error for missing table")
		}
	})

	t.Run("Delete with invalid table", func(t *testing.T) {
		_, err := NewBuilder(db).Table("invalid-table").WhereEq("id", 1).Delete()
		if err == nil {
			t.Error("expected error for invalid table")
		}
	})
}

func TestBuilderDeleteUnsupportedClauses(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("del_unsup").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
	)

	tests := []struct {
		name    string
		builder func() *Builder
	}{
		{"Delete with JOIN", func() *Builder { return NewBuilder(db).Table("del_unsup").Join("other", "1=1").WhereEq("id", 1) }},
		{"Delete with GROUP BY", func() *Builder { return NewBuilder(db).Table("del_unsup").GroupBy("id").WhereEq("id", 1) }},
		{"Delete with HAVING", func() *Builder { return NewBuilder(db).Table("del_unsup").Having("1=1").WhereEq("id", 1) }},
		{"Delete with ORDER BY", func() *Builder { return NewBuilder(db).Table("del_unsup").OrderBy("id").WhereEq("id", 1) }},
		{"Delete with LIMIT", func() *Builder { return NewBuilder(db).Table("del_unsup").Limit(1).WhereEq("id", 1) }},
		{"Delete with OFFSET", func() *Builder { return NewBuilder(db).Table("del_unsup").Offset(1).WhereEq("id", 1) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.builder().Delete()
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestBuilderFirst(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("first_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
	)

	NewBuilder(db).Table("first_test").InsertBatch([]map[string]any{
		{"name": "first"},
		{"name": "second"},
	})

	t.Run("First returns single row", func(t *testing.T) {
		row, err := NewBuilder(db).Table("first_test").OrderBy("id", Asc).First()
		if err != nil {
			t.Fatalf("first failed: %v", err)
		}

		var id int
		var name string
		if err := row.Scan(&id, &name); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		if name != "first" {
			t.Errorf("expected 'first', got '%s'", name)
		}
	})

	t.Run("First without table", func(t *testing.T) {
		_, err := NewBuilder(db).First()
		if err == nil {
			t.Error("expected error for missing table")
		}
	})
}

func TestBuilderCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("count_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
	)

	db.Exec(`INSERT INTO "count_test" DEFAULT VALUES`)
	db.Exec(`INSERT INTO "count_test" DEFAULT VALUES`)
	db.Exec(`INSERT INTO "count_test" DEFAULT VALUES`)

	t.Run("Count returns row count", func(t *testing.T) {
		count, err := NewBuilder(db).Table("count_test").Count()
		if err != nil {
			t.Fatalf("count failed: %v", err)
		}
		if count != 3 {
			t.Errorf("expected 3, got %d", count)
		}
	})

	t.Run("Count without table", func(t *testing.T) {
		_, err := NewBuilder(db).Count()
		if err == nil {
			t.Error("expected error for missing table")
		}
	})
}

func TestBuilderOrderBy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("order_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "value", Type: "INTEGER"},
	)

	NewBuilder(db).Table("order_test").InsertBatch([]map[string]any{
		{"value": 3}, {"value": 1}, {"value": 2},
	})

	t.Run("OrderBy Asc", func(t *testing.T) {
		row, err := NewBuilder(db).Table("order_test").OrderBy("value", Asc).First()
		if err != nil {
			t.Fatalf("first failed: %v", err)
		}

		var id, value int
		row.Scan(&id, &value)
		if value != 1 {
			t.Errorf("expected 1, got %d", value)
		}
	})

	t.Run("OrderBy Desc", func(t *testing.T) {
		row, err := NewBuilder(db).Table("order_test").OrderBy("value", Desc).First()
		if err != nil {
			t.Fatalf("first failed: %v", err)
		}

		var id, value int
		row.Scan(&id, &value)
		if value != 3 {
			t.Errorf("expected 3, got %d", value)
		}
	})

	t.Run("OrderBy default Asc", func(t *testing.T) {
		row, err := NewBuilder(db).Table("order_test").OrderBy("value").First()
		if err != nil {
			t.Fatalf("first failed: %v", err)
		}

		var id, value int
		row.Scan(&id, &value)
		if value != 1 {
			t.Errorf("expected 1, got %d", value)
		}
	})
}

func TestBuilderLimitOffset(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("limit_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true},
	)

	for i := 1; i <= 10; i++ {
		NewBuilder(db).Table("limit_test").Insert(map[string]any{"id": i})
	}

	t.Run("Limit", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("limit_test").Limit(3).Get()
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 3 {
			t.Errorf("expected 3, got %d", count)
		}
	})

	t.Run("Limit with offset", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("limit_test").Limit(3).Offset(5).Get()
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 3 {
			t.Errorf("expected 3, got %d", count)
		}
	})

	t.Run("Limit with two args", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("limit_test").Limit(2, 3).Get()
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 3 {
			t.Errorf("expected 3, got %d", count)
		}
	})

	t.Run("Limit no args", func(t *testing.T) {
		builder := NewBuilder(db).Table("limit_test").Limit()
		if builder.WithLimit != nil {
			t.Error("expected nil limit")
		}
	})
}

func TestBuilderGroupBy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("group_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "category", Type: "TEXT"},
		Column{Name: "value", Type: "INTEGER"},
	)

	NewBuilder(db).Table("group_test").InsertBatch([]map[string]any{
		{"category": "A", "value": 10},
		{"category": "A", "value": 20},
		{"category": "B", "value": 30},
	})

	t.Run("GroupBy", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("group_test").Select("category").GroupBy("category").Get()
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 2 {
			t.Errorf("expected 2 groups, got %d", count)
		}
	})

	t.Run("GroupBy empty", func(t *testing.T) {
		builder := NewBuilder(db).Table("group_test").GroupBy()
		if len(builder.GroupByList) != 0 {
			t.Error("expected empty GroupByList")
		}
	})

	t.Run("GroupBy with invalid column", func(t *testing.T) {
		builder := NewBuilder(db).Table("group_test").GroupBy("invalid-col")
		if len(builder.GroupByList) != 0 {
			t.Error("expected invalid column to be skipped")
		}
	})
}

func TestBuilderHaving(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("having_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "category", Type: "TEXT"},
		Column{Name: "value", Type: "INTEGER"},
	)

	NewBuilder(db).Table("having_test").InsertBatch([]map[string]any{
		{"category": "A", "value": 10},
		{"category": "A", "value": 20},
		{"category": "B", "value": 5},
	})

	tests := []struct {
		name    string
		builder func() *Builder
	}{
		{"Having", func() *Builder {
			return NewBuilder(db).Table("having_test").GroupBy("category").Having("SUM(value) > ?", 20)
		}},
		{"HavingEq", func() *Builder {
			return NewBuilder(db).Table("having_test").GroupBy("category", "value").HavingEq("category", "A")
		}},
		{"HavingNotEq", func() *Builder {
			return NewBuilder(db).Table("having_test").GroupBy("category", "value").HavingNotEq("category", "B")
		}},
		{"HavingGt", func() *Builder {
			return NewBuilder(db).Table("having_test").GroupBy("category", "value").HavingGt("value", 5)
		}},
		{"HavingLt", func() *Builder {
			return NewBuilder(db).Table("having_test").GroupBy("category", "value").HavingLt("value", 20)
		}},
		{"HavingGe", func() *Builder {
			return NewBuilder(db).Table("having_test").GroupBy("category", "value").HavingGe("value", 10)
		}},
		{"HavingLe", func() *Builder {
			return NewBuilder(db).Table("having_test").GroupBy("category", "value").HavingLe("value", 20)
		}},
		{"HavingIn", func() *Builder {
			return NewBuilder(db).Table("having_test").GroupBy("category", "value").HavingIn("value", []any{10, 20})
		}},
		{"HavingNotIn", func() *Builder {
			return NewBuilder(db).Table("having_test").GroupBy("category", "value").HavingNotIn("value", []any{5})
		}},
		{"HavingBetween", func() *Builder {
			return NewBuilder(db).Table("having_test").GroupBy("category", "value").HavingBetween("value", 5, 20)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := tt.builder().Get()
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
			rows.Close()
		})
	}
}

func TestBuilderHavingNullNotNull(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("hav_null").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "cat", Type: "TEXT"},
		Column{Name: "flag", Type: "INTEGER"},
	)

	NewBuilder(db).Table("hav_null").Insert(map[string]any{"cat": "A", "flag": 1})
	NewBuilder(db).Table("hav_null").Insert(map[string]any{"cat": "B", "flag": 2})

	t.Run("HavingNull", func(t *testing.T) {
		builder := NewBuilder(db).Table("hav_null").GroupBy("cat", "flag").HavingNull("flag")
		if len(builder.Error) > 0 {
			t.Errorf("unexpected error: %v", builder.Error)
		}
	})

	t.Run("HavingNotNull", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("hav_null").GroupBy("cat", "flag").HavingNotNull("flag").Get()
		if err != nil {
			t.Fatalf("HavingNotNull failed: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 2 {
			t.Errorf("expected 2, got %d", count)
		}
	})
}

func TestBuilderOrHaving(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("or_hav").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "cat", Type: "TEXT"},
		Column{Name: "val", Type: "INTEGER"},
	)

	NewBuilder(db).Table("or_hav").InsertBatch([]map[string]any{
		{"cat": "A", "val": 10},
		{"cat": "B", "val": 20},
		{"cat": "C", "val": 30},
	})

	tests := []struct {
		name    string
		builder func() *Builder
	}{
		{"OrHaving", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHaving("val = ?", 10)
		}},
		{"OrHavingEq", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingEq("cat", "A")
		}},
		{"OrHavingNotEq", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingNotEq("cat", "X")
		}},
		{"OrHavingGt", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingGt("val", 25)
		}},
		{"OrHavingLt", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingLt("val", 15)
		}},
		{"OrHavingGe", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingGe("val", 20)
		}},
		{"OrHavingLe", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingLe("val", 10)
		}},
		{"OrHavingIn", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingIn("val", []any{10, 30})
		}},
		{"OrHavingNotIn", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingNotIn("val", []any{20})
		}},
		{"OrHavingNull", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingNull("val")
		}},
		{"OrHavingNotNull", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingNotNull("val")
		}},
		{"OrHavingBetween", func() *Builder {
			return NewBuilder(db).Table("or_hav").GroupBy("cat", "val").Having("1=0").OrHavingBetween("val", 15, 25)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := tt.builder().Get()
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
			rows.Close()
		})
	}
}

func TestBuilderJoin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("users_j").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
	)
	NewBuilder(db).Table("orders_j").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "user_id", Type: "INTEGER"},
		Column{Name: "amount", Type: "REAL"},
	)

	NewBuilder(db).Table("users_j").Insert(map[string]any{"name": "Alice"})
	NewBuilder(db).Table("orders_j").Insert(map[string]any{"user_id": 1, "amount": 99.99})

	t.Run("Inner Join", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("users_j").
			Select("*").
			Join("orders_j", `"users_j"."id" = "orders_j"."user_id"`).
			Get()
		if err != nil {
			t.Fatalf("join failed: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("expected at least one row")
		}
	})

	t.Run("Left Join", func(t *testing.T) {
		NewBuilder(db).Table("users_j").Insert(map[string]any{"name": "Bob"})

		rows, err := NewBuilder(db).Table("users_j").
			Select("*").
			LeftJoin("orders_j", `"users_j"."id" = "orders_j"."user_id"`).
			Get()
		if err != nil {
			t.Fatalf("left join failed: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count < 2 {
			t.Errorf("expected at least 2 rows, got %d", count)
		}
	})

	t.Run("Join with invalid table", func(t *testing.T) {
		_, err := NewBuilder(db).Table("users_j").Join("invalid-table", "1=1").Get()
		if err == nil {
			t.Error("expected error for invalid join table")
		}
	})

	t.Run("Join with empty ON", func(t *testing.T) {
		_, err := NewBuilder(db).Table("users_j").Join("orders_j", "").Get()
		if err == nil {
			t.Error("expected error for empty ON clause")
		}
	})
}

func TestBuilderContext(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("ctx_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
	)

	t.Run("Context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := NewBuilder(db).Table("ctx_test").Context(ctx).Insert(map[string]any{"name": "test"})
		if err != nil {
			t.Fatalf("insert with context failed: %v", err)
		}
	})

	t.Run("Context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := NewBuilder(db).Table("ctx_test").Context(ctx).Insert(map[string]any{"name": "test"})
		if err == nil {
			t.Error("expected error for cancelled context")
		}
	})

	t.Run("Context for Get", func(t *testing.T) {
		ctx := context.Background()
		rows, err := NewBuilder(db).Table("ctx_test").Context(ctx).Get()
		if err != nil {
			t.Fatalf("get with context failed: %v", err)
		}
		rows.Close()
	})

	t.Run("Context for First", func(t *testing.T) {
		ctx := context.Background()
		_, err := NewBuilder(db).Table("ctx_test").Context(ctx).First()
		if err != nil {
			t.Fatalf("first with context failed: %v", err)
		}
	})

	t.Run("Context for Count", func(t *testing.T) {
		ctx := context.Background()
		_, err := NewBuilder(db).Table("ctx_test").Context(ctx).Count()
		if err != nil {
			t.Fatalf("count with context failed: %v", err)
		}
	})
}

func TestBuilderTotal(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("total_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
	)

	for i := 0; i < 10; i++ {
		NewBuilder(db).Table("total_test").Insert(map[string]any{"name": "item"})
	}

	t.Run("Total with limit", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("total_test").Total().Limit(3).Get()
		if err != nil {
			t.Fatalf("get with total failed: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var total int
			var id int
			var name string
			if err := rows.Scan(&total, &id, &name); err != nil {
				t.Fatalf("scan failed: %v", err)
			}
			if total != 10 {
				t.Errorf("expected total 10, got %d", total)
			}
		}
	})
}

func TestBuilderErrorPropagation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("err_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true},
	)

	t.Run("Error propagates to Insert", func(t *testing.T) {
		_, err := NewBuilder(db).Table("err_test").WhereEq("invalid-col", 1).Insert(map[string]any{"id": 1})
		if err == nil {
			t.Error("expected error to propagate")
		}
	})

	t.Run("Error propagates to Update", func(t *testing.T) {
		_, err := NewBuilder(db).Table("err_test").WhereEq("invalid-col", 1).Update(map[string]any{"id": 1})
		if err == nil {
			t.Error("expected error to propagate")
		}
	})

	t.Run("Error propagates to Delete", func(t *testing.T) {
		_, err := NewBuilder(db).Table("err_test").WhereEq("invalid-col", 1).Delete()
		if err == nil {
			t.Error("expected error to propagate")
		}
	})

	t.Run("Error collected in Get", func(t *testing.T) {
		builder := NewBuilder(db).Table("err_test").WhereEq("invalid-col", 1)
		if len(builder.Error) == 0 {
			t.Error("expected error to be collected")
		}
	})

	t.Run("Error propagates to First", func(t *testing.T) {
		_, err := NewBuilder(db).Table("err_test").WhereEq("invalid-col", 1).First()
		if err == nil {
			t.Error("expected error to propagate")
		}
	})

	t.Run("Error propagates to Count", func(t *testing.T) {
		_, err := NewBuilder(db).Table("err_test").WhereEq("invalid-col", 1).Count()
		if err == nil {
			t.Error("expected error to propagate")
		}
	})

	t.Run("Error propagates to InsertBatch", func(t *testing.T) {
		_, err := NewBuilder(db).Table("err_test").WhereEq("invalid-col", 1).InsertBatch([]map[string]any{{"id": 1}})
		if err == nil {
			t.Error("expected error to propagate")
		}
	})
}

func TestValidateColumn(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid_column", false},
		{"Valid123", false},
		{"_underscore", false},
		{"", true},
		{"invalid-column", true},
		{"123invalid", true},
		{"SELECT", true},
		{"INSERT", true},
		{string(make([]byte, 129)), true},
	}

	for _, tt := range tests {
		err := ValidateColumn(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateColumn(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{"hello", "'hello'"},
		{42, "42"},
		{int64(42), "42"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
		{struct{ Name string }{"test"}, "'{test}'"},
	}

	for _, tt := range tests {
		result := FormatValue(tt.input)
		if result != tt.expected {
			t.Errorf("FormatValue(%v) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestDeprecatedMethods(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("dep_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		Column{Name: "name", Type: "TEXT"},
	)

	ctx := context.Background()

	t.Run("InsertContext", func(t *testing.T) {
		_, err := NewBuilder(db).Table("dep_test").InsertContext(ctx, map[string]any{"name": "test"})
		if err != nil {
			t.Fatalf("InsertContext failed: %v", err)
		}
	})

	t.Run("InsertReturningID", func(t *testing.T) {
		id, err := NewBuilder(db).Table("dep_test").InsertReturningID(map[string]any{"name": "test2"})
		if err != nil {
			t.Fatalf("InsertReturningID failed: %v", err)
		}
		if id < 1 {
			t.Error("expected valid id")
		}
	})

	t.Run("InsertContextReturningID", func(t *testing.T) {
		id, err := NewBuilder(db).Table("dep_test").InsertContextReturningID(ctx, map[string]any{"name": "test3"})
		if err != nil {
			t.Fatalf("InsertContextReturningID failed: %v", err)
		}
		if id < 1 {
			t.Error("expected valid id")
		}
	})

	t.Run("InsertConflict", func(t *testing.T) {
		_, err := NewBuilder(db).Table("dep_test").InsertConflict(Ignore, map[string]any{"name": "test4"})
		if err != nil {
			t.Fatalf("InsertConflict failed: %v", err)
		}
	})

	t.Run("InsertContexConflict", func(t *testing.T) {
		_, err := NewBuilder(db).Table("dep_test").InsertContexConflict(ctx, Ignore, map[string]any{"name": "test5"})
		if err != nil {
			t.Fatalf("InsertContexConflict failed: %v", err)
		}
	})

	t.Run("InsertConflictReturningID", func(t *testing.T) {
		id, err := NewBuilder(db).Table("dep_test").InsertConflictReturningID(Ignore, map[string]any{"name": "test6"})
		if err != nil {
			t.Fatalf("InsertConflictReturningID failed: %v", err)
		}
		if id < 1 {
			t.Error("expected valid id")
		}
	})

	t.Run("InsertContextConflictReturningID", func(t *testing.T) {
		id, err := NewBuilder(db).Table("dep_test").InsertContextConflictReturningID(ctx, Ignore, map[string]any{"name": "test7"})
		if err != nil {
			t.Fatalf("InsertContextConflictReturningID failed: %v", err)
		}
		if id < 1 {
			t.Error("expected valid id")
		}
	})

	t.Run("GetContext", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("dep_test").GetContext(ctx)
		if err != nil {
			t.Fatalf("GetContext failed: %v", err)
		}
		rows.Close()
	})

	t.Run("FirstContext", func(t *testing.T) {
		row, err := NewBuilder(db).Table("dep_test").FirstContext(ctx)
		if err != nil {
			t.Fatalf("FirstContext failed: %v", err)
		}
		var id int
		var name string
		row.Scan(&id, &name)
	})

	t.Run("CountContext", func(t *testing.T) {
		count, err := NewBuilder(db).Table("dep_test").CountContext(ctx)
		if err != nil {
			t.Fatalf("CountContext failed: %v", err)
		}
		if count < 1 {
			t.Error("expected count > 0")
		}
	})

	t.Run("UpdateContext", func(t *testing.T) {
		_, err := NewBuilder(db).Table("dep_test").WhereEq("id", 1).UpdateContext(ctx, map[string]any{"name": "updated"})
		if err != nil {
			t.Fatalf("UpdateContext failed: %v", err)
		}
	})

	t.Run("GetWithTotal", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("dep_test").Limit(1).GetWithTotal()
		if err != nil {
			t.Fatalf("GetWithTotal failed: %v", err)
		}
		rows.Close()
	})

	t.Run("GetWithTotalContext", func(t *testing.T) {
		rows, err := NewBuilder(db).Table("dep_test").Limit(1).GetWithTotalContext(ctx)
		if err != nil {
			t.Fatalf("GetWithTotalContext failed: %v", err)
		}
		rows.Close()
	})
}

func TestInvalidColumnValidation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("val_test").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true},
	)

	methods := []struct {
		name    string
		builder func() *Builder
	}{
		{"Increase", func() *Builder { return NewBuilder(db).Table("val_test").Increase("invalid-col") }},
		{"Decrease", func() *Builder { return NewBuilder(db).Table("val_test").Decrease("invalid-col") }},
		{"Toggle", func() *Builder { return NewBuilder(db).Table("val_test").Toggle("invalid-col") }},
		{"WhereEq", func() *Builder { return NewBuilder(db).Table("val_test").WhereEq("invalid-col", 1) }},
		{"WhereNotEq", func() *Builder { return NewBuilder(db).Table("val_test").WhereNotEq("invalid-col", 1) }},
		{"WhereGt", func() *Builder { return NewBuilder(db).Table("val_test").WhereGt("invalid-col", 1) }},
		{"WhereLt", func() *Builder { return NewBuilder(db).Table("val_test").WhereLt("invalid-col", 1) }},
		{"WhereGe", func() *Builder { return NewBuilder(db).Table("val_test").WhereGe("invalid-col", 1) }},
		{"WhereLe", func() *Builder { return NewBuilder(db).Table("val_test").WhereLe("invalid-col", 1) }},
		{"WhereIn", func() *Builder { return NewBuilder(db).Table("val_test").WhereIn("invalid-col", []any{1}) }},
		{"WhereNotIn", func() *Builder { return NewBuilder(db).Table("val_test").WhereNotIn("invalid-col", []any{1}) }},
		{"WhereNull", func() *Builder { return NewBuilder(db).Table("val_test").WhereNull("invalid-col") }},
		{"WhereNotNull", func() *Builder { return NewBuilder(db).Table("val_test").WhereNotNull("invalid-col") }},
		{"WhereBetween", func() *Builder { return NewBuilder(db).Table("val_test").WhereBetween("invalid-col", 1, 2) }},
		{"OrWhereEq", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereEq("invalid-col", 1) }},
		{"OrWhereNotEq", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereNotEq("invalid-col", 1) }},
		{"OrWhereGt", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereGt("invalid-col", 1) }},
		{"OrWhereLt", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereLt("invalid-col", 1) }},
		{"OrWhereGe", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereGe("invalid-col", 1) }},
		{"OrWhereLe", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereLe("invalid-col", 1) }},
		{"OrWhereIn", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereIn("invalid-col", []any{1}) }},
		{"OrWhereNotIn", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereNotIn("invalid-col", []any{1}) }},
		{"OrWhereNull", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereNull("invalid-col") }},
		{"OrWhereNotNull", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereNotNull("invalid-col") }},
		{"OrWhereBetween", func() *Builder { return NewBuilder(db).Table("val_test").OrWhereBetween("invalid-col", 1, 2) }},
		{"HavingEq", func() *Builder { return NewBuilder(db).Table("val_test").HavingEq("invalid-col", 1) }},
		{"HavingNotEq", func() *Builder { return NewBuilder(db).Table("val_test").HavingNotEq("invalid-col", 1) }},
		{"HavingGt", func() *Builder { return NewBuilder(db).Table("val_test").HavingGt("invalid-col", 1) }},
		{"HavingLt", func() *Builder { return NewBuilder(db).Table("val_test").HavingLt("invalid-col", 1) }},
		{"HavingGe", func() *Builder { return NewBuilder(db).Table("val_test").HavingGe("invalid-col", 1) }},
		{"HavingLe", func() *Builder { return NewBuilder(db).Table("val_test").HavingLe("invalid-col", 1) }},
		{"HavingIn", func() *Builder { return NewBuilder(db).Table("val_test").HavingIn("invalid-col", []any{1}) }},
		{"HavingNotIn", func() *Builder { return NewBuilder(db).Table("val_test").HavingNotIn("invalid-col", []any{1}) }},
		{"HavingNull", func() *Builder { return NewBuilder(db).Table("val_test").HavingNull("invalid-col") }},
		{"HavingNotNull", func() *Builder { return NewBuilder(db).Table("val_test").HavingNotNull("invalid-col") }},
		{"HavingBetween", func() *Builder { return NewBuilder(db).Table("val_test").HavingBetween("invalid-col", 1, 2) }},
		{"OrHavingEq", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingEq("invalid-col", 1) }},
		{"OrHavingNotEq", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingNotEq("invalid-col", 1) }},
		{"OrHavingGt", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingGt("invalid-col", 1) }},
		{"OrHavingLt", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingLt("invalid-col", 1) }},
		{"OrHavingGe", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingGe("invalid-col", 1) }},
		{"OrHavingLe", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingLe("invalid-col", 1) }},
		{"OrHavingIn", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingIn("invalid-col", []any{1}) }},
		{"OrHavingNotIn", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingNotIn("invalid-col", []any{1}) }},
		{"OrHavingNull", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingNull("invalid-col") }},
		{"OrHavingNotNull", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingNotNull("invalid-col") }},
		{"OrHavingBetween", func() *Builder { return NewBuilder(db).Table("val_test").OrHavingBetween("invalid-col", 1, 2) }},
	}

	for _, tt := range methods {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.builder()
			if len(builder.Error) == 0 {
				t.Errorf("%s should collect error for invalid column", tt.name)
			}
		})
	}
}

func TestEmptyValuesErrors(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	NewBuilder(db).Table("empty_val").Create(
		Column{Name: "id", Type: "INTEGER", IsPrimary: true},
	)

	methods := []struct {
		name    string
		builder func() *Builder
	}{
		{"WhereIn empty", func() *Builder { return NewBuilder(db).Table("empty_val").WhereIn("id", []any{}) }},
		{"WhereNotIn empty", func() *Builder { return NewBuilder(db).Table("empty_val").WhereNotIn("id", []any{}) }},
		{"OrWhereIn empty", func() *Builder { return NewBuilder(db).Table("empty_val").OrWhereIn("id", []any{}) }},
		{"OrWhereNotIn empty", func() *Builder { return NewBuilder(db).Table("empty_val").OrWhereNotIn("id", []any{}) }},
		{"HavingIn empty", func() *Builder { return NewBuilder(db).Table("empty_val").HavingIn("id", []any{}) }},
		{"HavingNotIn empty", func() *Builder { return NewBuilder(db).Table("empty_val").HavingNotIn("id", []any{}) }},
		{"OrHavingIn empty", func() *Builder { return NewBuilder(db).Table("empty_val").OrHavingIn("id", []any{}) }},
		{"OrHavingNotIn empty", func() *Builder { return NewBuilder(db).Table("empty_val").OrHavingNotIn("id", []any{}) }},
	}

	for _, tt := range methods {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.builder()
			if len(builder.Error) == 0 {
				t.Errorf("%s should collect error for empty values", tt.name)
			}
		})
	}
}
