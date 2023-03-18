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
package provider

import (
	"context"

	"github.com/CrunchyData/terraform-provider-crunchybridge/internal/bridgeapi"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceClusterIDs() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source for retreiving Cluster identifiers from the user-provided label",
		ReadContext: dataSourceClusterIDsRead,
		Schema: map[string]*schema.Schema{
			// "Request" Fields
			"team_id": {
				Description: "Limits the cluster mapping to those clusters belonging to the identified team.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			// "Result / Computed Fields"
			"cluster_ids_by_name": {
				Computed:    true,
				Description: "A mapping of cluster names to their respective cluster IDs in [EID format](https://docs.crunchybridge.com/api-concepts/eid).",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceClusterIDsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bridgeapi.Client)

	account, err := client.Account(ctx)
	if err != nil {
		return diag.Errorf("failed to get account details: %v", err)
	}

	d.SetId(account.ID)

	teamID := d.Get("team_id").(string)

	var clusters []bridgeapi.ClusterDetail

	if teamID == "" {
		clusters, err = client.GetAllClusters(ctx)
		if err != nil {
			return diag.Errorf("failed to get all clusters: %v", err)
		}
	} else {
		clusters, err = client.ClustersForTeam(ctx, teamID)
		if err != nil {
			return diag.Errorf("failed to get clusters for team: %v", err)
		}
	}

	clusterMap := map[string]string{}
	for _, cluster := range clusters {
		clusterMap[cluster.Name] = cluster.ID
	}

	err = d.Set("cluster_ids_by_name", clusterMap)
	if err != nil {
		return diag.Errorf("failed to set 'cluster_ids_by_name': %v", err)
	}

	return nil
}
