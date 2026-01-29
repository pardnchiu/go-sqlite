package core

import (
	"database/sql"
	"fmt"
	"strings"
)

func NewBuilder(db *sql.DB) *Builder {
	return &Builder{
		DB: db,
	}
}

func (b *Builder) Raw() *sql.DB {
	return b.DB
}

func (b *Builder) Table(name string) *Builder {
	b.TableName = &name
	return b
}

func (b *Builder) Create(columns ...Column) error {
	if b.TableName == nil {
		return fmt.Errorf("table name is required")
	}
	if len(columns) == 0 {
		return fmt.Errorf("no columns defined")
	}

	var sb strings.Builder
	sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	sb.WriteString(quote(*b.TableName))
	sb.WriteString(" (")

	for i, col := range columns {
		if i > 0 {
			sb.WriteString(", ")
		}
		if err := ValidateColumn(col.Name); err != nil {
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
		parts = append(parts, fmt.Sprintf("DEFAULT %v", FormatValue(c.Default)))
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

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	if len(b.WhereList) == 0 && (len(force) == 0 || !force[0]) {
		return 0, fmt.Errorf("delete without where need to use force = true")
	}

	if b.TableName == nil {
		return 0, fmt.Errorf("table name is required")
	}

	if err := ValidateColumn(*b.TableName); err != nil {
		return 0, err
	}

	if len(b.JoinList) > 0 {
		return 0, fmt.Errorf("SQLite DELETE does not support JOIN")
	}

	if len(b.GroupByList) > 0 {
		return 0, fmt.Errorf("SQLite DELETE does not support GROUP BY")
	}

	if len(b.HavingList) > 0 || len(b.HavingArgs) > 0 {
		return 0, fmt.Errorf("SQLite DELETE does not support HAVING")
	}

	if len(b.OrderByList) > 0 {
		return 0, fmt.Errorf("SQLite DELETE does not support ORDER BY")
	}

	if b.WithLimit != nil || b.WithOffset != nil {
		return 0, fmt.Errorf("SQLite DELETE does not support LIMIT / OFFSET")
	}

	var sb strings.Builder
	sb.WriteString("DELETE FROM ")
	sb.WriteString(quote(*b.TableName))
	sb.WriteString(b.buildWhere())

	result, err := b.ExecAutoAsignContext(sb.String(), b.WhereArgs...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func builderClear(b *Builder) {
	b.SelectList = []string{}
	b.UpdateList = []string{}
	b.WhereList = []Where{}
	b.WhereArgs = []any{}
	b.JoinList = []Join{}
	b.ConflictMode = nil
	b.OrderByList = []string{}
	b.GroupByList = []string{}
	b.HavingList = []Where{}
	b.HavingArgs = []any{}
	b.WithLimit = nil
	b.WithOffset = nil
	b.WithTotal = false
	b.WithContext = nil
}

func (b *Builder) ExecAutoAsignContext(query string, args ...any) (sql.Result, error) {
	var result sql.Result
	var err error
	if b.WithContext != nil {
		result, err = b.DB.ExecContext(b.WithContext, query, args...)
	} else {
		result, err = b.DB.Exec(query, args...)
	}
	if err != nil && strings.Contains(err.Error(), "readonly") {
		return nil, fmt.Errorf("write operation on read-only db: %w", err)
	}
	return result, nil
}
