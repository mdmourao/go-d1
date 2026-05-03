package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// errCloser is an io.ReadCloser whose Close always returns an error.
type errCloser struct {
	io.Reader
}

func (errCloser) Close() error { return errors.New("close failed") }

// roundTripperFunc lets a function be used as an http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNewClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second
	cfAccessClientID := "test-client-id"
	cfAccessClientSecret := "test-client-secret"

	client := NewClient("http://example.com", logger, timeout, cfAccessClientID, cfAccessClientSecret)

	assert.Equal(t, "http://example.com", client.proxyURL)
	assert.Equal(t, logger, client.logger)
	assert.Equal(t, timeout, client.httpClient.Timeout)
	assert.Equal(t, cfAccessClientID, client.cfAccessClientID)
	assert.Equal(t, cfAccessClientSecret, client.cfAccessClientSecret)
}

func TestPing(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"sql":"SELECT 1","args":null,"isExec":false}`, string(body))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, logger, timeout, "", "")

	err := client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestPing_ServerError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger, timeout, "", "")

	err := client.Ping(context.Background())
	assert.Error(t, err)
}

func TestPing_WithCFAccessHeaders(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id", r.Header.Get("CF-Access-Client-Id"))
		assert.Equal(t, "secret", r.Header.Get("CF-Access-Client-Secret"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, logger, timeout, "id", "secret")

	err := client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestExec(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"sql":"INSERT INTO test (name) VALUES (?)","args":["test"],"isExec":true}`, string(body))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"changes":1,"last_row_id":42}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, logger, timeout, "", "")

	resp, err := client.Exec(context.Background(), "INSERT INTO test (name) VALUES (?)", []any{"test"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.Changes)
	assert.Equal(t, int64(42), resp.LastRowID)
}

func TestExec_ServerError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger, timeout, "", "")

	_, err := client.Exec(context.Background(), "INSERT INTO test (name) VALUES (?)", []any{"test"})
	assert.Error(t, err)
}

func TestExec_WithCFAccessHeaders(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id", r.Header.Get("CF-Access-Client-Id"))
		assert.Equal(t, "secret", r.Header.Get("CF-Access-Client-Secret"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"changes":1,"last_row_id":42}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, logger, timeout, "id", "secret")

	resp, err := client.Exec(context.Background(), "INSERT INTO test (name) VALUES (?)", []any{"test"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.Changes)
	assert.Equal(t, int64(42), resp.LastRowID)
}

func TestExecInvalidResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL, logger, timeout, "", "")

	_, err := client.Exec(context.Background(), "INSERT INTO test (name) VALUES (?)", []any{"test"})
	assert.Error(t, err)
}

func TestQuery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"sql":"SELECT id FROM test WHERE name = ?","args":["test"],"isExec":false}`, string(body))

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

	client := NewClient(server.URL, logger, timeout, "", "")

	resp, err := client.Query(context.Background(), "SELECT id FROM test WHERE name = ?", []any{"test"})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{
        "columns": ["id"],
        "rows": [
            [1],
            [2]
        ]
    }`), resp)
}

func TestQuery_ServerError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, logger, timeout, "", "")

	_, err := client.Query(context.Background(), "SELECT id FROM test WHERE name = ?", []any{"test"})
	assert.Error(t, err)
}

func TestQuery_WithCFAccessHeaders(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	timeout := 5 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "id", r.Header.Get("CF-Access-Client-Id"))
		assert.Equal(t, "secret", r.Header.Get("CF-Access-Client-Secret"))

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

	client := NewClient(server.URL, logger, timeout, "id", "secret")

	resp, err := client.Query(context.Background(), "SELECT id FROM test WHERE name = ?", []any{"test"})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{
				"columns": ["id"],
				"rows": [
					[1],
					[2]
				]
			}`), resp)
}

func TestDo_MarshalError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := NewClient("http://example.com", logger, 5*time.Second, "", "")

	_, err := client.Query(context.Background(), "SELECT 1", []any{make(chan int)})
	assert.Error(t, err)

	var marshalErr *json.UnsupportedTypeError
	assert.ErrorAs(t, err, &marshalErr)
}

func TestDoInvalidURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := NewClient("http://[::1]:invalid", logger, 5*time.Second, "", "")

	_, err := client.Query(context.Background(), "SELECT 1", nil)
	assert.Error(t, err)
}

func TestDo_HTTPClientError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	client := NewClient(server.URL, logger, 5*time.Second, "", "")

	_, err := client.Query(context.Background(), "SELECT 1", nil)
	assert.Error(t, err)
}

func TestDo_HTTPClientError_ContextCanceled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := NewClient("http://example.com", logger, 5*time.Second, "", "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Query(ctx, "SELECT 1", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// stupid test to get coverage on the defer func that logs body close errors
func TestDo_BodyCloseError_LogsWarning(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	client := NewClient("http://example.com", logger, 5*time.Second, "", "")
	client.httpClient.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       errCloser{Reader: bytes.NewReader([]byte(`[]`))},
			Header:     make(http.Header),
		}, nil
	})

	_, err := client.Query(context.Background(), "SELECT 1", nil)
	assert.NoError(t, err)

	logs := logBuf.String()
	assert.Contains(t, logs, "failed to close response body")
	assert.Contains(t, logs, "close failed")
}
