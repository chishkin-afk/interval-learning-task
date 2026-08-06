package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/chishkin/intask/notifier/internal/domain/user"
	"github.com/chishkin/intask/notifier/internal/infrastructure/config"
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

// GetUserByID fetches a user by their UUID from the backend.
//
// Returns user.ErrNotFound if the backend responds with 404.
// Other non-2xx responses yield *HTTPError.
func (c *Client) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	req, err := c.newGetRequest(id)
	if err != nil {
		return nil, err
	}

	c.log.Debug("sending get request into backend",
		slog.String("request", req.URL.Path),
	)

	var user userRecord
	if err := doRequest(ctx, c.client, req, &user); err != nil {
		return nil, err
	}

	return recordToUser(&user), nil
}

func (c *Client) newGetRequest(id uuid.UUID) (*http.Request, error) {
	req, err := http.NewRequest("GET", c.getUserByIDURL(id), nil)
	if err != nil {
		return nil, fmt.Errorf("can't create request: %w", err)
	}

	return req, nil
}

func (c *Client) getUserByIDURL(id uuid.UUID) string {
	return fmt.Sprintf("%s/auth/user/%s",
		c.cfg.Backend.Addr,
		id.String(),
	)
}

// doRequest executes an HTTP request and decodes the JSON response body into out.
//
// out must be a non-nil pointer to the target type. If the response status is
// 204 No Content, out is left untouched and nil is returned.
//
// Non-2xx responses yield an *HTTPError carrying the status code, allowing
// callers to branch on specific codes via errors.As.
//
// Example:
//
//	var u User
//	err := doRequest(ctx, client, req, &u)
func doRequest[T any](ctx context.Context, client *http.Client, req *http.Request, body *T) error {
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("can't do req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return &HTTPError{
			msg:  resp.Status,
			code: resp.StatusCode,
			body: string(body),
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
