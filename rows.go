package god1

import "database/sql/driver"

// https://pkg.go.dev/database/sql/driver#Rows
// driver.Rows

var _ driver.Rows = (*Rows)(nil)

type Rows struct{}

func (r *Rows) Columns() []string {
	return nil
}

func (r *Rows) Close() error {
	return ErrNotImplemented
}

func (r *Rows) Next(dest []driver.Value) error {
	return ErrNotImplemented
}
