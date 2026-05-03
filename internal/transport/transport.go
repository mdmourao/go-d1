package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	proxyURL             string
	logger               *slog.Logger
	httpClient           *http.Client
	cfAccessClientID     string
	cfAccessClientSecret string
}

func NewClient(url string, logger *slog.Logger, timeout time.Duration, cfAccessClientID, cfAccessClientSecret string) *Client {
	return &Client{
		proxyURL: url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger:               logger,
		cfAccessClientID:     cfAccessClientID,
		cfAccessClientSecret: cfAccessClientSecret,
	}
}

func (c *Client) do(ctx context.Context, payload Payload) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.proxyURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if c.cfAccessClientID != "" && c.cfAccessClientSecret != "" {
		req.Header.Set("CF-Access-Client-Id", c.cfAccessClientID)
		req.Header.Set("CF-Access-Client-Secret", c.cfAccessClientSecret)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warn("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) Query(ctx context.Context, sql string, args []any) ([]byte, error) {
	return c.do(ctx, Payload{SQL: sql, Args: args})
}

func (c *Client) Exec(ctx context.Context, sql string, args []any) (Response, error) {
	body, err := c.do(ctx, Payload{SQL: sql, Args: args, IsExec: true})
	if err != nil {
		return Response{}, err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return Response{}, err
	}

	return response, nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.Query(ctx, "SELECT 1", nil)
	return err
}
