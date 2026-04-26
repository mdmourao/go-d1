package god1

import "database/sql/driver"

// https://pkg.go.dev/database/sql/driver#Conn
// driver.Conn

var _ driver.Conn = (*Conn)(nil)

type Conn struct{}

func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return &Stmt{}, ErrNotImplemented
}

func (c *Conn) Close() error {
	return ErrNotImplemented
}

func (c *Conn) Begin() (driver.Tx, error) {
	return nil, ErrNotImplemented
}
