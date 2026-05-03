package god1

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mdmourao/go-d1/internal/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepare(t *testing.T) {
	c := &conn{}
	s, err := c.Prepare("SELECT * FROM users WHERE id = ?")
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users WHERE id = ?", s.(*stmt).query)
}

func TestClose(t *testing.T) {
	c := &conn{}
	err := c.Close()
	assert.NoError(t, err)
}

func TestBegin(t *testing.T) {
	c := &conn{}
	_, err := c.Begin()
	assert.ErrorIs(t, err, ErrTransactionsNotSupported)
}

func TestBeginTx(t *testing.T) {
	c := &conn{}
	_, err := c.BeginTx(context.Background(), driver.TxOptions{})
	assert.ErrorIs(t, err, ErrTransactionsNotSupported)
}

func TestCheckNamedValueErrNamedParametersNotSupported(t *testing.T) {
	c := &conn{}
	err := c.CheckNamedValue(&driver.NamedValue{Name: "id", Value: 1})
	assert.ErrorIs(t, err, ErrNamedParametersNotSupported)
}

func TestCheckNamedValueNoError(t *testing.T) {
	c := &conn{}
	err := c.CheckNamedValue(&driver.NamedValue{Ordinal: 1, Value: 1})
	assert.NoError(t, err)
}

func TestCheckNamedValueErrBlobNotSupported(t *testing.T) {
	c := &conn{}
	err := c.CheckNamedValue(&driver.NamedValue{Ordinal: 1, Value: []byte{0x01, 0x02}})
	assert.ErrorIs(t, err, ErrBlobNotSupported)
}

func TestCheckNamedValueInt64(t *testing.T) {
	tests := []struct {
		name    string
		value   int64
		wantErr error
	}{
		{
			name:    "zero",
			value:   0,
			wantErr: nil,
		},
		{
			name:    "positive within range",
			value:   12345,
			wantErr: nil,
		},
		{
			name:    "negative within range",
			value:   -12345,
			wantErr: nil,
		},
		{
			name:    "max safe integer",
			value:   jsMaxSafeInteger,
			wantErr: nil,
		},
		{
			name:    "negative max safe integer",
			value:   -jsMaxSafeInteger,
			wantErr: nil,
		},
		{
			name:    "max safe integer plus one",
			value:   jsMaxSafeInteger + 1,
			wantErr: ErrInt64OutOfJSSafeRange,
		},
		{
			name:    "negative max safe integer minus one",
			value:   -jsMaxSafeInteger - 1,
			wantErr: ErrInt64OutOfJSSafeRange,
		},
		{
			name:    "math max int64",
			value:   math.MaxInt64,
			wantErr: ErrInt64OutOfJSSafeRange,
		},
		{
			name:    "math min int64",
			value:   math.MinInt64,
			wantErr: ErrInt64OutOfJSSafeRange,
		},
	}

	c := &conn{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nv := &driver.NamedValue{Ordinal: 1, Value: tt.value}
			err := c.CheckNamedValue(nv)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.value, nv.Value)
			}
		})
	}
}

func TestCheckNamedValueDefaultConverterError(t *testing.T) {
	c := &conn{}
	nv := &driver.NamedValue{Ordinal: 1, Value: make(chan int)}
	err := c.CheckNamedValue(nv)
	assert.Error(t, err)
}

func TestCheckNamedValueDefaultConverterConverts(t *testing.T) {
	c := &conn{}
	nv := &driver.NamedValue{Ordinal: 1, Value: int(42)}
	err := c.CheckNamedValue(nv)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), nv.Value)
}

func TestQueryContextSqliteVersionProbe(t *testing.T) {
	c := &conn{
		sqliteVersion: "3.42.0",
		logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	rows, err := c.QueryContext(context.Background(), "SELECT sqlite_version()", nil)
	assert.NoError(t, err)

	dest := make([]driver.Value, 1)
	err = rows.Next(dest)
	assert.NoError(t, err)
	assert.Equal(t, "3.42.0", dest[0])
}

func TestBuildParams(t *testing.T) {
	args := []driver.NamedValue{
		{Ordinal: 1, Value: "Alice"},
		{Ordinal: 2, Value: 30},
	}
	params, err := buildParams(args)
	assert.NoError(t, err)
	assert.Equal(t, []any{"Alice", 30}, params)

	args = []driver.NamedValue{}
	params, err = buildParams(args)
	assert.NoError(t, err)
	assert.Len(t, params, 0)

	args = []driver.NamedValue{
		{Ordinal: 1, Value: "Alice", Name: "name"},
	}
	_, err = buildParams(args)
	assert.ErrorIs(t, err, ErrNamedParametersNotSupported)
}

func TestPing(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotPayload transport.Payload
		var gotMethod, gotContentType string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotContentType = r.Header.Get("Content-Type")
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		c := &conn{
			client: transport.NewClient(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, "", ""),
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		err := c.Ping(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "application/json", gotContentType)
		assert.Equal(t, "SELECT 1", gotPayload.SQL)
		assert.False(t, gotPayload.IsExec)
	})

	t.Run("non-200 returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := &conn{
			client: transport.NewClient(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, "", ""),
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		err := c.Ping(context.Background())
		assert.Error(t, err)
	})

	t.Run("context cancelled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		c := &conn{
			client: transport.NewClient(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, "", ""),
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := c.Ping(ctx)
		assert.Error(t, err)
	})
}

