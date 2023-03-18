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
	"net/url"

	"github.com/CrunchyData/terraform-provider-crunchybridge/internal/bridgeapi"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	idConfigName     = "application_id"
	secretConfigName = "application_secret"
	urlConfigName    = "bridgeapi_url"
	tokenConfigName  = "require_token_swap"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"crunchybridge_account":       dataSourceAccount(),
				"crunchybridge_cloudprovider": dataSourceCloudProvider(),
				"crunchybridge_cluster":       dataSourceCluster(),
				"crunchybridge_clusterids":    dataSourceClusterIDs(),
				"crunchybridge_clusterroles":  dataSourceRoles(),
				"crunchybridge_clusterstatus": dataSourceStatus(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"crunchybridge_cluster": resourceCluster(),
			},
			Schema: map[string]*schema.Schema{
				idConfigName: {
					Type:        schema.TypeString,
					Description: "The application id component of the Crunchy Bridge API key.",
					DefaultFunc: schema.EnvDefaultFunc("APPLICATION_ID", ""),
					Required:    true,
				},
				secretConfigName: {
					Type:        schema.TypeString,
					Description: "The application secret component of the Crunchy Bridge API key.",
					DefaultFunc: schema.EnvDefaultFunc("APPLICATION_SECRET", ""),
					Required:    true,
				},
				tokenConfigName: {
					Type:        schema.TypeBool,
					Description: "When true, forces an exchange of the API key for a short-lived bearer token.",
					Optional:    true,
				},
				urlConfigName: {
					Type:        schema.TypeString,
					Description: "The API URL for the Crunchy Bridge platform API. Most users should not need to change this value.",
					DefaultFunc: schema.EnvDefaultFunc("BRIDGE_API_URL", "https://api.crunchybridge.com"),
					Required:    true,
				},
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		// Provider.UserAgent provides a UserAgent string with the
		// passed parameters, Terraform version, SDK version, and other
		// bits:
		userAgent := p.UserAgent("terraform-provider-crunchybridge", version)

		id := d.Get(idConfigName).(string)
		secret := d.Get(secretConfigName).(string)
		if (id == "") || (secret == "") {
			return nil, diag.Errorf("%s and %s must be configured to non-empty strings for this provider", idConfigName, secretConfigName)
		}
		login := bridgeapi.Login{
			Key:    id,
			Secret: secret,
		}

		apiUrl, err := url.Parse(d.Get(urlConfigName).(string))
		if err != nil {
			return nil, diag.FromErr(err)
		}

		options := []bridgeapi.ClientOption{
			bridgeapi.WithContext(ctx),
			bridgeapi.WithUserAgent(userAgent),
		}

		swapReq := d.Get(tokenConfigName).(bool)
		if swapReq {
			options = append(options, bridgeapi.WithTokenExchange(), bridgeapi.WithImmediateLogin())
		}

		c, err := bridgeapi.NewClient(apiUrl, login, options...)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return c, nil
	}
}
