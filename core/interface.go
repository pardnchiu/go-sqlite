package core

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
	Map map[string]*sql.DB
	Mu  sync.Mutex
}

// * Builder is NOT safe for concurrent use by multiple goroutines
type Builder struct {
	DB           *sql.DB
	TableName    *string
	SelectList   []string
	UpdateList   []string
	WhereList    []Where
	WhereArgs    []any
	JoinList     []Join
	ConflictMode *conflict
	OrderByList  []string
	GroupByList  []string
	HavingList   []Where
	HavingArgs   []any
	WithLimit    *int
	WithOffset   *int
	WithTotal    bool
	WithContext  context.Context
	Error        []error
}

type Where struct {
	Condition string
	Operator  string
}

type Join struct {
	Mode  string
	Table string
	On    string
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
	Builder *Builder
	All     bool
}

type conflict uint32

type direction uint32
