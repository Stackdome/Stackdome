// Package turnstile verifies Cloudflare Turnstile challenge tokens.
package turnstile

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const (
	CanonicalSiteverifyEndpoint       = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	maxTokenLength                    = 2048
	maxResponseBytes            int64 = 64 * 1024
)

// Verifier verifies one Turnstile token for a client IP address.
type Verifier interface {
	Verify(context.Context, string, netip.Addr) error
}

// ClientSpec contains the secret verification settings that are never exposed
// through the public application configuration endpoint.
type ClientSpec struct {
	Secret           string
	ExpectedHostname string
	ExpectedAction   string
	Timeout          time.Duration
}

type client struct {
	secret           string
	expectedHostname string
	expectedAction   string
	endpoint         string
	httpClient       *http.Client
}

// NewClient constructs a verifier that always uses Cloudflare's canonical
// siteverify endpoint.
func NewClient(spec ClientSpec) (Verifier, error) {
	if strings.TrimSpace(spec.Secret) == "" {
		return nil, errors.New("turnstile secret is required")
	}
	if strings.TrimSpace(spec.ExpectedHostname) == "" {
		return nil, errors.New("turnstile expected hostname is required")
	}
	if strings.TrimSpace(spec.ExpectedAction) == "" {
		return nil, errors.New("turnstile expected action is required")
	}
	if spec.Timeout <= 0 {
		return nil, errors.New("turnstile verification timeout must be positive")
	}

	return newClient(spec, CanonicalSiteverifyEndpoint), nil
}

func newClient(spec ClientSpec, endpoint string) Verifier {
	return &client{
		secret:           spec.Secret,
		expectedHostname: spec.ExpectedHostname,
		expectedAction:   spec.ExpectedAction,
		endpoint:         endpoint,
		httpClient: &http.Client{
			Timeout: spec.Timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

type siteverifyResponse struct {
	Success  bool   `json:"success"`
	Hostname string `json:"hostname"`
	Action   string `json:"action"`
}

func (c *client) Verify(ctx context.Context, token string, remoteIP netip.Addr) error {
	if token == "" || len(token) > maxTokenLength || !remoteIP.IsValid() {
		return errors.New("turnstile verification failed: invalid input")
	}

	form := url.Values{
		"secret":   {c.secret},
		"response": {token},
		"remoteip": {remoteIP.Unmap().String()},
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return errors.New("turnstile verification failed: request creation failed")
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return errors.New("turnstile verification failed: request failed")
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return errors.New("turnstile verification failed: unexpected HTTP status")
	}

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return errors.New("turnstile verification failed: response read failed")
	}
	if int64(len(responseBody)) > maxResponseBytes {
		return errors.New("turnstile verification failed: response too large")
	}

	var result siteverifyResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return errors.New("turnstile verification failed: malformed response")
	}
	if !result.Success {
		return errors.New("turnstile verification failed: challenge rejected")
	}
	if result.Hostname != c.expectedHostname {
		return errors.New("turnstile verification failed: hostname mismatch")
	}
	if result.Action != c.expectedAction {
		return errors.New("turnstile verification failed: action mismatch")
	}

	return nil
}

var _ Verifier = (*client)(nil)
