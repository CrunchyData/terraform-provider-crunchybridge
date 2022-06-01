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
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) Account() (Account, error) {
	if err := c.login(); err != nil {
		return Account{}, err
	}

	route := fmt.Sprint(c.apiTarget, routeAccount)

	req, err := http.NewRequest(http.MethodGet, route, nil)
	if err != nil {
		return Account{}, fmt.Errorf("during account detail request: %w", err)
	}
	c.setCommonHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return Account{}, fmt.Errorf("during account detail request prep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Account{}, fmt.Errorf("unexpected response status from API, status: %d", resp.StatusCode)
	}

	var acct Account
	err = json.NewDecoder(resp.Body).Decode(&acct)
	if err != nil {
		return Account{}, fmt.Errorf("error unmarshaling response body (account detail): %w", err)
	}

	return acct, nil
}

func (c *Client) AccountTeams() (Teams, error) {
	if err := c.login(); err != nil {
		return []Team{}, err
	}

	route := fmt.Sprint(c.apiTarget, routeTeams)

	req, err := http.NewRequest(http.MethodGet, route, nil)
	if err != nil {
		return []Team{}, fmt.Errorf("during account detail request: %w", err)
	}
	c.setCommonHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return []Team{}, fmt.Errorf("during account detail request prep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Team{}, fmt.Errorf("unexpected response status from API, status: %d", resp.StatusCode)
	}

	response := map[string][]Team{
		"teams": {},
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return []Team{}, fmt.Errorf("error unmarshaling response body (account detail): %w", err)
	}

	list := response["teams"]
	return list, nil
}

func (c *Client) Providers() ([]Provider, error) {
	if err := c.login(); err != nil {
		return []Provider{}, err
	}

	route := fmt.Sprint(c.apiTarget, routeProviders)

	req, err := http.NewRequest(http.MethodGet, route, nil)
	if err != nil {
		return []Provider{}, fmt.Errorf("during provider detail request: %w", err)
	}
	c.setCommonHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return []Provider{}, fmt.Errorf("during provider detail request prep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Provider{}, fmt.Errorf("unexpected response status from API, status: %d", resp.StatusCode)
	}

	response := map[string][]Provider{
		"providers": {},
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return []Provider{}, fmt.Errorf("error unmarshaling response body (provider list): %w", err)
	}

	list := response["providers"]
	return list, nil
}
