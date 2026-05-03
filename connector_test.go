package god1

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := newDefaultConfig()
	err := WithLogger(logger)(c)
	assert.NoError(t, err)
	assert.Equal(t, logger, c.logger)
}

func TestWithRequestTimeout(t *testing.T) {
	timeout := 30 * time.Second
	c := newDefaultConfig()
	err := WithRequestTimeout(timeout)(c)
	assert.NoError(t, err)
	assert.Equal(t, timeout, c.clientTimeout)
}

func TestWithRequestTimeoutNegative(t *testing.T) {
	c := newDefaultConfig()
	err := WithRequestTimeout(-10 * time.Second)(c)
	assert.ErrorIs(t, err, ErrInvalidRequestTimeout)
	assert.Equal(t, defaultClientTimeout, c.clientTimeout)
}

func TestWithCloudflareAccess(t *testing.T) {
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	c := newDefaultConfig()
	err := WithCloudflareAccess(clientID, clientSecret)(c)
	assert.NoError(t, err)
	assert.Equal(t, clientID, c.cfAccessClientID)
	assert.Equal(t, clientSecret, c.cfAccessClientSecret)
}

func TestWithSQLiteVersion(t *testing.T) {
	version := "3.35.5"
	c := newDefaultConfig()
	err := WithSQLiteVersion(version)(c)
	assert.NoError(t, err)
	assert.Equal(t, version, c.sqliteVersion)
}

func TestWithSQLiteVersionEmpty(t *testing.T) {
	c := newDefaultConfig()
	err := WithSQLiteVersion("")(c)
	assert.ErrorIs(t, err, ErrInvalidSQLiteVersion)
	assert.Equal(t, defaultSQLiteVersion, c.sqliteVersion)
}

func TestNewConnectorInvalidDSN(t *testing.T) {
	_, err := NewConnector("invalid-dsn")
	assert.Error(t, err)
}

func TestNewConnectorValidDSN(t *testing.T) {
	connector, err := NewConnector("https://example.com")
	assert.NoError(t, err)
	assert.NotNil(t, connector)
}

func TestConnectorConnect(t *testing.T) {
	connector, err := NewConnector("https://example.com")
	assert.NoError(t, err)

	conn, err := connector.Connect(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestNewConnectorWithInvalidOptions(t *testing.T) {
	_, err := NewConnector("https://example.com", WithRequestTimeout(-10*time.Second))
	assert.ErrorIs(t, err, ErrInvalidRequestTimeout)

	_, err = NewConnector("https://example.com", WithCloudflareAccess("", ""))
	assert.ErrorIs(t, err, ErrInvalidCloudflareAccessCredentials)

	_, err = NewConnector("https://example.com", WithSQLiteVersion(""))
	assert.ErrorIs(t, err, ErrInvalidSQLiteVersion)
}

func TestConnectContextCancellation(t *testing.T) {
	connector, err := NewConnector("https://example.com")
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = connector.Connect(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestDriver(t *testing.T) {
	d := &d1Driver{}

	c, err := d.OpenConnector("https://example.com")
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, c.Driver())
}
