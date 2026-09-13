package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"
)

// apiClient talks to the runplanner REST API. The API authenticates with a
// session cookie, so the client keeps a cookie jar and logs in on demand.
type apiClient struct {
	baseURL  string
	email    string
	password string
	http     *http.Client

	mu       sync.Mutex
	loggedIn bool
}

func newAPIClient(baseURL, email, password string) (*apiClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &apiClient{
		baseURL:  strings.TrimRight(baseURL, "/"),
		email:    email,
		password: password,
		http:     &http.Client{Jar: jar, Timeout: 60 * time.Second},
	}, nil
}

// apiError carries the error message the API reported, so tool callers see the
// real validation failure instead of a bare status code.
type apiError struct {
	status  int
	message string
}

func (e *apiError) Error() string {
	if e.message == "" {
		return fmt.Sprintf("runplanner API returned %d", e.status)
	}
	return e.message
}

// do sends a request, logging in first and retrying once if the session is not
// (or no longer) valid.
func (c *apiClient) do(ctx context.Context, method, path string, body any, out any) error {
	if err := c.ensureLogin(ctx); err != nil {
		return err
	}
	err := c.request(ctx, method, path, body, out)
	var apiErr *apiError
	if errors.As(err, &apiErr) && apiErr.status == http.StatusUnauthorized {
		c.invalidateSession()
		if err := c.ensureLogin(ctx); err != nil {
			return err
		}
		return c.request(ctx, method, path, body, out)
	}
	return err
}

func (c *apiClient) ensureLogin(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loggedIn {
		return nil
	}
	creds := map[string]string{"email": c.email, "password": c.password}
	if err := c.request(ctx, http.MethodPost, "/api/auth/login", creds, nil); err != nil {
		return fmt.Errorf("login as %s failed: %w", c.email, err)
	}
	c.loggedIn = true
	return nil
}

func (c *apiClient) invalidateSession() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loggedIn = false
	if jar, err := cookiejar.New(nil); err == nil {
		c.http.Jar = jar
	}
}

func (c *apiClient) request(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("calling %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return &apiError{status: resp.StatusCode, message: errorMessage(raw)}
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decoding response from %s %s: %w", method, path, err)
	}
	return nil
}

// errorMessage pulls the {"error": "..."} field the API uses for failures.
func errorMessage(raw []byte) string {
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(raw, &payload); err == nil && payload.Error != "" {
		return payload.Error
	}
	return strings.TrimSpace(string(raw))
}
