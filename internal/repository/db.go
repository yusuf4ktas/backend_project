package repository

import (
	"context"
	"database/sql"
)

// DBTX to write repository methods that can work with either a normal database connection/database transaction for atomic operations.
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
