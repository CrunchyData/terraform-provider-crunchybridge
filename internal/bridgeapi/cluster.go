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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

func (c *Client) CreateCluster(ctx context.Context, cr CreateRequest) (_ string, outErr error) {
	route := c.resolve(routeClusters)

	reqPayload, err := json.Marshal(cr)
	if err != nil {
		return "", fmt.Errorf("error during cluster request encoding: %w", err)
	}

	idempotencyOpt := func(req *http.Request) {
		if c.useIdempotencyKey {
			// Set Idempotency Key based on payload content
			//
			// API is expecting UUID for the value, but we're using
			// UUIDv5 so that the key matches the request payload.
			idemKey := uuid.NewSHA1(BridgeProviderNS, reqPayload)

			req.Header.Set("Idempotency-Key", idemKey.String())
		}
	}

	resp, err := c.do(
		ctx, http.MethodPost, route,
		bytes.NewReader(reqPayload), idempotencyOpt,
	)
	if err != nil {
		return "", err
	}

	defer safeClose(&outErr, resp.Body, "response body")

	if resp.StatusCode != http.StatusCreated {
		return "", errorFromAPIMessageResponse(resp)
	}

	var idOnly struct {
		ID string `json:"id"`
	}

	err = json.NewDecoder(resp.Body).Decode(&idOnly)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return idOnly.ID, nil
}

func (c *Client) DeleteCluster(ctx context.Context, id string) error {
	route := c.resolve(fmt.Sprintf(routeCluster, id))

	return c.doExec(ctx, http.MethodDelete, route, http.StatusOK)
}

func (c *Client) ClusterDetail(ctx context.Context, id string) (ClusterDetail, error) {
	route := c.resolve(fmt.Sprintf(routeCluster, id))

	var detail ClusterDetail

	err := c.doJSON(ctx, http.MethodGet, route, &detail)
	if err != nil {
		return ClusterDetail{}, err
	}

	return detail, nil
}

func (c *Client) ClusterStatus(ctx context.Context, id string) (ClusterStatus, error) {
	route := c.resolve(fmt.Sprintf(routeClusterStatus, id))

	var status ClusterStatus

	err := c.doJSON(ctx, http.MethodGet, route, &status)
	if err != nil {
		return ClusterStatus{}, err
	}

	return status, nil
}

func (c *Client) ClusterRoles(ctx context.Context, id string) ([]ClusterRole, error) {
	roles := []string{"postgres", "application"}

	defaultRoles := make([]ClusterRole, len(roles))

	for i, role := range roles {
		route := c.resolve(fmt.Sprintf(routeClusterRole, id, role))

		var roleInfo ClusterRole

		err := c.doJSON(ctx, http.MethodGet, route, &roleInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster role %q: %w", role, err)
		}

		defaultRoles[i] = roleInfo
	}

	return defaultRoles, nil
}

func (c *Client) ClustersForTeam(
	ctx context.Context, teamID string,
) ([]ClusterDetail, error) {
	route := c.resolve(routeClusters, url.Values{
		"team_id": []string{teamID},
	})

	var details struct {
		Clusters []ClusterDetail `json:"clusters"`
	}

	err := c.doJSON(ctx, http.MethodGet, route, &details)
	if err != nil {
		return nil, err
	}

	return details.Clusters, nil
}

func (c *Client) GetAllClusters(ctx context.Context) ([]ClusterDetail, error) {
	teams, err := c.AccountTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get team memberships: %w", err)
	}

	var allClusters []ClusterDetail

	for _, team := range teams {
		teamClusters, err := c.ClustersForTeam(ctx, team.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get clusters for team %s: %w", team.ID, err)
		}

		allClusters = append(allClusters, teamClusters...)
	}

	return allClusters, nil
}

func (c *Client) UpdateCluster(
	ctx context.Context, id string, ur ClusterUpdateRequest,
) (outErr error) {
	route := c.resolve(fmt.Sprintf(routeCluster, id))

	reqPayload, err := json.Marshal(ur)
	if err != nil {
		return fmt.Errorf("error during cluser update encoding: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPatch, route, bytes.NewReader(reqPayload))
	if err != nil {
		return err
	}

	defer safeClose(&outErr, resp.Body, "response body")

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	default:
		return errorFromAPIMessageResponse(resp)
	}
}

func (c *Client) UpgradeCluster(
	ctx context.Context, id string, ur ClusterUpgradeRequest,
) (outErr error) {
	route := c.resolve(fmt.Sprintf(routeClusterUpgrade, id))

	reqPayload, err := json.Marshal(ur)
	if err != nil {
		return fmt.Errorf("error during cluser update encoding: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, route, bytes.NewReader(reqPayload))
	if err != nil {
		return err
	}

	defer safeClose(&outErr, resp.Body, "response body")

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	default:
		return errorFromAPIMessageResponse(resp)
	}
}
