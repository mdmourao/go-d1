package god1

import (
	"database/sql/driver"
	"log/slog"

	"github.com/mdmourao/go-d1/internal/transport"
)

// https://pkg.go.dev/database/sql/driver#Conn
// driver.Conn

var (
	_ driver.Conn = (*Conn)(nil)
)

type Conn struct {
	client *transport.Client
	logger *slog.Logger
}

func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	return &Stmt{
		query: query,
		conn:  c,
	}, nil
}

func (c *Conn) Close() error {
	return nil
}

func (c *Conn) Begin() (driver.Tx, error) {
	return Tx{}, ErrTransactionsNotSupported
}
