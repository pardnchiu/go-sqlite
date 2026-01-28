package goSqlite

import (
	"context"
	"database/sql"
	"sync"
)

type Config struct {
	Key      string `json:"key,omitempty"`
	Path     string `json:"path"`
	Lifetime int    `json:"lifetime,omitempty"` // sec
}

type Connector struct {
	db map[string]*sql.DB
	mu sync.Mutex
}

// * Builder is NOT safe for concurrent use by multiple goroutines
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
	groupBy    []string
	havingList []Where
	havingArgs []any
	limit      *int
	offset     *int
	withTotal  bool
	context    context.Context
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

type Union struct {
	builder *Builder
	all     bool
}

type conflict uint32

type direction uint32
