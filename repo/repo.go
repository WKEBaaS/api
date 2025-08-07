// Package repo provides repository interfaces and implementations for data access.
package repo

import "errors"

var (
	ErrTransactionFailed = errors.New("database transaction failed")
	ErrDatabaseError     = errors.New("database error")
)
