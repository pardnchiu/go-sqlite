package core

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

func (b *Builder) First() (*sql.Row, error) {
	defer builderClear(b)

	if len(b.OrderByList) == 0 {
		b.OrderByList = []string{"ROWID DESC"}
	} else {
		for i, order := range b.OrderByList {
			if strings.Contains(order, " ASC") {
				b.OrderByList[i] = strings.Replace(order, " ASC", " DESC", 1)
			} else if strings.Contains(order, " DESC") {
				b.OrderByList[i] = strings.Replace(order, " DESC", " ASC", 1)
			}
		}
	}

	b.Limit(1)

	if len(b.Error) > 0 {
		return nil, b.Error[0]
	}

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}

	args := append(b.WhereArgs, b.HavingArgs...)
	var row *sql.Row
	if b.WithContext != nil {
		row = b.DB.QueryRowContext(b.WithContext, query, args...)
	} else {
		row = b.DB.QueryRow(query, args...)
	}

	if b.WithBind != nil {
		targetVal := reflect.ValueOf(b.WithBind)
		if targetVal.Kind() != reflect.Pointer {
			return row, fmt.Errorf("target must be a pointer")
		}

		targetElem := targetVal.Elem()
		if targetElem.Kind() != reflect.Struct {
			return row, fmt.Errorf("target must be struct")
		}

		return row, findRow(row, targetElem)
	}
	return row, nil
}

func (b *Builder) Last() (*sql.Row, error) {
	defer builderClear(b)

	b.Limit(1)

	if len(b.Error) > 0 {
		return nil, b.Error[0]
	}

	query, err := selectBuilder(b, false)
	if err != nil {
		return nil, err
	}

	args := append(b.WhereArgs, b.HavingArgs...)
	var row *sql.Row
	if b.WithContext != nil {
		row = b.DB.QueryRowContext(b.WithContext, query, args...)
	} else {
		row = b.DB.QueryRow(query, args...)
	}

	if b.WithBind != nil {
		targetVal := reflect.ValueOf(b.WithBind)
		if targetVal.Kind() != reflect.Pointer {
			return row, fmt.Errorf("target must be a pointer")
		}

		targetElem := targetVal.Elem()
		if targetElem.Kind() != reflect.Struct {
			return row, fmt.Errorf("target must be struct")
		}

		return row, findRow(row, targetElem)
	}
	return row, nil
}

func (b *Builder) Count() (int64, error) {
	defer builderClear(b)

	if len(b.Error) > 0 {
		return 0, b.Error[0]
	}

	query, err := selectBuilder(b, true)
	if err != nil {
		return 0, err
	}

	args := append(b.WhereArgs, b.HavingArgs...)
	var count int64
	if b.WithContext != nil {
		err = b.DB.QueryRowContext(b.WithContext, query, args...).Scan(&count)
	} else {
		err = b.DB.QueryRow(query, args...).Scan(&count)
	}
	return count, err
}
