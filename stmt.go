package god1

import "database/sql/driver"

// https://pkg.go.dev/database/sql/driver#Stmt
// driver.Stmt

var _ driver.Stmt = (*Stmt)(nil)

type Stmt struct{}

func (s *Stmt) Close() error {
	return ErrNotImplemented
}

func (s *Stmt) NumInput() int {
	return -1
}

func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, ErrNotImplemented
}

func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return nil, ErrNotImplemented
}
