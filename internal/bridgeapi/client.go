/*
Copyright 2021 Crunchy Data Solutions, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package bridgeapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	routeAccount        = "/account"
	routeCluster        = "/clusters/%s"
	routeClusterUpgrade = "/clusters/%s/upgrade"
	routeClusters       = "/clusters"
	routeClusterRole    = "/clusters/%s/roles/%s"
	routeClusterStatus  = "/clusters/%s/status"
	routeProviders      = "/providers"
	routeTeams          = "/teams"
)

var (
	BridgeProviderNS = uuid.MustParse("cc67b0e5-7152-4d54-85ff-49a5c17fbbfe")

	// Maximum time construct for Golang
	// Unix time uses an offset of 62135596801 to cover pre-start-of-epoch times
	maxTime = time.Unix(1<<63-62135596801, 999999999)
)

type ClientOption func(*Client) error

type Client struct {
	sync.RWMutex
	activeToken       string
	activeTokenID     string
	apiTarget         *url.URL
	client            *http.Client
	credential        Login
	legacyAuth        bool
	useIdempotencyKey bool
	userAgent         string
	tokenExpires      time.Time
}

func NewClient(apiURL *url.URL, cred Login, opts ...ClientOption) (*Client, error) {
	if apiURL == nil {
		return nil, errors.New("cannot create client to nil URL target")
	}

	// Defaults unless overridden by options
	c := &Client{
		apiTarget:  apiURL,
		client:     &http.Client{},
		credential: cred,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("error during client initialization: %w", err)
		}
	}

	return c, nil
}

// WithHTTPClient allows the use of a custom-configured HTTP client for API
// requests, Client defaults to a default http.Client{} otherwise
// Setter - always returns nil error
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) error {
		c.client = hc
		return nil
	}
}

// WithUserAgent configures a UserAgent string to use in all requests to the API
// Setter - always returns nil error
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) error {
		c.userAgent = ua
		return nil
	}
}

// WithImmediateLogin triggers a login instead of waiting for lazy-initialization
// to occcur once a data function is called
func WithImmediateLogin() ClientOption {
	return func(c *Client) error {
		err := c.login()
		return err
	}
}

// WithContext allows the client to be aware of a parent context. Currently, it
// is used to invalidate the access token when the provided context closes
func WithContext(ctx context.Context) ClientOption {
	return func(c *Client) error {
		go func() {
			// Wait for context.Done() signal
			<-ctx.Done()

			// since we're async, this error can't go anywhere useful, but still dump to stderr at least
			if err := c.logout(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to logout: %s", err)
			}
		}()

		return nil
	}
}

// WithIdempotencyKey causes the client to send an Idempotency Key header on cluster create
// N.B. This may have unexpected behavior tied to cached responses after system state
// changes invalidate the correctness of those responses
func WithIdempotencyKey() ClientOption {
	return func(c *Client) error {
		c.useIdempotencyKey = true
		return nil
	}
}

// WithTokenExchange instructs the client to force a token exchange for a short-
// lived token (gen 1 auth) instead of the user-managed bearer token (gen 2 auth)
func WithTokenExchange() ClientOption {
	return func(c *Client) error {
		c.legacyAuth = true
		return nil
	}
}

func (c *Client) login() error {
	// No-op if already logged in, maybe add forced login later for error handling
	c.RLock()
	tokenCurrent := (c.activeToken != "" && time.Until(c.tokenExpires) > 0)
	c.RUnlock()
	if tokenCurrent {
		return nil
	}

	if c.legacyAuth {
		req, err := http.NewRequest(http.MethodPost, c.apiTarget.String()+"/access-tokens", nil)
		if err != nil {
			return fmt.Errorf("error creating token login request: %w", err)
		}
		req.SetBasicAuth(c.credential.Key, c.credential.Secret)

		// Ensure only one attempting to refresh token
		c.Lock()
		defer c.Unlock()

		resp, err := c.client.Do(req)
		if err != nil {
			return fmt.Errorf("error submitting login request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("API returned status %d for login [%s]", resp.StatusCode, c.credential.Key)
		} else if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("API returned unexpected response %d for login [%s]", resp.StatusCode, c.credential.Key)
		}

		var tr tokenResponse
		err = json.NewDecoder(resp.Body).Decode(&tr)
		if err != nil {
			return fmt.Errorf("error unmarshaling token response body: %w", err)
		}

		c.activeToken = tr.Token
		c.activeTokenID = tr.TokenID
		c.tokenExpires = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	} else {
		if !strings.HasPrefix(c.credential.Secret, "cbkey_") {
			return ErrorOldSecretFormat
		}

		c.activeToken = c.credential.Secret
		c.tokenExpires = maxTime
	}

	return nil
}

// logout allows the authentication system to release the session
func (c *Client) logout() error {
	// No-op if already not logged in or token already expired
	c.RLock()
	tokenCurrent := (c.activeToken != "" && time.Until(c.tokenExpires) > 0)
	c.RUnlock()
	if !c.legacyAuth || !tokenCurrent {
		c.activeToken = ""
		return nil
	}

	route := fmt.Sprintf("%s%s/%s", c.apiTarget, "/access-tokens", c.activeTokenID)

	req, err := http.NewRequest(http.MethodDelete, route, nil)
	if err != nil {
		return fmt.Errorf("error creating token delete request: %w", err)
	}
	// Ensure lock occurs after here, since obtaining RLock will fail/deadlock otherwise
	c.setCommonHeaders(req)

	// Ensure only one attempting to delete token
	c.Lock()
	defer c.Unlock()

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("error submitting delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned unexpected response %d for login [%s]", resp.StatusCode, c.credential.Key)
	}

	c.activeToken = ""
	c.activeTokenID = ""
	c.tokenExpires = time.Now().Add(-1 * time.Second) // move to clear < 0 range of comparison

	return nil
}

// Close allows an explicit request to log out of the current session
// There is no explicit login, as login is triggered for every client call
// to ensure an active session state.
func (c *Client) Close() error {
	// Right now, needs nothing more than invalidating the access token
	return c.logout()
}

// helper to set up auth with current bearer token
func (c *Client) setRequestBearer(req *http.Request) {
	c.RLock()
	defer c.RUnlock()
	req.Header.Set("Authorization", "Bearer "+c.activeToken)
}

func (c *Client) setRequestUserAgent(req *http.Request) {
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
}

// helper to ensure headers used in all requests are set consistently
func (c *Client) setCommonHeaders(req *http.Request) {
	c.setRequestBearer(req)
	c.setRequestUserAgent(req)
}

func (c *Client) resolve(path string, params ...url.Values) *url.URL {
	u := c.apiTarget.ResolveReference(&url.URL{Path: path})

	q := u.Query()

	for _, p := range params {
		for name, value := range p {
			q[name] = value
		}
	}

	u.RawQuery = q.Encode()

	return u
}
