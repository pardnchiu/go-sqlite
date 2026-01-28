package core

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

//go:embed sql_keywords.json
var sqlKeys []byte

var (
	columnRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	keyMap      map[string]bool
	maxLength   = 128
)

func init() {
	var keywords []string
	if err := json.Unmarshal(sqlKeys, &keywords); err != nil {
		panic(fmt.Sprintf("failed to load reserved keywords: %v", err))
	}

	keyMap = make(map[string]bool, len(keywords))
	for _, key := range keywords {
		keyMap[key] = true
	}
}

func quote(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func ValidateColumn(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("identifier is required")
	}

	if len(name) > maxLength {
		return fmt.Errorf("out of maximum length: %s", name)
	}

	if !columnRegex.MatchString(name) {
		return fmt.Errorf("invalid identifier: %s", name)
	}

	upperName := strings.ToUpper(name)
	if keyMap[upperName] {
		return fmt.Errorf("cannot be used as identifier: %s", name)
	}
	return nil
}

func FormatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("'%v'", val)
	}
}
