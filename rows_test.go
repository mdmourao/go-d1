package god1

import (
	"database/sql/driver"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRows(t *testing.T) {
	data := []byte(`{
        "columns": ["id", "name", "active", "score"],
        "rows": [
            [1, "Gopher", true, 1.5],
            [2, "Camel", false, null]
        ]
    }`)

	r, err := newRows(data)
	assert.NoError(t, err)
	assert.Equal(t, []string{"id", "name", "active", "score"}, r.columns)
	assert.Len(t, r.data, 2)
	assert.Equal(t, int64(1), r.data[0][0])
	assert.Equal(t, "Gopher", r.data[0][1])
	assert.Equal(t, true, r.data[0][2])
	assert.Equal(t, 1.5, r.data[0][3])

	assert.Equal(t, int64(2), r.data[1][0])
	assert.Equal(t, "Camel", r.data[1][1])
	assert.Equal(t, false, r.data[1][2])
	assert.Nil(t, r.data[1][3])

	assert.Equal(t, []string{"id", "name", "active", "score"}, r.Columns())
	assert.Nil(t, r.Close())
}

func TestNewRowsNoColumns(t *testing.T) {
	data := []byte(`{
				"rows": [
					[1, "Gopher", true, 1.5],
					[2, "Camel", false, null]
				]
			}`)

	_, err := newRows(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrColumns)
}

func TestNewRowsNoRows(t *testing.T) {
	data := []byte(`{
				"columns": ["id", "name", "active", "score"]
			}`)

	_, err := newRows(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRows)
}

func TestNewRowsInvalidColumns(t *testing.T) {
	data := []byte(`{
				"columns": "id, name, active, score",
				"rows": [
					[1, "Gopher", true, 1.5],
					[2, "Camel", false, null]
				]
			}`)

	_, err := newRows(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrColumns)
}

func TestNewRowsInvalidRows(t *testing.T) {
	data := []byte(`{
				"columns": ["id", "name", "active", "score"],
				"rows": "invalid rows"
			}`)

	_, err := newRows(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRows)
}

func TestNewRowsMissingColumns(t *testing.T) {
	data := []byte(`{
				"columns": ["id"],
				"rows": [
					[1, "Gopher", true, 1.5],
					[2, "Camel", false, null]
				]
			}`)
	_, err := newRows(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrParseRows)
}

func TestNewRowsInvalidCell(t *testing.T) {
	data := []byte(`{
				"columns": ["id", "name", "active", "score"],
				"rows": [
					[1, "Gopher", true, 1.5],
					[2, "Camel", false, {"invalid": "cell"}]
				]
			}`)
	_, err := newRows(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrParseRows)
}

func TestNewRowsColumnCountMismatch(t *testing.T) {
	data := []byte(`{
				"columns": ["id", "name", "active", "score"],
				"rows": [
					[1, "Gopher", true],
					[2, "Camel", false, null]
				]
			}`)
	_, err := newRows(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrParseRows)
}

func TestNewRowsEmpty(t *testing.T) {
	data := []byte(`{
				"columns": ["id", "name", "active", "score"],
				"rows": []
			}`)
	r, err := newRows(data)
	assert.NoError(t, err)
	assert.Equal(t, []string{"id", "name", "active", "score"}, r.columns)
	assert.Len(t, r.data, 0)
}

func TestNewRowsNullValues(t *testing.T) {
	data := []byte(`{
				"columns": ["id", "name", "active", "score"],
				"rows": [
					[1, null, true, 1.5],
					[2, "Camel", false, null]
				]
			}`)
	r, err := newRows(data)
	assert.NoError(t, err)
	assert.Equal(t, []string{"id", "name", "active", "score"}, r.columns)
	assert.Len(t, r.data, 2)
	assert.Equal(t, int64(1), r.data[0][0])
	assert.Nil(t, r.data[0][1])
	assert.Equal(t, true, r.data[0][2])
	assert.Equal(t, 1.5, r.data[0][3])

	assert.Equal(t, int64(2), r.data[1][0])
	assert.Equal(t, "Camel", r.data[1][1])
	assert.Equal(t, false, r.data[1][2])
	assert.Nil(t, r.data[1][3])
}

func TestNewRowsErrParseRows(t *testing.T) {
	data := []byte(`{ "columns": ["a"], "rows": [ "not-an-array" ] }`)
	_, err := newRows(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrParseRows)
}

func TestNewRowsErrParseRows2(t *testing.T) {
	data := []byte(`{
  "columns": ["a", "b"],
  "rows": [
    [ {"x": 1}, "ok" ]
  ]
}`)
	_, err := newRows(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrParseRows)
}

func TestNewRowsErrParseRows3(t *testing.T) {
	data := []byte(`{ "columns": ["a","b"], "rows": [ [ [1,2], null ] ] }`)
	_, err := newRows(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrParseRows)
}

func TestNext(t *testing.T) {
	data := []byte(`{
				"columns": ["id", "name"],
				"rows": [
					[1, "Gopher"],
					[2, "Camel"]
				]
			}`)
	r, err := newRows(data)
	assert.NoError(t, err)

	var row []driver.Value = make([]driver.Value, len(r.columns))
	err = r.Next(row)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), row[0])
	assert.Equal(t, "Gopher", row[1])

	err = r.Next(row)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), row[0])
	assert.Equal(t, "Camel", row[1])

	err = r.Next(row)
	assert.ErrorIs(t, err, io.EOF)
}
