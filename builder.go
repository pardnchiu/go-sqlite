package goSqlite

import (
	"database/sql"
	"fmt"
	"strings"
)

func NewBuilder(db *sql.DB) *Builder {
	return &Builder{
		db: db,
	}
}

func (b *Builder) Raw() *sql.DB {
	return b.db
}

func (b *Builder) Table(name string) *Builder {
	b.table = &name
	return b
}

func (b *Builder) Create(columns ...Column) error {
	if b.table == nil {
		return fmt.Errorf("table name is required")
	}
	if len(columns) == 0 {
		return fmt.Errorf("no columns defined")
	}

	var sb strings.Builder
	sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	sb.WriteString(quote(*b.table))
	sb.WriteString(" (")

	for i, col := range columns {
		if i > 0 {
			sb.WriteString(", ")
		}
		if err := validateColumn(col.Name); err != nil {
			return err
		}
		sb.WriteString(quote(col.Name))
		sb.WriteString(" ")
		sb.WriteString(buildColumn(col))
	}

	sb.WriteString(")")

	_, err := b.ExecAutoAsignContext(sb.String())
	return err
}

func buildColumn(c Column) string {
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
		parts = append(parts, fmt.Sprintf("DEFAULT %v", formatValue(c.Default)))
	}

	if c.ForeignKey != nil {
		parts = append(parts, fmt.Sprintf("REFERENCES %s(%s)",
			quote(c.ForeignKey.Table),
			quote(c.ForeignKey.Column)))
	}

	return strings.Join(parts, " ")
}

func (b *Builder) Delete(force ...bool) (int64, error) {
	defer builderClear(b)

	if len(b.whereList) == 0 && (len(force) == 0 || !force[0]) {
		return 0, fmt.Errorf("delete without where need to use force = true")
	}

	if b.table == nil {
		return 0, fmt.Errorf("table name is required")
	}

	if err := validateColumn(*b.table); err != nil {
		return 0, err
	}

	if len(b.joinList) > 0 {
		return 0, fmt.Errorf("SQLite DELETE does not support JOIN")
	}

	if len(b.groupBy) > 0 {
		return 0, fmt.Errorf("SQLite DELETE does not support GROUP BY")
	}

	if len(b.havingList) > 0 || len(b.havingArgs) > 0 {
		return 0, fmt.Errorf("SQLite DELETE does not support HAVING")
	}

	if len(b.orderBy) > 0 {
		return 0, fmt.Errorf("SQLite DELETE does not support ORDER BY")
	}

	if b.limit != nil || b.offset != nil {
		return 0, fmt.Errorf("SQLite DELETE does not support LIMIT / OFFSET")
	}

	var sb strings.Builder
	sb.WriteString("DELETE FROM ")
	sb.WriteString(quote(*b.table))
	sb.WriteString(b.buildWhere())

	result, err := b.ExecAutoAsignContext(sb.String(), b.whereArgs...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func builderClear(b *Builder) {
	b.selectList = []string{}
	b.updateList = []string{}
	b.whereList = []Where{}
	b.whereArgs = []any{}
	b.joinList = []Join{}
	b.conflict = nil
	b.orderBy = []string{}
	b.groupBy = []string{}
	b.havingList = []Where{}
	b.havingArgs = []any{}
	b.limit = nil
	b.offset = nil
	b.withTotal = false
	b.context = nil
}

func (b *Builder) ExecAutoAsignContext(query string, args ...any) (sql.Result, error) {
	if b.context != nil {
		return b.db.ExecContext(b.context, query, args...)
	} else {
		return b.db.Exec(query, args...)
	}
}
