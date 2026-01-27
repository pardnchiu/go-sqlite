package goSqlite

import (
	"fmt"
	"regexp"
)

var (
	columnRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

func quote(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func validateColumn(name string) error {
	if !columnRegex.MatchString(name) {
		return fmt.Errorf("invalid identifier: %s", name)
	}
	return nil
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("'%v'", val)
	}
}
