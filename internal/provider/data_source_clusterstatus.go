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

func dataSourceStatus() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source for retreiving Cluster resource data",
		ReadContext: dataSourceStatusRead,
		Schema: map[string]*schema.Schema{
			// "Request" Fields
			"id": {
				Description: "The unique ID of the cluster in [EID format](https://docs.crunchybridge.com/api-concepts/eid).",
				Type:        schema.TypeString,
				Required:    true,
			},
			// "Result / Computed Fields"
			"oldest_backup": {
				Computed:    true,
				Description: "The cluster's oldest backup. May be null if no backup has occurred yet.",
				Type:        schema.TypeString,
			},
			"state": {
				Computed:    true,
				Description: "The state of the cluster. `creating`, `destroying`, `ready`, or `restarting`",
				Type:        schema.TypeString,
			},
			"disk_available_mb": {
				Computed:    true,
				Description: "Available disk space remaining in MB (megabytes).",
				Type:        schema.TypeInt,
			},
			"disk_total_size_mb": {
				Computed:    true,
				Description: "Total disk size in MB (megabytes).",
				Type:        schema.TypeInt,
			},
			"disk_used_mb": {
				Computed:    true,
				Description: "Amount of disk currently in use in MB (megabytes).",
				Type:        schema.TypeInt,
			},
			"operations": {
				Computed:    true,
				Description: "An ongoing upgrade operation (like a version upgrade or resize) within a database cluster.",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flavor": {
							Computed:    true,
							Description: "The kind of upgrade. [Enum](https://docs.crunchybridge.com/api/cluster/#cluster-upgrade-operation)",
							Type:        schema.TypeString,
						},
						"state": {
							Computed:    true,
							Description: "The state of the ongoing operation. [Enum](https://docs.crunchybridge.com/api/cluster/#cluster-upgrade-operation)",
							Type:        schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func dataSourceStatusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bridgeapi.Client)

	id := d.Get("id").(string)
	d.SetId(id)

	cs, err := client.ClusterStatus(ctx, id)
	if err != nil {
		return diag.Errorf("failed to get cluster status: %v", err)
	}

	var diags diag.Diagnostics

	set := map[string]interface{}{
		"state":              cs.State,
		"disk_available_mb":  cs.DiskUsage.Available,
		"disk_total_size_mb": cs.DiskUsage.Total,
		"disk_used_mb":       cs.DiskUsage.Used,
	}

	for k, v := range set {
		err := d.Set(k, v)
		if err != nil {
			diags = append(diags, diag.Errorf(
				"failed to set %q: %v", k, err,
			)...)
		}
	}

	// Don't set oldest_backup if time zero value (i.e. was originally nulled at API level)
	if !cs.OldestBackup.IsZero() {
		err = d.Set("oldest_backup", cs.OldestBackup.Format(time.RFC3339))
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if len(cs.OngoingUpgrade.Operations) > 0 {
		var updates []interface{}

		for _, step := range cs.OngoingUpgrade.Operations {
			updates = append(updates, map[string]interface{}{
				"flavor": step.Flavor,
				"state":  step.State,
			})
		}

		err = d.Set("operations", updates)
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	return diags
}
