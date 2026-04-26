package god1

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
)

// https://pkg.go.dev/database/sql/driver#Rows
// driver.Rows

var _ driver.Rows = (*Rows)(nil)

type Rows struct {
	data    []map[string]any
	columns []string
	index   int
}

// newRows decodes a JSON array of row objects, recovering the column
// order from the first row's raw JSON since unmarshalling into
// map[string]any loses key order.
func newRows(data []byte) (*Rows, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return &Rows{}, nil
	}

	columns, err := jsonObjectKeys(raw[0])
	if err != nil {
		return nil, err
	}

	rows := make([]map[string]any, len(raw))
	for i, r := range raw {
		if err := json.Unmarshal(r, &rows[i]); err != nil {
			return nil, err
		}
	}
	return &Rows{data: rows, columns: columns}, nil
}

// TODO - tests and review
// jsonObjectKeys returns the keys of a JSON object in source order.
func jsonObjectKeys(data []byte) ([]string, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	if _, err := dec.Token(); err != nil { // opening '{'
		return nil, err
	}
	var keys []string
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("expected string key, got %v", t)
		}
		keys = append(keys, key)
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil, err
		}
	}
	return keys, nil
}

func (r *Rows) Columns() []string {
	if r.columns == nil {
		return []string{}
	}
	return r.columns
}

func (r *Rows) Close() error {
	return nil
}

func (r *Rows) Next(dest []driver.Value) error {
	if len(r.data) == 0 || len(r.columns) == 0 || r.index >= len(r.data) {
		return io.EOF
	}

	row := r.data[r.index]
	for i, col := range r.columns {
		dest[i] = row[col]
	}

	r.index++
	return nil
}
