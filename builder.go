package goSqlite

import (
	"database/sql"
	"fmt"
	"strings"
)

type Builder struct {
	db         *sql.DB
	table      *string
	selectList []string
	updateList []string
	whereList  []Where
	whereArgs  []any
	joinList   []Join
	conflict   *conflict
	orderBy    []string
	limit      *int
	offset     *int
	withTotal  bool
}

type Where struct {
	condition string
	operator  string
}

type Join struct {
	mode  string
	table string
	on    string
}

type Column struct {
	Name         string
	Type         string
	IsPrimary    bool
	IsNullable   bool
	AutoIncrease bool
	IsUnique     bool
	Default      any
	ForeignKey   *Foreign
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

	_, err := b.db.Exec(sb.String())
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

func builderClear(b *Builder) {
	b.selectList = []string{}
	b.updateList = []string{}
	b.whereList = []Where{}
	b.whereArgs = []any{}
	b.joinList = []Join{}
	b.conflict = nil
	b.orderBy = []string{}
	b.limit = nil
	b.offset = nil
	b.withTotal = false
}
