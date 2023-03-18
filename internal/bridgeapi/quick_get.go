/*
Copyright 2022 Crunchy Data Solutions, Inc.

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
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (c *Client) Account(ctx context.Context) (Account, error) {
	route := c.resolve(routeAccount)

	var account Account

	err := c.doJSON(ctx, http.MethodGet, route, &account)
	if err != nil {
		return Account{}, err
	}

	return account, nil
}

func (c *Client) AccountTeams(ctx context.Context) (Teams, error) {
	route := c.resolve(routeTeams)

	var response struct {
		Teams Teams `json:"teams"`
	}

	err := c.doJSON(ctx, http.MethodGet, route, &response)
	if err != nil {
		return nil, err
	}

	return response.Teams, nil
}

func (c *Client) Providers(ctx context.Context) ([]Provider, error) {
	route := c.resolve(routeProviders)

	var response struct {
		Providers []Provider `json:"providers"`
	}

	err := c.doJSON(ctx, http.MethodGet, route, &response)
	if err != nil {
		return nil, err
	}

	return response.Providers, nil
}

func (c *Client) doJSON(
	ctx context.Context,
	method string, route *url.URL, o any,
) (outErr error) {
	resp, err := c.do(ctx, method, route, nil)
	if err != nil {
		return err
	}

	defer safeClose(&outErr, resp.Body, "response body")

	if resp.StatusCode != http.StatusOK {
		return errorFromAPIMessageResponse(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(o)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return nil
}

func (c *Client) doExec(
	ctx context.Context,
	method string, route *url.URL, statusCode int,
) (outErr error) {
	resp, err := c.do(ctx, method, route, nil)
	if err != nil {
		return err
	}

	defer safeClose(&outErr, resp.Body, "response body")

	if resp.StatusCode != statusCode {
		return errorFromAPIMessageResponse(resp)
	}

	return nil
}

func (c *Client) do(
	ctx context.Context,
	method string, route *url.URL, body io.Reader,
	opts ...func(req *http.Request),
) (*http.Response, error) {
	if err := c.login(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx, method, route.String(), body,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setCommonHeaders(req)

	for i := range opts {
		opts[i](req)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	return resp, nil
}
