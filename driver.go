package god1

import (
	"database/sql"
	"database/sql/driver"
)

func init() {
	sql.Register("god1", &Driver{})
}

// https://pkg.go.dev/database/sql/driver#Driver
// driver.Driver

var _ driver.Driver = (*Driver)(nil)

type Driver struct{}

func (d *Driver) Open(name string) (driver.Conn, error) {
	return nil, ErrNotImplemented
}
