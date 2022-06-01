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

func dataSourceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source for retreiving Account team resource data",
		ReadContext: dataSourceAccountRead,
		Schema: map[string]*schema.Schema{
			// "Result / Computed Fields"
			"id": {
				Computed:    true,
				Description: "The account's ID in [EID Format](https://docs.crunchybridge.com/api-concepts/eid)",
				Type:        schema.TypeString,
			},
			"default_team": {
				Computed:    true,
				Description: "The ID of the account's default team in [EID Format](https://docs.crunchybridge.com/api-concepts/eid)",
				Type:        schema.TypeString,
			},
			"personal_team": {
				Computed:    true,
				Description: "The ID of the account's personal team in [EID Format](https://docs.crunchybridge.com/api-concepts/eid)",
				Type:        schema.TypeString,
			},
			"team_membership": {
				Computed:    true,
				Description: "A listing of the account's team membership and role associated with those teams.",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"team_id": {
							Computed:    true,
							Description: "",
							Type:        schema.TypeString,
						},
						"team_name": {
							Computed:    true,
							Description: "",
							Type:        schema.TypeString,
						},
						"team_role": {
							Computed:    true,
							Description: "",
							Type:        schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func dataSourceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*bridgeapi.Client)

	acct, err := client.Account()
	if err != nil {
		return diag.FromErr(err)
	}

	diags := []diag.Diagnostic{}

	err = d.Set("id", acct.ID)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	} else {
		d.SetId(acct.ID)
	}

	// Null default_team_id means personal team is default which matches the user id.
	// Make the substitution here so DefaultTeamID is presented to the user as expected
	if acct.DefaultTeamID == "" {
		acct.DefaultTeamID = acct.ID
	}
	err = d.Set("default_team", acct.DefaultTeamID)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	err = d.Set("personal_team", acct.ID)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	teams, err := client.AccountTeams()
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// REVIEW: Given personal team, should always be true, remove?
	if len(teams) > 0 {
		teamInfo := []interface{}{}
		for _, team := range teams {
			teamItem := map[string]interface{}{
				"team_id":   team.ID,
				"team_name": team.Name,
				"team_role": team.Role,
			}
			teamInfo = append(teamInfo, teamItem)
		}
		err = d.Set("team_membership", teamInfo)
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	return diag.Diagnostics(diags)
}
