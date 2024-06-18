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
	"time"

	"github.com/CrunchyData/terraform-provider-crunchybridge/internal/bridgeapi"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source for retreiving Cluster resource data",
		ReadContext: dataSourceClusterRead,
		Schema: map[string]*schema.Schema{
			// "Request" Fields
			"id": {
				Description: "The unique ID of the cluster in [EID format](https://docs.crunchybridge.com/api-concepts/eid)",
				Type:        schema.TypeString,
				Required:    true,
			},
			// "Result / Computed Fields"
			"cpu": {
				Computed:    true,
				Description: "The number of CPU units on the cluster's instance",
				Type:        schema.TypeInt,
			},
			"created_at": {
				Computed:    true,
				Description: "Creation time formatted as [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339).",
				Type:        schema.TypeString,
			},
			"is_ha": {
				Computed:    true,
				Description: "Whether the cluster is high availability, meaning that it has a secondary it can fail over to quickly in case the primary becomes unavailable.",
				Type:        schema.TypeBool,
			},
			"postgres_version_id": {
				Computed:    true,
				Description: "The cluster's major Postgres version. For example, `16`.",
				Type:        schema.TypeInt,
			},
			"maintenance_window_start": {
				Computed: true,
				Description: "The hour of day which a maintenance window can possibly start. " +
					"This should be an integer from `0` to `23` representing the hour of day which " +
					"maintenance is allowed to start, with `0` representing midnight UTC. " +
					"Maintenance windows are typically three hours long starting from this " +
					"hour. A `null` value means that no explicit maintenance window has been " +
					"set and that maintenance is allowed to occur at any time.",
				Type: schema.TypeInt,
			},
			"memory": {
				Computed:    true,
				Description: "The total amount of memory available on the cluster's instance in GB (gigabytes).",
				Type:        schema.TypeFloat,
			},
			"name": {
				Computed:    true,
				Description: "A human-readable name for the cluster.",
				Type:        schema.TypeString,
			},
			"plan_id": {
				Computed:    true,
				Description: "The ID of the [cluster's plan](https://docs.crunchybridge.com/concepts/plans-pricing/). Determines instance, CPU, and memory.",
				Type:        schema.TypeString,
			},
			"provider_id": {
				Computed:    true,
				Description: "The [cloud provider](https://docs.crunchybridge.com/api/provider) where the cluster is located.",
				Type:        schema.TypeString,
			},
			"region_id": {
				Computed:    true,
				Description: "The [provider region](https://docs.crunchybridge.com/api/provider#region) where the cluster is located.",
				Type:        schema.TypeString,
			},
			"storage": {
				Computed:    true,
				Description: "The amount of storage available to the cluster in GB (gigabytes).",
				Type:        schema.TypeInt,
			},
			"team_id": {
				Computed:    true,
				Description: "The ID of the parent [team](https://docs.crunchybridge.com/concepts/teams/) for the cluster.",
				Type:        schema.TypeString,
			},
			"updated_at": {
				Computed:    true,
				Description: "Time at which the cluster was last updated formatted as [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339).",
				Type:        schema.TypeString,
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bridgeapi.Client)

	id := d.Get("id").(string)
	d.SetId(id)

	cd, err := client.ClusterDetail(id)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := []diag.Diagnostic{}

	// Apparently, this pain is reward for not having another level of indirection on the schema
	//
	// ResourceData.Set() only understands setting map values, so if the schema matches the return value
	// from the API, you have to map each field yourself, which is decent for decoupling but inconvenient.
	//
	// Maybe one day auto-populate with json struct tag <-> schema mapping?
	err = d.Set("id", cd.ID)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("cpu", cd.CPU)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("created_at", cd.Created.Format(time.RFC3339))
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("is_ha", cd.HighAvailability)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("postgres_version_id", cd.PGMajorVersion)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("maintenance_window_start", cd.MaintWindowStart)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("memory", cd.MemoryGB)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("name", cd.Name)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("plan_id", cd.PlanID)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("provider_id", cd.ProviderID)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("region_id", cd.RegionID)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("storage", cd.StorageGB)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("team_id", cd.TeamID)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("updated_at", cd.Updated.Format(time.RFC3339))
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diag.Diagnostics(diags)
}
