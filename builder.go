package goSqlite

import (
	"database/sql"
	"fmt"
	"strings"
)

type Builder struct {
	db    *sql.DB
	table *string
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
		db: db,
	}
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
	sb.WriteString(*b.table)
	sb.WriteString(" (")

	for i, col := range columns {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(col.Name)
		sb.WriteString(" ")
		sb.WriteString(b.buildColumn(col))
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
