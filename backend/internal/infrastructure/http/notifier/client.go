package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	"github.com/google/uuid"
)

type Client struct {
	cfg    *config.Config
	log    *slog.Logger
	client *http.Client
}

// New creates a backend Client. It returns an error if required dependencies
// are missing, failing fast rather than panicking on first use.
func New(cfg *config.Config, log *slog.Logger, client *http.Client) *Client {
	return &Client{
		cfg:    cfg,
		log:    log,
		client: client,
	}
}

// SendMsg delivers msg to the notifier service for the given userID.
// It returns an *HTTPError (matchable via errors.As) on non-2xx responses.
func (c *Client) SendMsg(ctx context.Context, userID uuid.UUID, msg string) error {
	req, err := c.newSendRequest(userID, msg)
	if err != nil {
		return err
	}

	return doRequest(ctx, c.client, req, &struct{}{})
}

// TODO: вынести маршал в mapper
func (c *Client) newSendRequest(userID uuid.UUID, msg string) (*http.Request, error) {
	body, err := json.Marshal(sendMsg{
		UserID: userID,
		Msg:    msg,
	})
	if err != nil {
		return nil, fmt.Errorf("can't marshal body: %w", err)
	}

	req, err := http.NewRequest("POST", c.sendMsgURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("can't create request: %w", err)
	}

	return req, nil
}

func (c *Client) sendMsgURL() string {
	return fmt.Sprintf("%s/send", c.cfg.Notifier.Addr)
}

// TODO: разделить на две функции: с бади и без
func doRequest[T any](ctx context.Context, client *http.Client, req *http.Request, body *T) error {
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("can't do req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		errBody, _ := io.ReadAll(resp.Body)
		return &HTTPError{
			msg:  resp.Status,
			code: resp.StatusCode,
			body: string(errBody),
		}
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(body); err != nil {
		return fmt.Errorf("can't decode resp body: %w", err)
	}

	return nil
}
