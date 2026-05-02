package god1

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io"
	"strconv"

	"github.com/buger/jsonparser"
)

// https://pkg.go.dev/database/sql/driver#Rows
// driver.Rows

var _ driver.Rows = (*rows)(nil)

type rows struct {
	data    [][]driver.Value
	columns []string
	index   int
}

// newRows decodes a JSON array of row objects
// jsonparser promisses big performance with a smaller memory footprint
func newRows(data []byte) (*rows, error) {
	var columns []string
	_, err := jsonparser.ArrayEach(data, func(v []byte, _ jsonparser.ValueType, _ int, _ error) {
		s, _ := jsonparser.ParseString(v)
		columns = append(columns, s)
	}, "columns")
	if err != nil {
		return nil, err
	}

	var rowsData [][]driver.Value
	var parseErr error
	_, err = jsonparser.ArrayEach(data, func(rowVal []byte, _ jsonparser.ValueType, _ int, _ error) {
		if parseErr != nil {
			return
		}
		row := make([]driver.Value, 0, len(columns))
		_, e := jsonparser.ArrayEach(rowVal, func(cell []byte, t jsonparser.ValueType, _ int, _ error) {
			if parseErr != nil {
				return
			}
			v, perr := parseJsonValue(cell, t)
			if perr != nil {
				parseErr = perr
				return
			}
			row = append(row, v)
		})
		if e != nil && parseErr == nil {
			parseErr = e
		}
		rowsData = append(rowsData, row)
	}, "rows")
	if err != nil {
		return nil, err
	}
	if parseErr != nil {
		return nil, parseErr
	}
	return &rows{columns: columns, data: rowsData}, nil
}

func parseJsonValue(b []byte, t jsonparser.ValueType) (driver.Value, error) {
	switch t {
	case jsonparser.String:
		return jsonparser.ParseString(b)
	case jsonparser.Number:
		s := string(b)
		if !bytes.ContainsAny(b, ".eE") {
			return strconv.ParseInt(s, 10, 64)
		}
		return strconv.ParseFloat(s, 64)
	case jsonparser.Boolean:
		return jsonparser.ParseBoolean(b)
	case jsonparser.Null:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported JSON value type: %v", t)
	}
}

func (r *rows) Columns() []string {
	return r.columns
}

func (r *rows) Close() error {
	return nil
}

func (r *rows) Next(dest []driver.Value) error {
	if r.index >= len(r.data) {
		return io.EOF
	}

	copy(dest, r.data[r.index])

	r.index++
	return nil
}
