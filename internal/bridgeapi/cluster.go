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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

func (c *Client) CreateCluster(cr CreateRequest) (string, error) {
	if err := c.login(); err != nil {
		return "", err
	}

	reqPayload, err := json.Marshal(cr)
	if err != nil {
		return "", fmt.Errorf("error during cluser request encoding: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.apiTarget.String()+routeClusters, bytes.NewReader(reqPayload))
	if err != nil {
		return "", fmt.Errorf("during create cluster request: %w", err)
	}

	c.setCommonHeaders(req)

	// Set Idempotency Key based on payload content
	//
	// API is expecting UUID for the value, but we're using UUIDv5 so that
	// the key matches the request payload
	if c.useIdempotencyKey {
		idemKey := uuid.NewSHA1(BridgeProviderNS, reqPayload)
		req.Header.Set("Idempotency-Key", idemKey.String())
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("during create cluster: %w", err)
	}
	defer resp.Body.Close()

	// APIMessage is the default response format when the API function doesn't
	// return the documented response type
	var mesg APIMessage
	if resp.StatusCode != http.StatusCreated {
		err = json.NewDecoder(resp.Body).Decode(&mesg)
		if err != nil {
			// Move forward with errors based on http code
			mesg.Message = "unable to retrieve further error details"
		}
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		var idOnly struct {
			ID string `json:"id"`
		}
		err = json.NewDecoder(resp.Body).Decode(&idOnly)
		if err != nil {
			return "", fmt.Errorf("unable to retrieve cluster ID from successful create response: %w", err)
		} else {
			return idOnly.ID, nil
		}
	case http.StatusBadRequest:
		return "", fmt.Errorf("create API bad request message %w: %s, request_id: %s", ErrorBadRequest, mesg.Message, mesg.RequestID)
	case http.StatusConflict:
		return "", fmt.Errorf("create API conflict message %w: %s, request_id: %s", ErrorConflict, mesg.Message, mesg.RequestID)
	default:
		return "", fmt.Errorf("unrecognized return status from create call, code: %d, message: %s", resp.StatusCode, mesg.Message)
	}
}

func (c *Client) DeleteCluster(id string) error {
	if err := c.login(); err != nil {
		return err
	}

	route := fmt.Sprintf("%s%s/%s", c.apiTarget, routeClusters, id)

	req, err := http.NewRequest(http.MethodDelete, route, nil)
	if err != nil {
		return fmt.Errorf("during cluster delete request: %w", err)
	}
	c.setCommonHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("during cluster delete request prep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status from API, status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) ClusterDetail(id string) (ClusterDetail, error) {
	if err := c.login(); err != nil {
		return ClusterDetail{}, err
	}

	route := fmt.Sprintf("%s%s/%s", c.apiTarget, routeClusters, id)

	req, err := http.NewRequest(http.MethodGet, route, nil)
	if err != nil {
		return ClusterDetail{}, fmt.Errorf("during cluster detail request: %w", err)
	}
	c.setCommonHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return ClusterDetail{}, fmt.Errorf("during cluster detail request prep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ClusterDetail{}, fmt.Errorf("unexpected response status from API, status: %d", resp.StatusCode)
	}

	var detail ClusterDetail
	err = json.NewDecoder(resp.Body).Decode(&detail)
	if err != nil {
		return ClusterDetail{}, fmt.Errorf("error unmarshaling response body (cluster detail): %w", err)
	}

	return detail, nil
}

func (c *Client) ClusterStatus(id string) (ClusterStatus, error) {
	if err := c.login(); err != nil {
		return ClusterStatus{}, err
	}

	statusRoute := fmt.Sprintf(routeClusterStatus, id)
	route := fmt.Sprint(c.apiTarget, statusRoute)

	req, err := http.NewRequest(http.MethodGet, route, nil)
	if err != nil {
		return ClusterStatus{}, fmt.Errorf("during cluster status request: %w", err)
	}
	c.setCommonHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return ClusterStatus{}, fmt.Errorf("during cluster status request prep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ClusterStatus{}, fmt.Errorf("unexpected response status from API, status: %d", resp.StatusCode)
	}

	var status ClusterStatus
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		return ClusterStatus{}, fmt.Errorf("error unmarshaling response body (cluster status): %w", err)
	}

	return status, nil
}

func (c *Client) ClusterRoles(id string) ([]ClusterRole, error) {
	if err := c.login(); err != nil {
		return []ClusterRole{}, err
	}

	roleRoute := fmt.Sprintf(routeClusterRole, id)
	roles := []string{"postgres", "application"}

	defaultRoles := make([]ClusterRole, 0, 2)
	for _, role := range roles {
		route := fmt.Sprintf("%s%s/%s", c.apiTarget, roleRoute, role)

		req, err := http.NewRequest(http.MethodGet, route, nil)
		if err != nil {
			return []ClusterRole{}, fmt.Errorf("during cluster role [%s] request: %w", role, err)
		}
		c.setCommonHeaders(req)

		resp, err := c.client.Do(req)
		if err != nil {
			return []ClusterRole{}, fmt.Errorf("during cluster role [%s] request prep: %w", role, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return []ClusterRole{}, fmt.Errorf("unexpected response status from API, role: %s, status: %d", role, resp.StatusCode)
		}

		var roleInfo ClusterRole
		err = json.NewDecoder(resp.Body).Decode(&roleInfo)
		if err != nil {
			return []ClusterRole{}, fmt.Errorf("error unmarshaling response body (cluster role: %s): %w", role, err)
		}
		defaultRoles = append(defaultRoles, roleInfo)
	}

	return defaultRoles, nil
}

func (c *Client) ClustersForTeam(team_id string) ([]ClusterDetail, error) {
	if err := c.login(); err != nil {
		return []ClusterDetail{}, err
	}

	route := fmt.Sprint(c.apiTarget, routeClusters)

	req, err := http.NewRequest(http.MethodGet, route, nil)
	if err != nil {
		return []ClusterDetail{}, fmt.Errorf("during get clusters request: %w", err)
	}
	c.setCommonHeaders(req)

	params := url.Values{}
	params.Add("team_id", team_id)
	req.URL.RawQuery = params.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return []ClusterDetail{}, fmt.Errorf("during get clusters request prep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []ClusterDetail{}, fmt.Errorf("unexpected response status from API, status: %d", resp.StatusCode)
	}

	details := struct {
		Clusters []ClusterDetail
	}{}
	err = json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		return []ClusterDetail{}, fmt.Errorf("error unmarshaling response body (get clusters): %w", err)
	}

	return details.Clusters, nil
}

func (c *Client) GetAllClusters() ([]ClusterDetail, error) {
	teams, err := c.AccountTeams()
	if err != nil {
		return []ClusterDetail{}, fmt.Errorf("error while querying team membership: %w", err)
	}

	allClusters := []ClusterDetail{}

	for _, team := range teams {
		teamClusters, err := c.ClustersForTeam(team.ID)
		if err != nil {
			return []ClusterDetail{}, fmt.Errorf("error while querying clusters for team %s: %w", team.ID, err)
		}

		allClusters = append(allClusters, teamClusters...)
	}

	return allClusters, nil
}

func (c *Client) UpdateCluster(id string, ur ClusterUpdateRequest) error {
	if err := c.login(); err != nil {
		return err
	}

	route := fmt.Sprintf("%s%s/%s", c.apiTarget, routeClusters, id)

	reqPayload, err := json.Marshal(ur)
	if err != nil {
		return fmt.Errorf("error during cluser update encoding: %w", err)
	}
	req, err := http.NewRequest(http.MethodPatch, route, bytes.NewReader(reqPayload))
	if err != nil {
		return fmt.Errorf("during cluster update request: %w", err)
	}
	c.setCommonHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("during cluster update request prep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected response status from API, status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) UpgradeCluster(id string, ur ClusterUpgradeRequest) error {
	if err := c.login(); err != nil {
		return err
	}

	route := fmt.Sprintf("%s%s/%s/upgrade", c.apiTarget, routeClusters, id)

	reqPayload, err := json.Marshal(ur)
	if err != nil {
		return fmt.Errorf("error during cluser update encoding: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, route, bytes.NewReader(reqPayload))
	if err != nil {
		return fmt.Errorf("during cluster update request: %w", err)
	}
	c.setCommonHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("during cluster update request prep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected response status from API, status: %d", resp.StatusCode)
	}

	return nil
}
