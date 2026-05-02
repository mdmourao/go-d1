package god1

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
)

var defaultDriver = &d1Driver{}

func init() { sql.Register("god1", defaultDriver) }

// https://pkg.go.dev/database/sql/driver#Driver
// driver.Driver

var _ driver.Driver = (*d1Driver)(nil)

type d1Driver struct{}

func (d *d1Driver) Open(name string) (driver.Conn, error) {
	c, err := NewConnector(name)
	if err != nil {
		return nil, err
	}
	return c.Connect(context.Background())
}

// allow drivers access to context and to avoid repeated parsing of driver configuration
func (d *d1Driver) OpenConnector(name string) (driver.Connector, error) {
	return NewConnector(name)
}

func parseDSN(dsn string) (*url.URL, error) {
	if dsn == "" {
		return nil, ErrDriverMissingDSN
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDriverInvalidDSN, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, ErrDriverInvalidDSNScheme
	}
	if u.Host == "" {
		return nil, ErrDriverInvalidDSNHost
	}

	return u, nil
}
