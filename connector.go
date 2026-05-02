package god1

import (
	"context"
	"database/sql/driver"
	"io"
	"log/slog"
	"time"

	"github.com/mdmourao/go-d1/internal/transport"
)

// https://pkg.go.dev/database/sql/driver#Connector
// driver.Connector

const defaultClientTimeout = 15 * time.Second
const defaultSQLiteVersion = "3.53.0"

type Option func(*config) error

type config struct {
	logger               *slog.Logger
	clientTimeout        time.Duration
	sqliteVersion        string
	cfAccessClientID     string
	cfAccessClientSecret string
}

func WithLogger(l *slog.Logger) Option {
	return func(c *config) error {
		if l != nil {
			c.logger = l
		}
		return nil
	}
}

func WithRequestTimeout(t time.Duration) Option {
	return func(c *config) error {
		if t > 0 {
			c.clientTimeout = t
		}
		return nil
	}
}

func WithCloudflareAccess(clientID, clientSecret string) Option {
	return func(c *config) error {
		c.cfAccessClientID = clientID
		c.cfAccessClientSecret = clientSecret
		return nil
	}
}

func WithSQLiteVersion(v string) Option {
	return func(c *config) error {
		if v != "" {
			c.sqliteVersion = v
		}
		return nil
	}
}

func NewConnector(dsn string, opts ...Option) (driver.Connector, error) {
	u, err := parseDSN(dsn)
	if err != nil {
		return nil, err
	}

	cfg := &config{
		logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		clientTimeout: defaultClientTimeout,
		sqliteVersion: defaultSQLiteVersion,
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return &connector{
		client:        transport.NewClient(u.String(), cfg.logger, cfg.clientTimeout, cfg.cfAccessClientID, cfg.cfAccessClientSecret),
		logger:        cfg.logger,
		sqliteVersion: cfg.sqliteVersion,
	}, nil
}

var _ driver.Connector = (*connector)(nil)

type connector struct {
	client        *transport.Client
	logger        *slog.Logger
	sqliteVersion string
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &conn{client: c.client, logger: c.logger, sqliteVersion: c.sqliteVersion}, nil
}

func (c *connector) Driver() driver.Driver { return defaultDriver }
