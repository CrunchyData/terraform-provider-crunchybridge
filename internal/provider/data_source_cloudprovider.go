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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceCloudProvider() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source for retreiving Cluster resource data",
		ReadContext: dataSourceCloudProviderRead,
		Schema: map[string]*schema.Schema{
			// "Request" Fields
			"provider_id": {
				Description:  "The [cloud provider](https://docs.crunchybridge.com/api/provider) hosting clusters. Allows `aws`, `gcp`, or `azure`",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"aws", "gcp", "azure"}, false),
			},
			// "Result / Computed Fields"
			"plans": {
				Computed:    true,
				Description: "Listing of available plans for this provider.",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"plan_id": {
							Computed:    true,
							Description: "Primary ID for the plan.",
							Type:        schema.TypeString,
						},
						"plan_cpu": {
							Computed:    true,
							Description: "The number of CPU cores on the plan’s instance.",
							Type:        schema.TypeInt,
						},
						"plan_memory": {
							Computed:    true,
							Description: "The amount of memory on the plan’s instance in gigabytes.",
							Type:        schema.TypeInt,
						},
						"plan_name": {
							Computed:    true,
							Description: "The plan’s public display name.",
							Type:        schema.TypeString,
						},
					},
				},
			},
			"regions": {
				Computed:    true,
				Description: "Listing of valid regions for this provider.",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_id": {
							Computed:    true,
							Description: "Primary ID for the region.",
							Type:        schema.TypeString,
						},
						"region_location": {
							Computed:    true,
							Description: "Location is the name of the city or state where the region is hosted.",
							Type:        schema.TypeString,
						},
						"region_name": {
							Computed:    true,
							Description: "The region’s public name.",
							Type:        schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func dataSourceCloudProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bridgeapi.Client)

	id := d.Get("provider_id").(string)
	d.SetId("cloudprovider_" + id)
	diags := []diag.Diagnostic{}

	providers, err := client.Providers()
	if err != nil {
		return diag.FromErr(err)
	}

	var found bridgeapi.Provider
	for _, p := range providers {
		if p.ID == id {
			found = p
		}
	}
	if found.ID == "" {
		return diag.Errorf("Unable to find provider information for provider [%s]", id)
	}

	plans := []map[string]interface{}{}
	regions := []map[string]string{}

	for _, plan := range found.Plans {
		values := map[string]interface{}{
			"plan_id":     plan.ID,
			"plan_cpu":    plan.CPU,
			"plan_memory": plan.Memory,
			"plan_name":   plan.Name,
		}
		plans = append(plans, values)
	}
	d.Set("plans", plans)

	for _, region := range found.Regions {
		values := map[string]string{
			"region_id":       region.ID,
			"region_location": region.Location,
			"region_name":     region.Name,
		}
		regions = append(regions, values)
	}
	d.Set("regions", regions)

	return diag.Diagnostics(diags)
}
