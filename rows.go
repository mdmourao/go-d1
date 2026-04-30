package god1

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/buger/jsonparser"
)

// https://pkg.go.dev/database/sql/driver#Rows
// driver.Rows

var _ driver.Rows = (*Rows)(nil)

type Rows struct {
	data    [][]driver.Value
	columns []string
	index   int
}

// newRows decodes a JSON array of row objects, recovering the column
// order from the first row's raw JSON
// jsonparser promisses big performance with a smaller memory footprint
func newRows(data []byte) (*Rows, error) {
	var columns []string

	firstObj, dataType, _, err := jsonparser.Get(data, "[0]")
	if err != nil {
		if err == jsonparser.KeyPathNotFoundError {
			return &Rows{}, nil
		}
		return nil, err
	}

	if dataType != jsonparser.Object {
		return nil, fmt.Errorf("expected object at index 0, got %v", dataType)
	}

	err = jsonparser.ObjectEach(firstObj, func(key []byte, value []byte, dt jsonparser.ValueType, offset int) error {
		columns = append(columns, string(key))
		return nil
	})
	if err != nil {
		return nil, err
	}

	var rows [][]driver.Value
	var parseErr error

	_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if parseErr != nil {
			return
		}
		if err != nil {
			parseErr = err
			return
		}

		row := make([]driver.Value, len(columns))

		for i, col := range columns {
			valBytes, valType, _, valErr := jsonparser.Get(value, col)

			if errors.Is(valErr, jsonparser.KeyPathNotFoundError) {
				row[i] = nil
				continue
			} else if valErr != nil {
				parseErr = valErr
				return
			}

			row[i], parseErr = parseJsonValue(valBytes, valType)
			if parseErr != nil {
				return
			}
		}

		rows = append(rows, row)
	})

	if err != nil {
		return nil, err
	}
	if parseErr != nil {
		return nil, parseErr
	}

	return &Rows{data: rows, columns: columns}, nil
}

func parseJsonValue(b []byte, t jsonparser.ValueType) (driver.Value, error) {
	switch t {
	case jsonparser.String:
		return jsonparser.ParseString(b)
	case jsonparser.Number:
		if i, err := strconv.ParseInt(string(b), 10, 64); err == nil {
			return i, nil
		}
		return strconv.ParseFloat(string(b), 64)
	case jsonparser.Boolean:
		return jsonparser.ParseBoolean(b)
	case jsonparser.Null:
		return nil, nil
	case jsonparser.Array, jsonparser.Object:
		out := make([]byte, len(b))
		copy(out, b)
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported JSON value type: %v", t)
	}
}

func (r *Rows) Columns() []string {
	return r.columns
}

func (r *Rows) Close() error {
	return nil
}

func (r *Rows) Next(dest []driver.Value) error {
	if r.index >= len(r.data) {
		return io.EOF
	}

	copy(dest, r.data[r.index])

	r.index++
	return nil
}
