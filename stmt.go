package god1

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
)

// https://pkg.go.dev/database/sql/driver#Stmt
// driver.Stmt

var (
	_ driver.Stmt             = (*Stmt)(nil)
	_ driver.StmtExecContext  = (*Stmt)(nil)
	_ driver.StmtQueryContext = (*Stmt)(nil)

	sqliteVersionProbeRe = regexp.MustCompile(`(?i)^\s*SELECT\s+sqlite_version\(\)\s*(?:AS\s+\w+)?\s*;?\s*$`)
)

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
	namedArgs := make([]driver.NamedValue, len(args))
	for i, v := range args {
		namedArgs[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return s.ExecContext(context.Background(), namedArgs)
}

func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	s.conn.logger.Debug("exec context called", "query", s.query, "args", args)

	params, err := buildParams(args)
	if err != nil {
		return nil, err
	}

	response, err := s.conn.client.Exec(ctx, s.query, params)
	if err != nil {
		return nil, err
	}

	s.conn.logger.Debug("received response", "response", response)

	return &Result{
		rowsAffected: response.Changes,
		lastInsertId: response.LastRowID,
	}, nil
}

// SELECT
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	namedArgs := make([]driver.NamedValue, len(args))
	for i, v := range args {
		namedArgs[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return s.QueryContext(context.Background(), namedArgs)
}

func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	s.conn.logger.Debug("query context called", "query", s.query, "args", args)

	if sqliteVersionProbeRe.MatchString(s.query) {
		s.conn.logger.Debug("intercepted sqlite_version() probe, returning mock version")
		return &Rows{
			columns: []string{"sqlite_version()"},
			data: [][]driver.Value{
				{s.conn.sqliteVersion},
			},
		}, nil
	}

	params, err := buildParams(args)
	if err != nil {
		return nil, err
	}

	response, err := s.conn.client.Query(ctx, s.query, params)
	if err != nil {
		return nil, err
	}

	s.conn.logger.Debug("received response", "response", response)

	rows, err := newRows(response)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func buildParams(args []driver.NamedValue) ([]any, error) {
	if len(args) == 0 {
		return nil, nil
	}

	params := make([]any, len(args))
	for i, arg := range args {
		if arg.Name != "" {
			return nil, fmt.Errorf("named parameters are not supported")
		}
		params[i] = arg.Value
	}
	return params, nil
}
