package store

import "database/sql"

type Scannable interface {
	Scan(dest ...interface{}) error
}

type Querier interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}