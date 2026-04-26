package god1

import (
	"context"
	"database/sql/driver"
)

// https://pkg.go.dev/database/sql/driver#Stmt
// driver.Stmt

var _ driver.Stmt = (*Stmt)(nil)

type Stmt struct {
	query string
	conn  *Conn
}

func (s *Stmt) Close() error {
	return nil
}

func (s *Stmt) NumInput() int {
	return -1
}

// INSERT, UPDATE, DELETE
func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, ErrNotImplemented
}

// SELECT
// TODO Deprecated
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	s.conn.logger.Debug("Query called", "query", s.query, "args", args)

	// TODO - review parsing
	params := make([]any, len(args))
	for i, v := range args {
		params[i] = v
	}

	// TODO - review execute
	if _, err := s.conn.client.Execute(context.TODO(), s.query, params); err != nil {
		return nil, err
	}

	return nil, ErrNotImplemented
}
