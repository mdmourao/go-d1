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

func (c *Client) do(ctx context.Context, payload Payload) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.ProxyURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	clientID := os.Getenv("CF_ACCESS_CLIENT_ID")
	c.logger.With("client_id", clientID).Debug("Making request to transport proxy", "url", c.ProxyURL)

	req.Header.Set("CF-Access-Client-Id", clientID)
	req.Header.Set("CF-Access-Client-Secret", os.Getenv("CF_ACCESS_CLIENT_SECRET"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// todo - better error handling here
		return nil, errors.New("unexpected status code: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// TODO - handle return?
func (c *Client) Query(ctx context.Context, sql string, args []any) ([]byte, error) {
	// TODO - review logging
	c.logger.Debug("Client.Query called", "sql", sql, "args", args)
	return c.do(ctx, Payload{SQL: sql, Args: args})
}

func (c *Client) Exec(ctx context.Context, sql string, args []any) (Response, error) {
	c.logger.Debug("Client.Exec called", "sql", sql, "args", args)

	body, err := c.do(ctx, Payload{SQL: sql, Args: args, IsExec: true})
	if err != nil {
		return Response{}, err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return Response{}, err
	}

	c.logger.Debug("Received response from transport proxy", "response", response)

	return response, nil
}
