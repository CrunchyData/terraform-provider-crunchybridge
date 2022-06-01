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

func dataSourceRoles() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source for retreiving Cluster resource data",
		ReadContext: dataSourceRolesRead,
		Schema: map[string]*schema.Schema{
			// "Request" Fields
			"id": {
				Description: "The unique ID of the cluster in [EID format](https://docs.crunchybridge.com/api-concepts/eid).",
				Type:        schema.TypeString,
				Required:    true,
			},
			// "Result / Computed Fields"
			"superuser": {
				Computed:    true,
				Description: "Superuser role provided for the cluster.",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"application": {
				Computed:    true,
				Description: "Application role provided for the cluster.",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"user_roles": {
				Computed:    true,
				Description: "User roles provided for the cluster. Not currently used except to document the other role formats.",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Computed:    true,
							Description: "Name of the role in Postgres. Either `application` or `postgres`, enumeration of `u_` roles is not available yet.",
							Type:        schema.TypeString,
						},
						"password": {
							Computed:    true,
							Description: "Password of the role in Postgres.",
							Type:        schema.TypeString,
						},
						"team_id": {
							Computed:    true,
							Description: "The ID of the associated cluster's parent team in [EID format](https://docs.crunchybridge.com/api-concepts/eid).",
							Type:        schema.TypeString,
						},
						"uri": {
							Computed:    true,
							Description: "A full URI usable as a Postgres connection string for the named role.",
							Type:        schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func dataSourceRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bridgeapi.Client)

	id := d.Get("id").(string)
	d.SetId(id)

	roleList, err := client.ClusterRoles(id)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := []diag.Diagnostic{}
	userRoles := []map[string]string{}

	for _, roleItem := range roleList {
		roleMap := map[string]string{
			"name":       roleItem.Name,
			"password":   roleItem.Password,
			"cluster_id": roleItem.ClusterID,
			"uri":        roleItem.URI,
		}

		if roleItem.Name == "postgres" {
			err = d.Set("superuser", roleMap)
			if err != nil {
				diags = append(diags, diag.FromErr(err)...)
			}
		} else if roleItem.Name == "application" {
			err = d.Set("application", roleMap)
			if err != nil {
				diags = append(diags, diag.FromErr(err)...)
			}
		} else {
			userRoles = append(userRoles, roleMap)
		}
	}
	if len(userRoles) > 0 {
		err = d.Set("user_roles", userRoles)
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	return diag.Diagnostics(diags)
}
