package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Client struct {
	ProxyURL string
	logger   *slog.Logger
	HTTP     *http.Client
}

func NewClient(url string, logger *slog.Logger) *Client {
	return &Client{
		ProxyURL: url,
		HTTP: &http.Client{
			// TODO make this configurable?
			Timeout: 15 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) Execute(ctx context.Context, sql string, args []any) ([]byte, error) {
	c.logger.Debug("Client.Execute called", "sql", sql, "args", args)
	payload := Payload{
		SQL:  sql,
		Args: args,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.ProxyURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("CF_ACCESS_CLIENT_ID", os.Getenv("CF_ACCESS_CLIENT_ID"))
	req.Header.Set("CF_ACCESS_CLIENT_SECRET", os.Getenv("CF_ACCESS_CLIENT_SECRET"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("proxy error")
	}

	return io.ReadAll(resp.Body)
}
