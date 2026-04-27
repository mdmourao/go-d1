package god1

import (
	"context"
	"database/sql/driver"
	"strings"
)

// TODO - make this configurable?
const d1ReportedSQLiteVersion = "3.53.0"

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
	s.conn.logger.Debug("Exec called", "query", s.query, "args", args)

	// TODO - review parsing
	params := make([]any, len(args))
	for i, v := range args {
		params[i] = v
	}

	// TODO - review execute
	response, err := s.conn.client.Exec(context.TODO(), s.query, params)
	if err != nil {
		// TODO - error handling
		return nil, err
	}

	s.conn.logger.Debug("Received response", "response", response)

	return &Result{
		rowsAffected: response.Changes,
		lastInsertId: response.LastRowID,
	}, nil
}

// SELECT
// TODO Deprecated
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	s.conn.logger.Debug("Query called", "query", s.query, "args", args)

	// TODO - best way to detect this?
	if strings.Contains(s.query, "sqlite_version()") {
		s.conn.logger.Debug("Intercepted sqlite_version() probe", "version", d1ReportedSQLiteVersion)
		return &Rows{
			columns: []string{"sqlite_version()"},
			data: []map[string]any{
				{"sqlite_version()": d1ReportedSQLiteVersion},
			},
		}, nil
	}

	// TODO - review parsing
	params := make([]any, len(args))
	for i, v := range args {
		params[i] = v
	}

	// TODO - review execute
	data, err := s.conn.client.Query(context.TODO(), s.query, params)
	if err != nil {
		// TODO - error handling
		return nil, err
	}

	rows, err := newRows(data)
	if err != nil {
		return nil, err
	}

	s.conn.logger.Debug("Received results", "columns", rows.columns, "results", rows.data)

	return rows, nil
}
