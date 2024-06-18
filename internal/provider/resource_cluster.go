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

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func staticDefault(value interface{}) func() (interface{}, error) {
	return func() (interface{}, error) {
		return value, nil
	}
}

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Cluster resource for the Crunchy Bridge Terraform Provider",

		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			// "Request" Fields
			"name": {
				Description:  "A human-readable name for the cluster.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(5, 50),
			},
			"plan_id": {
				// Since DefaultFunc works on required fields, using Required label to reflect API requirements
				// and using static default definitions where we provide useful default values for ease-of-use
				DefaultFunc: staticDefault("hobby-2"),
				Description: "The ID of the [cluster's plan](https://docs.crunchybridge.com/concepts/plans-pricing/). Determines instance, CPU, and memory. Defaults to `hobby-2`.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"provider_id": {
				DefaultFunc:  staticDefault("aws"),
				Description:  "The [cloud provider](https://docs.crunchybridge.com/api/provider) where the cluster is located. Defaults to `aws`, allows `aws`, `gcp`, or `azure`",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"aws", "gcp", "azure"}, false),
			},
			"region_id": {
				DefaultFunc: staticDefault("us-west-1"),
				Description: "The [provider region](https://docs.crunchybridge.com/api/provider#region) where the cluster is located. Defaults to `us-west-1`",
				Required:    true,
				Type:        schema.TypeString,
			},
			"storage": {
				DefaultFunc:  staticDefault(100),
				Description:  "The amount of storage available to the cluster in GB (gigabytes). Defaults to 100.",
				Required:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(10),
			},
			"team_id": {
				Description:  "The ID of the parent [team](https://docs.crunchybridge.com/concepts/teams/) for the cluster.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(26, 26),
			},
			"is_ha": {
				// is_ha is set as required to ensure it exists in the state as non-nil for change detection, otherwise the terraform
				// difference engine is not guaranteed to identify a transition from unset/nil to true
				DefaultFunc: staticDefault(false),
				Description: "Whether the cluster is high availability, meaning that it has a secondary it can fail over to quickly in case the primary becomes unavailable. Defaults to `false`",
				Required:    true,
				Type:        schema.TypeBool,
			},
			"major_version": {
				Default:      16,
				Description:  "The cluster's major Postgres version. For example, `16`. Defaults to [Create Cluster](https://docs.crunchybridge.com/api/cluster/#create-cluster) defaults.",
				Optional:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(14),
			},
			"wait_until_ready": {
				Description: "Treats the create operation as incomplete until the cluster reports a ready status. Defaults to `false`",
				Optional:    true,
				Type:        schema.TypeBool,
			},
			// "Result / Computed Fields"
			"id": {
				Computed:    true,
				Description: "The unique ID of the cluster in [EID format](https://docs.crunchybridge.com/api-concepts/eid)",
				Type:        schema.TypeString,
			},
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
			"updated_at": {
				Computed:    true,
				Description: "Time at which the cluster was last updated formatted as [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339).",
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bridgeapi.Client)
	diags := []diag.Diagnostic{}

	tflog.Trace(ctx, "creating a cluster resource")

	req := bridgeapi.CreateRequest{
		HighAvailability: d.Get("is_ha").(bool),
		Name:             d.Get("name").(string),
		PGMajorVersion:   d.Get("major_version").(int),
		Plan:             d.Get("plan_id").(string),
		Provider:         d.Get("provider_id").(string),
		Region:           d.Get("region_id").(string),
		StorageGB:        d.Get("storage").(int),
		TeamID:           d.Get("team_id").(string),
	}

	tflog.Trace(ctx, "sending cluster resource create request to API")

	id, err := client.CreateCluster(req)
	if err != nil {
		return diag.Errorf("failed to create cluster: %s", err)
	}

	d.SetId(id)

	tflog.Trace(ctx, "successfully submitted create request")

	if waitReady := d.Get("wait_until_ready").(bool); waitReady {
		delay := 10 * time.Second // Set to terraform's notification status interval on create
		for ready, elapsed := false, time.Duration(0); !ready; elapsed += delay {
			status, err := client.ClusterStatus(id)
			if err != nil {
				tflog.Error(ctx, "error obtaining cluster ready status", map[string]interface{}{
					"error": err,
					"time":  elapsed.String(),
				})
			}
			ready = (status.State == "ready")
			if !ready {
				// terraform handles showing elapsed time, we don't need to here
				time.Sleep(delay)
			} else {
				tflog.Debug(ctx, "Completed waiting on cluster ready, "+elapsed.String()+" elapsed.")
			}
		}
	}

	readDiag := resourceClusterRead(ctx, d, meta)
	diags = append(diags, readDiag...)

	return diags
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bridgeapi.Client)

	id := d.Get("id").(string)

	cd, err := client.ClusterDetail(id)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := []diag.Diagnostic{}

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
	err = d.Set("major_version", cd.PGMajorVersion)
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

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bridgeapi.Client)
	diags := []diag.Diagnostic{}

	clusterID := d.Id()

	noUpgSupport := []string{"provider_id", "region_id", "team_id", "wait_until_ready"}
	for _, key := range noUpgSupport {
		if d.HasChange(key) {
			diags = append(diags, diag.Errorf("provider does not support in-place update for [%s]", key)...)
		}
	}
	// If unsupported fields have changed, error out so the user can correct them before applying good changes
	if len(diags) > 0 {
		return diags
	}

	// Update call on client
	if d.HasChange("name") {
		newName := d.Get("name").(string)
		err := client.UpdateCluster(clusterID, bridgeapi.ClusterUpdateRequest{
			Name: &newName,
		})
		if err != nil {
			diags = append(diags, diag.Errorf("error while updating cluster name: %s", err)...)
		}
	}

	// Upgrade call on client
	if d.HasChanges("plan_id", "is_ha", "storage", "major_version") {
		req := bridgeapi.ClusterUpgradeRequest{}
		var newPlan string
		var newDisk, newVer int
		var newHA bool

		if d.HasChange("plan_id") {
			newPlan = d.Get("plan_id").(string)
			req.PlanID = &newPlan
		}
		if d.HasChange("is_ha") {
			newHA = d.Get("is_ha").(bool)
			req.HighAvailability = &newHA
		}
		if d.HasChange("storage") {
			newDisk = d.Get("storage").(int)
			req.StorageGB = &newDisk
		}
		if d.HasChange("major_version") {
			newVer = d.Get("major_version").(int)
			req.PGMajorVersion = &newVer
		}

		err := client.UpgradeCluster(clusterID, req)
		if err != nil {
			diags = append(diags, diag.Errorf("error while upgrading cluster: %s", err)...)
		}
	}

	readDiag := resourceClusterRead(ctx, d, meta)
	diags = append(diags, readDiag...)

	return diags
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bridgeapi.Client)
	diags := []diag.Diagnostic{}

	clusterID := d.Id()

	err := client.DeleteCluster(clusterID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
