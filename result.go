package god1

import "database/sql/driver"

// https://pkg.go.dev/database/sql/driver#Result
// driver.Result

var _ driver.Result = (*Result)(nil)

type Result struct{}

func (r *Result) RowsAffected() (int64, error) {
	return 0, ErrNotImplemented
}

func (r *Result) LastInsertId() (int64, error) {
	return 0, ErrNotImplemented
}
