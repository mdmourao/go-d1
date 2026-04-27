package god1

import "database/sql/driver"

// https://pkg.go.dev/database/sql/driver#Result
// driver.Result

var _ driver.Result = (*Result)(nil)

type Result struct {
	rowsAffected int64
	lastInsertId int64
}

func (r *Result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func (r *Result) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}
