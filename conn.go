package god1

import (
	"context"
	"database/sql/driver"
	"log/slog"
	"regexp"

	"github.com/mdmourao/go-d1/internal/transport"
)

// https://pkg.go.dev/database/sql/driver#Conn
// https://pkg.go.dev/database/sql/driver#ConnBeginTx
// driver.Conn

var (
	_ driver.Conn              = (*conn)(nil)
	_ driver.ConnBeginTx       = (*conn)(nil)
	_ driver.Pinger            = (*conn)(nil)
	_ driver.QueryerContext    = (*conn)(nil)
	_ driver.ExecerContext     = (*conn)(nil)
	_ driver.NamedValueChecker = (*conn)(nil)

	sqliteVersionProbeRe = regexp.MustCompile(`(?i)^\s*SELECT\s+sqlite_version\(\)\s*(?:AS\s+\w+)?\s*;?\s*$`)
)

const jsMaxSafeInteger = 1<<53 - 1

type conn struct {
	client        *transport.Client
	logger        *slog.Logger
	sqliteVersion string
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return &stmt{
		query: query,
		conn:  c,
	}, nil
}

func (c *conn) Close() error {
	return nil
}

func (c *conn) Begin() (driver.Tx, error) {
	return nil, ErrTransactionsNotSupported
}

func (c *conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return nil, ErrTransactionsNotSupported
}

func (c *conn) Ping(ctx context.Context) error {
	return c.client.Ping(ctx)
}

func (c *conn) CheckNamedValue(nv *driver.NamedValue) error {
	if nv.Name != "" {
		return ErrNamedParametersNotSupported
	}
	switch v := nv.Value.(type) {
	case []byte:
		return ErrBlobNotSupported
	case int64:
		if v > jsMaxSafeInteger || v < -jsMaxSafeInteger {
			return ErrInt64OutOfJSSafeRange
		}
		return nil
	}
	converted, err := driver.DefaultParameterConverter.ConvertValue(nv.Value)
	if err != nil {
		return err
	}
	nv.Value = converted
	return nil
}

func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.logger.Debug("query context called", "query", query, "args", args)

	if sqliteVersionProbeRe.MatchString(query) {
		c.logger.Debug("intercepted sqlite_version() probe, returning mock version")
		return &rows{
			columns: []string{"sqlite_version()"},
			data: [][]driver.Value{
				{c.sqliteVersion},
			},
		}, nil
	}

	params, err := buildParams(args)
	if err != nil {
		return nil, err
	}

	response, err := c.client.Query(ctx, query, params)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("received response", "response", response)

	return newRows(response)
}

func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.logger.Debug("exec context called", "query", query, "args", args)

	params, err := buildParams(args)
	if err != nil {
		return nil, err
	}

	response, err := c.client.Exec(ctx, query, params)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("received response", "response", response)

	return &result{
		rowsAffected: response.Changes,
		lastInsertId: response.LastRowID,
	}, nil
}

func buildParams(args []driver.NamedValue) ([]any, error) {
	if len(args) == 0 {
		return nil, nil
	}

	params := make([]any, len(args))
	for i, arg := range args {
		if arg.Name != "" {
			return nil, ErrNamedParametersNotSupported
		}
		params[i] = arg.Value
	}
	return params, nil
}