func TestQueryContext(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotPayload transport.Payload
		var gotMethod, gotContentType string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotContentType = r.Header.Get("Content-Type")
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
        "columns": ["id"],
        "rows": [
            [1],
            [2]
        ]
    }`))
		}))
		defer server.Close()

		c := &conn{
			client: transport.NewClient(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, "", ""),
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		s := &stmt{
			query: "SELECT id from mascots",
			conn:  c,
		}

		rows, err := s.Query([]driver.Value{})
		assert.NoError(t, err)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "application/json", gotContentType)
		assert.Equal(t, "SELECT id from mascots", gotPayload.SQL)
		assert.False(t, gotPayload.IsExec)
		assert.Equal(t, []any(nil), gotPayload.Args)
		assert.Len(t, rows.Columns(), 1)

		assert.Equal(t, []string{"id"}, rows.Columns())
		var dest = make([]driver.Value, 1)
		err = rows.Next(dest)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), dest[0])

		err = rows.Next(dest)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), dest[0])

		err = rows.Next(dest)
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("context cancelled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"columns": ["id"],
				"rows": [
						[1],
						[2]
				]
		}`))
		}))
		defer server.Close()

		c := &conn{
			client: transport.NewClient(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, "", ""),
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := c.QueryContext(ctx, "SELECT id from mascots", nil)
		assert.Error(t, err)
	})

	t.Run("non-200 returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := &conn{
			client: transport.NewClient(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, "", ""),
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		_, err := c.QueryContext(context.Background(), "SELECT id from mascots", nil)
		assert.Error(t, err)
	})

	t.Run("invalid params", func(t *testing.T) {
		c := &conn{
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}
		_, err := c.QueryContext(context.Background(), "SELECT id from mascots", []driver.NamedValue{{Ordinal: 1, Value: 1, Name: "id"}})
		assert.Error(t, err)
	})
}

func TestExecContext(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotPayload transport.Payload
		var gotMethod, gotContentType string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotContentType = r.Header.Get("Content-Type")
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"changes": 3}`))
		}))
		defer server.Close()

		c := &conn{
			client: transport.NewClient(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, "", ""),
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		s := &stmt{
			query: "DELETE FROM mascots WHERE id > ?",
			conn:  c,
		}

		_, err := s.Exec([]driver.Value{10})
		assert.NoError(t, err)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "application/json", gotContentType)
		assert.Equal(t, "DELETE FROM mascots WHERE id > ?", gotPayload.SQL)
		assert.True(t, gotPayload.IsExec)
		assert.Equal(t, []any{float64(10)}, gotPayload.Args)
	})

	t.Run("last insert id", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"changes": 1, "last_row_id": 42}`))
		}))
		defer server.Close()

		c := &conn{
			client: transport.NewClient(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, "", ""),
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		s := &stmt{
			query: "INSERT INTO mascots (name) VALUES (?)",
			conn:  c,
		}

		result, err := s.Exec([]driver.Value{"Gopher"})
		assert.NoError(t, err)

		lastInsertId, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Equal(t, int64(42), lastInsertId)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)
	})

	t.Run("context cancelled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"changes": 3}`))
		}))
		defer server.Close()

		c := &conn{
			client: transport.NewClient(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, "", ""),
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := c.ExecContext(ctx, "DELETE FROM mascots WHERE id > ?", []driver.NamedValue{{Ordinal: 1, Value: 10}})
		assert.Error(t, err)
	})

	t.Run("non-200 returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := &conn{
			client: transport.NewClient(server.URL, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, "", ""),
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		_, err := c.ExecContext(context.Background(), "DELETE FROM mascots WHERE id > ?", []driver.NamedValue{{Ordinal: 1, Value: 10}})
		assert.Error(t, err)
	})

	t.Run("invalid params", func(t *testing.T) {
		c := &conn{
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}
		_, err := c.ExecContext(context.Background(), "DELETE FROM mascots WHERE id > ?", []driver.NamedValue{{Ordinal: 1, Value: 2, Name: "id"}})
		assert.Error(t, err)
	})
}
