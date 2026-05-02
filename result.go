package god1

import "database/sql/driver"

// https://pkg.go.dev/database/sql/driver#Result
// driver.Result

var _ driver.Result = (*result)(nil)

type result struct {
	rowsAffected int64
	lastInsertId int64
}

func (r *result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func (r *result) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}
