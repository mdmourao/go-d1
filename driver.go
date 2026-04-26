package god1

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"

	"github.com/mdmourao/go-d1/internal/transport"
)

func init() {
	sql.Register("god1", &Driver{})
}

// https://pkg.go.dev/database/sql/driver#Driver
// driver.Driver

var _ driver.Driver = (*Driver)(nil)

type Driver struct{}

func (d *Driver) Open(name string) (driver.Conn, error) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	u, err := url.Parse(name)
	if err != nil {
		logger.Error("invalid DSN", "error", err)
		return nil, fmt.Errorf("invalid DSN: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		logger.Error("invalid DSN: scheme must be http or https", "scheme", u.Scheme)
		return nil, errors.New("invalid DSN: scheme must be http or https")
	}
	if u.Host == "" {
		logger.Error("invalid DSN: missing host")
		return nil, errors.New("invalid DSN: missing host")
	}

	q := u.Query()
	token := q.Get("token")
	if token == "" {
		logger.Error("invalid DSN: missing token (use ?token=...)")
		return nil, errors.New("invalid DSN: missing token (use ?token=...)")
	}

	debug := q.Get("debug")
	// TODO - better parsing?
	if debug == "1" || debug == "true" {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	q.Del("token")
	u.User = nil
	u.RawQuery = q.Encode()

	client := transport.NewClient(u.String(), token, logger)

	logger.Debug("Opening connection", "dsn", u.String())
	return &Conn{
		client: client,
		logger: logger,
	}, nil
}
